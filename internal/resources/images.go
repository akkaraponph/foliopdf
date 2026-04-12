package resources

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
)

// ImageEntry holds parsed image data ready for PDF embedding.
type ImageEntry struct {
	Key        string // SHA-1 hash of data for deduplication
	Name       string // resource name label: "1", "2", ... → /Im1, /Im2
	Data       []byte // image bytes (JPEG: raw file; PNG: zlib-compressed pixels)
	Width      int    // pixels
	Height     int    // pixels
	ColorSpace string // "DeviceRGB", "DeviceGray", "DeviceCMYK"
	BPC        int    // bits per component (usually 8)
	Filter     string // "DCTDecode" for JPEG, "FlateDecode" for PNG
	ObjNum     int    // set during serialization

	// SMask holds zlib-compressed alpha-channel data for images with
	// transparency (e.g. RGBA PNGs). When non-nil, putImages writes a
	// separate SMask XObject and references it from the main image.
	SMaskData   []byte
	SMaskObjNum int
}

// ImageRegistry manages image registrations with deduplication.
type ImageRegistry struct {
	images  map[string]*ImageEntry // SHA-1 key → entry
	byName  map[string]*ImageEntry // user name → entry
	order   []string               // insertion order by SHA-1 key
	counter int
}

// NewImageRegistry creates an empty image registry.
func NewImageRegistry() *ImageRegistry {
	return &ImageRegistry{
		images: make(map[string]*ImageEntry),
		byName: make(map[string]*ImageEntry),
	}
}

// RegisterJPEG registers a JPEG image. If identical image data was already
// registered, the existing entry is reused (deduplication via SHA-1).
func (r *ImageRegistry) RegisterJPEG(name string, data io.Reader) (*ImageEntry, error) {
	raw, err := io.ReadAll(data)
	if err != nil {
		return nil, fmt.Errorf("read image %q: %w", name, err)
	}

	// SHA-1 hash for deduplication
	h := sha1.Sum(raw)
	key := fmt.Sprintf("%x", h)

	// Check for duplicate by content
	if existing, ok := r.images[key]; ok {
		// Map the new name to the existing entry too
		r.byName[name] = existing
		return existing, nil
	}

	// Probe JPEG dimensions and color space
	cfg, err := jpeg.DecodeConfig(bytesReader(raw))
	if err != nil {
		return nil, fmt.Errorf("decode JPEG %q: %w", name, err)
	}

	colorSpace := "DeviceRGB"
	// jpeg.DecodeConfig returns a color model; for grayscale JPEGs
	// the model will be image.GrayModel
	if cfg.ColorModel != nil {
		// Check if it's a gray model by probing
		r0, g0, b0, _ := cfg.ColorModel.Convert(cfg.ColorModel.Convert(gray128{})).RGBA()
		if r0 == g0 && g0 == b0 {
			// Might be gray, but JPEG in Go always decodes as YCbCr
			// For simplicity, default to DeviceRGB for MVP
		}
	}

	r.counter++
	entry := &ImageEntry{
		Key:        key,
		Name:       fmt.Sprintf("%d", r.counter),
		Data:       raw,
		Width:      cfg.Width,
		Height:     cfg.Height,
		ColorSpace: colorSpace,
		BPC:        8,
		Filter:     "DCTDecode",
	}

	r.images[key] = entry
	r.byName[name] = entry
	r.order = append(r.order, key)
	return entry, nil
}

// Get retrieves a registered image by its user-provided name.
func (r *ImageRegistry) Get(name string) (*ImageEntry, bool) {
	e, ok := r.byName[name]
	return e, ok
}

// All returns all unique images in registration order.
func (r *ImageRegistry) All() []*ImageEntry {
	result := make([]*ImageEntry, len(r.order))
	for i, key := range r.order {
		result[i] = r.images[key]
	}
	return result
}

// RegisterPNG registers a PNG image. The pixel data is extracted and
// zlib-compressed for FlateDecode embedding. If the PNG has an alpha
// channel, a separate SMask stream is created.
func (r *ImageRegistry) RegisterPNG(name string, data io.Reader) (*ImageEntry, error) {
	raw, err := io.ReadAll(data)
	if err != nil {
		return nil, fmt.Errorf("read image %q: %w", name, err)
	}

	// Deduplication by content hash
	h := sha1.Sum(raw)
	key := fmt.Sprintf("%x", h)
	if existing, ok := r.images[key]; ok {
		r.byName[name] = existing
		return existing, nil
	}

	// Decode PNG
	img, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("decode PNG %q: %w", name, err)
	}

	colorData, alphaData, colorSpace := extractPixels(img)

	compressedColor, err := zlibCompress(colorData)
	if err != nil {
		return nil, fmt.Errorf("compress PNG %q: %w", name, err)
	}

	bounds := img.Bounds()
	r.counter++
	entry := &ImageEntry{
		Key:        key,
		Name:       fmt.Sprintf("%d", r.counter),
		Data:       compressedColor,
		Width:      bounds.Dx(),
		Height:     bounds.Dy(),
		ColorSpace: colorSpace,
		BPC:        8,
		Filter:     "FlateDecode",
	}

	if alphaData != nil {
		compressed, err := zlibCompress(alphaData)
		if err != nil {
			return nil, fmt.Errorf("compress PNG alpha %q: %w", name, err)
		}
		entry.SMaskData = compressed
	}

	r.images[key] = entry
	r.byName[name] = entry
	r.order = append(r.order, key)
	return entry, nil
}

// extractPixels converts any image.Image to raw pixel bytes and an
// optional alpha channel. Returns (colorData, alphaData, colorSpace).
// alphaData is nil when all pixels are fully opaque.
func extractPixels(img image.Image) (colorData, alphaData []byte, colorSpace string) {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	// Detect grayscale images.
	isGray := false
	switch img.(type) {
	case *image.Gray:
		isGray = true
	}

	if isGray {
		colorSpace = "DeviceGray"
		colorData = make([]byte, w*h)
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				px := (y-bounds.Min.Y)*w + (x - bounds.Min.X)
				g := color.GrayModel.Convert(img.At(x, y)).(color.Gray)
				colorData[px] = g.Y
			}
		}
		return
	}

	// RGB/RGBA — extract via NRGBA to handle premultiplied alpha correctly.
	colorSpace = "DeviceRGB"
	colorData = make([]byte, w*h*3)
	alphaData = make([]byte, w*h)
	hasAlpha := false

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			px := (y-bounds.Min.Y)*w + (x - bounds.Min.X)
			nrgba := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
			colorData[px*3] = nrgba.R
			colorData[px*3+1] = nrgba.G
			colorData[px*3+2] = nrgba.B
			alphaData[px] = nrgba.A
			if nrgba.A != 0xFF {
				hasAlpha = true
			}
		}
	}

	if !hasAlpha {
		alphaData = nil
	}
	return
}

// zlibCompress compresses data with zlib default compression.
func zlibCompress(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw, err := zlib.NewWriterLevel(&buf, zlib.DefaultCompression)
	if err != nil {
		return nil, err
	}
	if _, err := zw.Write(data); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// bytesReader creates an io.Reader from a byte slice.
func bytesReader(b []byte) io.Reader {
	return &byteSliceReader{data: b, pos: 0}
}

type byteSliceReader struct {
	data []byte
	pos  int
}

func (r *byteSliceReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

// gray128 is a dummy color used to probe color models.
type gray128 struct{}

func (gray128) RGBA() (uint32, uint32, uint32, uint32) {
	return 0x8080, 0x8080, 0x8080, 0xFFFF
}

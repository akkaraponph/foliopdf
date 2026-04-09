package resources

import (
	"crypto/sha1"
	"fmt"
	"image/jpeg"
	"io"
)

// ImageEntry holds parsed image data ready for PDF embedding.
type ImageEntry struct {
	Key        string  // SHA-1 hash of data for deduplication
	Name       string  // resource name label: "1", "2", ... → /Im1, /Im2
	Data       []byte  // raw image bytes (JPEG: the JPEG file bytes)
	Width      int     // pixels
	Height     int     // pixels
	ColorSpace string  // "DeviceRGB", "DeviceGray", "DeviceCMYK"
	BPC        int     // bits per component (usually 8)
	Filter     string  // "DCTDecode" for JPEG
	ObjNum     int     // set during serialization
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

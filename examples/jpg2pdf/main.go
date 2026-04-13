package main

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"log"
	"math"
	"os"

	"github.com/akkaraponph/foliopdf"
)

func main() {
	// Generate sample images.
	samples := createSampleImages()
	fmt.Printf("Generated %d sample images\n\n", len(samples))

	// 1. Auto-fit: each page sized to its image.
	err := foliopdf.ImagesToPDF("output/autofit.pdf", samples)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("output/autofit.pdf        — pages sized to each image")

	// 2. Fixed A4 pages, images scaled to fit with margins.
	err = foliopdf.ImagesToPDF("output/a4_fit.pdf", samples,
		foliopdf.ImagePageSize(foliopdf.A4),
		foliopdf.ImageMargin(36),
		foliopdf.ImageFit("fit"),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("output/a4_fit.pdf         — A4 pages, fit mode")

	// 3. Fixed A4, stretch to fill.
	err = foliopdf.ImagesToPDF("output/a4_stretch.pdf", samples,
		foliopdf.ImagePageSize(foliopdf.A4),
		foliopdf.ImageMargin(36),
		foliopdf.ImageFit("stretch"),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("output/a4_stretch.pdf     — A4 pages, stretch mode")

	// 4. High DPI (smaller auto-fit pages).
	err = foliopdf.ImagesToPDF("output/high_dpi.pdf", samples,
		foliopdf.ImageDPI(300),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("output/high_dpi.pdf       — auto-fit at 300 DPI")

	// 5. Landscape A4 with fill mode.
	err = foliopdf.ImagesToPDF("output/landscape.pdf", samples,
		foliopdf.ImagePageSize(foliopdf.A4.Landscape()),
		foliopdf.ImageMargin(24),
		foliopdf.ImageFit("fill"),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("output/landscape.pdf      — A4 landscape, fill mode")
}

// createSampleImages generates 3 sample images to demonstrate the feature.
func createSampleImages() []string {
	os.MkdirAll("samples", 0o755)

	paths := []string{
		createGradientJPEG("samples/photo1.jpg", 800, 600),
		createCheckerPNG("samples/photo2.png", 400, 400),
		createGradientJPEG("samples/photo3.jpg", 600, 900),
	}
	return paths
}

// createGradientJPEG creates a colorful gradient image.
func createGradientJPEG(path string, w, h int) string {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r := uint8(float64(x) / float64(w) * 255)
			g := uint8(float64(y) / float64(h) * 255)
			b := uint8(128 + 127*math.Sin(float64(x+y)*0.02))
			img.SetRGBA(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
	f, _ := os.Create(path)
	jpeg.Encode(f, img, &jpeg.Options{Quality: 90})
	f.Close()
	return path
}

// createCheckerPNG creates a checkerboard PNG with transparency.
func createCheckerPNG(path string, w, h int) string {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	sq := 40
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dark := ((x/sq)+(y/sq))%2 == 0
			if dark {
				img.SetRGBA(x, y, color.RGBA{R: 40, G: 80, B: 160, A: 230})
			} else {
				img.SetRGBA(x, y, color.RGBA{R: 220, G: 230, B: 245, A: 255})
			}
		}
	}
	f, _ := os.Create(path)
	png.Encode(f, img)
	f.Close()
	return path
}

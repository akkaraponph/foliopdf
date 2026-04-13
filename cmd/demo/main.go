package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"os"

	"github.com/akkaraponph/presspdf"
)

func main() {
	doc := presspdf.New(presspdf.WithCompression(false))
	doc.SetTitle("Folio Demo")
	doc.SetAuthor("Bob")
	doc.SetFont("helvetica", "", 16)

	page := doc.AddPage(presspdf.A4)
	page.TextAt(40, 30, "Hello Folio!")

	doc.SetFont("helvetica", "B", 12)
	page.TextAt(40, 45, "This is bold text")

	doc.SetFont("times", "", 12)
	page.TextAt(40, 60, "This is Times Roman")

	// Draw some shapes
	page.Line(40, 70, 200, 70)
	page.Rect(40, 80, 160, 60, "D")

	doc.SetFillColor(200, 220, 255)
	page.Rect(50, 90, 140, 40, "DF")

	// MultiCell demo
	doc.SetFont("helvetica", "", 10)
	doc.SetFillColor(240, 240, 240)
	page.SetXY(40, 150)
	page.MultiCell(130, 5, "This is a MultiCell block with automatic word wrapping. "+
		"The text flows within the specified width and wraps at word boundaries. "+
		"Each line becomes its own cell.", "1", "L", true)

	// Image demo
	jpgData := makeTestImage(120, 80)
	doc.RegisterImage("gradient", bytes.NewReader(jpgData))
	page.TextAt(40, 200, "JPEG Image:")
	page.DrawImageRect("gradient", 40, 205, 60, 40)

	// Second page
	doc.SetFont("courier", "", 14)
	p2 := doc.AddPage(presspdf.Letter)
	p2.TextAt(30, 30, "Page 2 - Letter size")
	p2.Line(30, 40, 250, 40)

	// Cell demo
	doc.SetFont("helvetica", "", 10)
	p2.SetXY(30, 50)
	p2.Cell(60, 8, "Left aligned", "1", "L", false, 0)
	p2.Cell(60, 8, "Centered", "1", "C", false, 0)
	p2.Cell(60, 8, "Right aligned", "1", "R", false, 1)

	path := "/tmp/presspdf_demo.pdf"
	if err := doc.Save(path); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("PDF saved to %s (%d pages)\n", path, 2)
}

func makeTestImage(w, h int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, color.RGBA{
				R: uint8(x * 255 / w),
				G: uint8(y * 255 / h),
				B: 128,
				A: 255,
			})
		}
	}
	var buf bytes.Buffer
	jpeg.Encode(&buf, img, &jpeg.Options{Quality: 75})
	return buf.Bytes()
}

package presspdf

import (
	"fmt"

	"github.com/akkaraponph/presspdf/internal/barcode"
	"github.com/akkaraponph/presspdf/internal/state"
)

// QR code error correction levels.
const (
	ECLow      = 0 // ~7% recovery
	ECMedium   = 1 // ~15% recovery
	ECQuartile = 2 // ~25% recovery
	ECHigh     = 3 // ~30% recovery
)

// Barcode128 draws a Code 128 barcode at position (x, y) with the given
// width and height in user units. The data string is encoded automatically
// using Code B (ASCII) or Code C (all-numeric).
func (p *Page) Barcode128(x, y, w, h float64, data string) {
	p = p.active()
	if p.doc.err != nil {
		return
	}

	bars := barcode.Code128(data)
	if len(bars) == 0 {
		p.doc.err = fmt.Errorf("Barcode128: empty data")
		return
	}

	// Compute total modules.
	totalModules := 0
	for _, bw := range bars {
		totalModules += bw
	}

	k := p.doc.k
	moduleW := w * k / float64(totalModules)
	barH := h * k
	xPt := state.ToPointsX(x, k)
	yPt := state.ToPointsY(y+h, p.h, k)

	// Draw bars (alternating black/white starting with black).
	curX := xPt
	for i, bw := range bars {
		barW := float64(bw) * moduleW
		if i%2 == 0 { // black bar
			p.stream.Rect(curX, yPt, barW, barH)
			p.stream.Fill()
		}
		curX += barW
	}
}

// BarcodeEAN13 draws an EAN-13 barcode at position (x, y) with the given
// width and height in user units. digits must be 12 or 13 numeric characters.
// If 12 digits, the check digit is computed automatically.
func (p *Page) BarcodeEAN13(x, y, w, h float64, digits string) {
	p = p.active()
	if p.doc.err != nil {
		return
	}

	modules, err := barcode.EAN13(digits)
	if err != nil {
		p.doc.err = fmt.Errorf("BarcodeEAN13: %w", err)
		return
	}

	k := p.doc.k
	moduleW := w * k / float64(len(modules))
	barH := h * k
	xPt := state.ToPointsX(x, k)
	yPt := state.ToPointsY(y+h, p.h, k)

	// Draw each module.
	for i, dark := range modules {
		if dark {
			mx := xPt + float64(i)*moduleW
			p.stream.Rect(mx, yPt, moduleW, barH)
			p.stream.Fill()
		}
	}
}

// Barcode128WithText draws a Code 128 barcode with human-readable text
// centered below it. textSize is the font size in points for the label.
func (p *Page) Barcode128WithText(x, y, w, h float64, data string, textSize float64) {
	p.Barcode128(x, y, w, h, data)
	if p.doc.err != nil {
		return
	}
	p.drawBarcodeText(x, y+h, w, data, textSize)
}

// BarcodeEAN13WithText draws an EAN-13 barcode with human-readable text
// centered below it. textSize is the font size in points for the label.
func (p *Page) BarcodeEAN13WithText(x, y, w, h float64, digits string, textSize float64) {
	p.BarcodeEAN13(x, y, w, h, digits)
	if p.doc.err != nil {
		return
	}
	p.drawBarcodeText(x, y+h, w, digits, textSize)
}

// drawBarcodeText draws centered text below a barcode.
func (p *Page) drawBarcodeText(x, y, w float64, text string, textSize float64) {
	// Save font state.
	savedFamily := p.doc.fontFamily
	savedStyle := p.doc.fontStyle
	savedSize := p.doc.fontSizePt

	// Use current font at specified size.
	family := savedFamily
	if family == "" {
		family = "helvetica"
	}
	p.doc.SetFont(family, "", textSize)

	// Center the text.
	textW := p.GetStringWidth(text)
	tx := x + (w-textW)/2
	ty := y + 1 // 1mm gap below barcode
	p.TextAt(tx, ty, text)

	// Restore font state.
	p.doc.SetFont(savedFamily, savedStyle, savedSize)
}

// QRCode draws a QR code at position (x, y) with the given size (width
// and height are equal) in user units. ecLevel: 0=Low, 1=Medium, 2=Quartile,
// 3=High. The data is encoded in byte mode, supporting versions 1-10.
func (p *Page) QRCode(x, y, size float64, data string, ecLevel int) {
	p = p.active()
	if p.doc.err != nil {
		return
	}

	matrix, err := barcode.QRCode(data, ecLevel)
	if err != nil {
		p.doc.err = fmt.Errorf("QRCode: %w", err)
		return
	}

	k := p.doc.k
	modules := len(matrix)
	moduleSize := size * k / float64(modules)
	xPt := state.ToPointsX(x, k)
	yPt := state.ToPointsY(y+size, p.h, k) // bottom-left corner

	// Draw black modules.
	for r := 0; r < modules; r++ {
		for c := 0; c < modules; c++ {
			if matrix[r][c] {
				mx := xPt + float64(c)*moduleSize
				my := yPt + float64(modules-1-r)*moduleSize // flip Y
				p.stream.Rect(mx, my, moduleSize, moduleSize)
				p.stream.Fill()
			}
		}
	}
}

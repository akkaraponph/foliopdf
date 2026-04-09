package folio

import (
	"fmt"
	"strings"

	"github.com/akkaraponph/folio/internal/content"
	"github.com/akkaraponph/folio/internal/resources"
	"github.com/akkaraponph/folio/internal/state"
)

// Page represents a single PDF page with drawing methods.
type Page struct {
	doc    *Document
	stream content.Stream
	size   PageSize
	w, h   float64 // page dimensions in user units

	// cursor position in user units
	x, y float64

	// page-local font state
	fontFamily string
	fontStyle  string
	fontSizePt float64
	fontEntry  *resources.FontEntry
}

// --- Text methods ---

// TextAt draws text at the given position (in user units, top-left origin).
func (p *Page) TextAt(x, y float64, text string) {
	if p.doc.err != nil {
		return
	}
	fe := p.effectiveFontEntry()
	if fe == nil {
		p.doc.err = fmt.Errorf("TextAt: no font set")
		return
	}

	k := p.doc.k
	xPt := state.ToPointsX(x, k)
	// Position text baseline: y + 0.7 * fontSize (approximate ascender)
	fontSize := p.effectiveFontSizePt() / k
	yPt := state.ToPointsY(y+0.7*fontSize, p.h, k)

	// Handle text color
	tc := p.doc.textColor
	needColor := !tc.IsBlack()
	if needColor {
		p.stream.SaveState()
		p.stream.SetFillColorRGB(tc.R, tc.G, tc.B)
	}

	p.stream.BeginText()
	p.stream.SetFont("F"+fe.Index, p.effectiveFontSizePt())
	p.stream.MoveText(xPt, yPt)
	p.stream.ShowText(pdfEscape(text))
	p.stream.EndText()

	if needColor {
		p.stream.RestoreState()
	}
}

// SetFont sets the font for this page.
func (p *Page) SetFont(family, style string, size float64) {
	if p.doc.err != nil {
		return
	}
	fe, err := p.doc.fonts.Register(family, style)
	if err != nil {
		p.doc.err = fmt.Errorf("Page.SetFont: %w", err)
		return
	}
	p.fontFamily = family
	p.fontStyle = style
	p.fontSizePt = size
	p.fontEntry = fe
	p.applyFont(fe, size)
}

// GetStringWidth returns the width of s in user units using the current font.
func (p *Page) GetStringWidth(s string) float64 {
	fe := p.effectiveFontEntry()
	if fe == nil {
		return 0
	}
	w := resources.StringWidth(fe, s)
	return float64(w) * p.effectiveFontSizePt() / 1000.0 / p.doc.k
}

// --- Drawing methods ---

// Line draws a line segment from (x1,y1) to (x2,y2) in user units.
func (p *Page) Line(x1, y1, x2, y2 float64) {
	if p.doc.err != nil {
		return
	}
	k := p.doc.k
	p.stream.MoveTo(
		state.ToPointsX(x1, k),
		state.ToPointsY(y1, p.h, k),
	)
	p.stream.LineTo(
		state.ToPointsX(x2, k),
		state.ToPointsY(y2, p.h, k),
	)
	p.stream.Stroke()
}

// Rect draws a rectangle. style: "D" (draw/stroke), "F" (fill), "DF" or "FD" (both).
func (p *Page) Rect(x, y, w, h float64, style string) {
	if p.doc.err != nil {
		return
	}
	k := p.doc.k
	p.stream.Rect(
		state.ToPointsX(x, k),
		state.ToPointsY(y+h, p.h, k), // bottom-left corner in PDF coords
		w*k,
		h*k,
	)
	style = strings.ToUpper(style)
	switch style {
	case "F":
		p.stream.Fill()
	case "DF", "FD":
		p.stream.FillStroke()
	default: // "D" or ""
		p.stream.Stroke()
	}
}

// DrawImageRect draws a registered image at (x, y) with the given width and height.
func (p *Page) DrawImageRect(name string, x, y, w, h float64) {
	if p.doc.err != nil {
		return
	}
	entry, ok := p.doc.images.Get(name)
	if !ok {
		p.doc.err = fmt.Errorf("DrawImageRect: image %q not registered", name)
		return
	}

	k := p.doc.k
	wPt := w * k
	hPt := h * k
	xPt := state.ToPointsX(x, k)
	yPt := state.ToPointsY(y+h, p.h, k) // bottom-left of image in PDF coords

	p.stream.DrawImage("Im"+entry.Name, wPt, 0, 0, hPt, xPt, yPt)
}

// --- Cell and MultiCell ---

// Cell draws a single-line cell at the current cursor position.
// w: cell width (0 = extend to right margin)
// h: cell height
// text: cell text
// border: "" (none), "1" (full), or combination of "L","T","R","B"
// align: "L" (left, default), "C" (center), "R" (right)
// fill: if true, fill background with current fill color
// ln: 0 = cursor right, 1 = next line, 2 = below
func (p *Page) Cell(w, h float64, text, border, align string, fill bool, ln int) {
	if p.doc.err != nil {
		return
	}
	d := p.doc
	k := d.k

	if w == 0 {
		w = p.w - d.rMargin - p.x
	}

	// Draw fill/border
	if fill || border == "1" {
		if fill {
			op := "re f"
			if border == "1" {
				op = "re B"
			}
			_ = op
			p.stream.Rect(
				state.ToPointsX(p.x, k),
				state.ToPointsY(p.y+h, p.h, k),
				w*k, h*k,
			)
			if border == "1" {
				p.stream.FillStroke()
			} else {
				p.stream.Fill()
			}
		} else if border == "1" {
			p.stream.Rect(
				state.ToPointsX(p.x, k),
				state.ToPointsY(p.y+h, p.h, k),
				w*k, h*k,
			)
			p.stream.Stroke()
		}
	}

	// Individual borders
	if strings.Contains(border, "L") {
		p.drawBorderLine(p.x, p.y, p.x, p.y+h)
	}
	if strings.Contains(border, "T") {
		p.drawBorderLine(p.x, p.y, p.x+w, p.y)
	}
	if strings.Contains(border, "R") {
		p.drawBorderLine(p.x+w, p.y, p.x+w, p.y+h)
	}
	if strings.Contains(border, "B") {
		p.drawBorderLine(p.x, p.y+h, p.x+w, p.y+h)
	}

	// Draw text
	if text != "" {
		fe := p.effectiveFontEntry()
		if fe == nil {
			d.err = fmt.Errorf("Cell: no font set")
			return
		}
		fontSize := p.effectiveFontSizePt()
		fontSizeUser := fontSize / k

		// Text X position based on alignment
		var dx float64
		sw := p.GetStringWidth(text)
		switch strings.ToUpper(align) {
		case "C":
			dx = (w - sw) / 2
		case "R":
			dx = w - d.cMargin - sw
		default: // "L"
			dx = d.cMargin
		}

		// Text Y: vertically center in cell
		textX := state.ToPointsX(p.x+dx, k)
		textY := state.ToPointsY(p.y+0.5*h+0.3*fontSizeUser, p.h, k)

		tc := d.textColor
		needColor := !tc.IsBlack()
		if needColor {
			p.stream.SaveState()
			p.stream.SetFillColorRGB(tc.R, tc.G, tc.B)
		}

		p.stream.BeginText()
		p.stream.SetFont("F"+fe.Index, fontSize)
		p.stream.MoveText(textX, textY)
		p.stream.ShowText(pdfEscape(text))
		p.stream.EndText()

		if needColor {
			p.stream.RestoreState()
		}
	}

	// Advance cursor
	switch ln {
	case 0:
		p.x += w
	case 1:
		p.x = d.lMargin
		p.y += h
	case 2:
		p.y += h
	}
}

// MultiCell draws multi-line text with automatic word wrapping.
// w: cell width (0 = extend to right margin)
// h: line height
// text: text content (may contain \n)
// border: "" (none), "1" (full), or combination of "L","T","R","B"
// align: "L", "C", "R", "J" (justified)
// fill: if true, fill background with current fill color
func (p *Page) MultiCell(w, h float64, text, border, align string, fill bool) {
	if p.doc.err != nil {
		return
	}
	d := p.doc

	if w == 0 {
		w = p.w - d.rMargin - p.x
	}

	fe := p.effectiveFontEntry()
	if fe == nil {
		d.err = fmt.Errorf("MultiCell: no font set")
		return
	}
	fontSize := p.effectiveFontSizePt()

	// Available width in 1/1000 font units
	wmax := (w - 2*d.cMargin) * 1000.0 / fontSize * d.k

	// Split text into lines, then wrap each line
	lines := p.wrapText(text, fe, wmax)

	for i, line := range lines {
		// Determine borders for this line
		b := ""
		if border == "1" {
			if i == 0 {
				b = "LTR"
			} else if i == len(lines)-1 {
				b = "LRB"
			} else {
				b = "LR"
			}
		} else {
			if i == 0 && strings.Contains(border, "T") {
				b += "T"
			}
			if strings.Contains(border, "L") {
				b += "L"
			}
			if strings.Contains(border, "R") {
				b += "R"
			}
			if i == len(lines)-1 && strings.Contains(border, "B") {
				b += "B"
			}
		}

		p.Cell(w, h, line, b, align, fill, 2)
	}

	// Move to left margin
	p.x = d.lMargin
}

// wrapText splits text by newlines, then wraps each line to fit within wmax
// (measured in 1/1000 font units).
func (p *Page) wrapText(text string, fe *resources.FontEntry, wmax float64) []string {
	var lines []string
	paragraphs := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")

	for _, para := range paragraphs {
		if para == "" {
			lines = append(lines, "")
			continue
		}

		line := ""
		lineWidth := 0

		words := strings.Fields(para)
		for i, word := range words {
			wordWidth := resources.StringWidth(fe, word)
			spaceWidth := 0
			if line != "" {
				spaceWidth = fe.Widths[' ']
			}

			if lineWidth+spaceWidth+wordWidth > int(wmax) && line != "" {
				lines = append(lines, line)
				line = word
				lineWidth = wordWidth
			} else {
				if i > 0 {
					line += " "
					lineWidth += spaceWidth
				}
				line += word
				lineWidth += wordWidth
			}
		}
		if line != "" {
			lines = append(lines, line)
		}
	}

	if len(lines) == 0 {
		lines = []string{""}
	}
	return lines
}

// drawBorderLine draws a single border line segment.
func (p *Page) drawBorderLine(x1, y1, x2, y2 float64) {
	k := p.doc.k
	p.stream.MoveTo(state.ToPointsX(x1, k), state.ToPointsY(y1, p.h, k))
	p.stream.LineTo(state.ToPointsX(x2, k), state.ToPointsY(y2, p.h, k))
	p.stream.Stroke()
}

// --- Cursor methods ---

// SetX sets the X cursor position.
func (p *Page) SetX(x float64) { p.x = x }

// SetY sets the Y cursor position.
func (p *Page) SetY(y float64) { p.y = y }

// SetXY sets both cursor positions.
func (p *Page) SetXY(x, y float64) { p.x = x; p.y = y }

// GetX returns the current X cursor position.
func (p *Page) GetX() float64 { return p.x }

// GetY returns the current Y cursor position.
func (p *Page) GetY() float64 { return p.y }

// --- Internal ---

// applyFont emits the font change to the content stream.
func (p *Page) applyFont(fe *resources.FontEntry, sizePt float64) {
	p.stream.BeginText()
	p.stream.SetFont("F"+fe.Index, sizePt)
	p.stream.EndText()
}

// effectiveFontEntry returns the page-level font, falling back to document.
func (p *Page) effectiveFontEntry() *resources.FontEntry {
	if p.fontEntry != nil {
		return p.fontEntry
	}
	return p.doc.fontEntry
}

// effectiveFontSizePt returns the page-level font size, falling back to document.
func (p *Page) effectiveFontSizePt() float64 {
	if p.fontSizePt > 0 {
		return p.fontSizePt
	}
	return p.doc.fontSizePt
}

package folio

import (
	"fmt"
	"strings"

	"github.com/akkaraponph/folio/internal/content"
	"github.com/akkaraponph/folio/internal/resources"
	"github.com/akkaraponph/folio/internal/state"
)

// templateEntry stores a registered page template (PDF Form XObject).
type templateEntry struct {
	name   string         // resource name: "Tpl1", "Tpl2", ...
	size   PageSize       // template dimensions
	stream content.Stream // content stream
	objNum int            // set during serialization
}

// Template provides drawing methods for building a reusable page template.
// Drawing on a Template works like drawing on a Page, but the result is
// stored as a PDF Form XObject that can be stamped onto any page.
type Template struct {
	doc    *Document
	entry  *templateEntry
	stream *content.Stream
	w, h   float64 // dimensions in user units

	// cursor
	x, y float64

	// page-local font state
	fontFamily string
	fontStyle  string
	fontSizePt float64
	fontEntry  *resources.FontEntry
}

// BeginTemplate starts building a new page template with the given
// dimensions. Draw on the returned Template, then call EndTemplate to
// finalize. The template can later be stamped onto any page with
// UseTemplate.
func (d *Document) BeginTemplate(size PageSize) *Template {
	entry := &templateEntry{
		name: fmt.Sprintf("Tpl%d", len(d.templates)+1),
		size: size,
	}
	d.templates = append(d.templates, entry)
	return &Template{
		doc:    d,
		entry:  entry,
		stream: &entry.stream,
		w:      size.WidthPt / d.k,
		h:      size.HeightPt / d.k,
		x:      d.lMargin,
		y:      d.tMargin,
	}
}

// EndTemplate finalizes the template and returns its name for use with
// Page.UseTemplate.
func (d *Document) EndTemplate() string {
	if len(d.templates) == 0 {
		return ""
	}
	return d.templates[len(d.templates)-1].name
}

// UseTemplate stamps a previously created template onto the page at
// position (x, y) with the given width and height in user units.
func (p *Page) UseTemplate(name string, x, y, w, h float64) {
	p = p.active()
	if p.doc.err != nil {
		return
	}
	k := p.doc.k

	// Find the template.
	var entry *templateEntry
	for _, t := range p.doc.templates {
		if t.name == name {
			entry = t
			break
		}
	}
	if entry == nil {
		p.doc.err = fmt.Errorf("UseTemplate: template %q not found", name)
		return
	}

	// Scale template to fit the requested size.
	sx := w * k / entry.size.WidthPt
	sy := h * k / entry.size.HeightPt
	tx := state.ToPointsX(x, k)
	ty := state.ToPointsY(y+h, p.h, k) // bottom-left in PDF coords

	p.stream.SaveState()
	p.stream.ConcatMatrix(sx, 0, 0, sy, tx, ty)
	p.stream.Raw(fmt.Sprintf("/%s Do", entry.name))
	p.stream.RestoreState()
}

// --- Template drawing methods ---
// These mirror the Page drawing methods but operate on the template's stream.

// Rect draws a rectangle on the template.
func (t *Template) Rect(x, y, w, h float64, style string) {
	k := t.doc.k
	t.stream.Rect(
		state.ToPointsX(x, k),
		state.ToPointsY(y+h, t.h, k),
		w*k,
		h*k,
	)
	style = strings.ToUpper(style)
	switch style {
	case "F":
		t.stream.Fill()
	case "DF", "FD":
		t.stream.FillStroke()
	default:
		t.stream.Stroke()
	}
}

// Line draws a line on the template.
func (t *Template) Line(x1, y1, x2, y2 float64) {
	k := t.doc.k
	t.stream.MoveTo(state.ToPointsX(x1, k), state.ToPointsY(y1, t.h, k))
	t.stream.LineTo(state.ToPointsX(x2, k), state.ToPointsY(y2, t.h, k))
	t.stream.Stroke()
}

// SetFillColorRGB sets the fill color on the template (0-255).
func (t *Template) SetFillColorRGB(r, g, b int) {
	c := state.ColorFromRGB(r, g, b)
	t.stream.SetFillColorRGB(c.R, c.G, c.B)
}

// SetDrawColorRGB sets the stroke color on the template (0-255).
func (t *Template) SetDrawColorRGB(r, g, b int) {
	c := state.ColorFromRGB(r, g, b)
	t.stream.SetStrokeColorRGB(c.R, c.G, c.B)
}

// SetLineWidth sets the line width on the template in user units.
func (t *Template) SetLineWidth(w float64) {
	t.stream.SetLineWidth(w * t.doc.k)
}

// TextAt draws text at the given position on the template.
func (t *Template) TextAt(x, y float64, text string) {
	fe := t.effectiveFontEntry()
	if fe == nil {
		t.doc.err = fmt.Errorf("Template.TextAt: no font set")
		return
	}
	k := t.doc.k
	fontSize := t.effectiveFontSizePt()
	fontSizeUser := fontSize / k
	xPt := state.ToPointsX(x, k)
	yPt := state.ToPointsY(y+0.7*fontSizeUser, t.h, k)

	t.stream.BeginText()
	t.stream.SetFont("F"+fe.Index, fontSize)
	t.stream.MoveText(xPt, yPt)
	if fe.Type == "TTF" {
		fe.AddUsedRunes(text)
		t.stream.ShowTextHex(textToHex(text))
	} else {
		t.stream.ShowText(pdfEscape(text))
	}
	t.stream.EndText()
}

// SetFont sets the font for this template.
func (t *Template) SetFont(family, style string, size float64) {
	fe, ok := t.doc.fonts.Get(family, style)
	if !ok {
		var err error
		fe, err = t.doc.fonts.Register(family, style)
		if err != nil {
			t.doc.err = fmt.Errorf("Template.SetFont: %w", err)
			return
		}
	}
	t.fontFamily = family
	t.fontStyle = style
	t.fontSizePt = size
	t.fontEntry = fe
}

// DrawImageRect draws a registered image on the template.
func (t *Template) DrawImageRect(name string, x, y, w, h float64) {
	entry, ok := t.doc.images.Get(name)
	if !ok {
		t.doc.err = fmt.Errorf("Template.DrawImageRect: image %q not registered", name)
		return
	}
	k := t.doc.k
	wPt := w * k
	hPt := h * k
	xPt := state.ToPointsX(x, k)
	yPt := state.ToPointsY(y+h, t.h, k)
	t.stream.DrawImage("Im"+entry.Name, wPt, 0, 0, hPt, xPt, yPt)
}

func (t *Template) effectiveFontEntry() *resources.FontEntry {
	if t.fontEntry != nil {
		return t.fontEntry
	}
	return t.doc.fontEntry
}

func (t *Template) effectiveFontSizePt() float64 {
	if t.fontSizePt > 0 {
		return t.fontSizePt
	}
	return t.doc.fontSizePt
}

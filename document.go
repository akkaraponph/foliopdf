package folio

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/akkaraponph/folio/internal/resources"
	"github.com/akkaraponph/folio/internal/state"
)

// Document is the root object for creating a PDF.
type Document struct {
	// metadata
	title    string
	author   string
	subject  string
	creator  string
	producer string

	// configuration
	unit     state.Unit
	k        float64 // scale factor: points per user unit
	compress bool
	defSize  PageSize
	lMargin  float64 // left margin in user units
	tMargin  float64 // top margin
	rMargin  float64 // right margin
	bMargin  float64 // bottom margin
	cMargin  float64 // cell margin

	// pages
	pages       []*Page
	currentPage *Page

	// resources
	fonts  *resources.FontRegistry
	images *resources.ImageRegistry

	// current font state (carried across pages)
	fontFamily string
	fontStyle  string
	fontSizePt float64
	fontEntry  *resources.FontEntry

	// current drawing state (carried across pages)
	drawColor state.Color
	fillColor state.Color
	textColor state.Color
	lineWidth float64

	// error accumulation
	err error
}

// New creates a new Document with the given options.
// Defaults: mm units, compression enabled, 10mm margins, A4 page size.
func New(opts ...Option) *Document {
	d := &Document{
		producer: "Folio",
		unit:     state.UnitMM,
		k:        state.ScaleFactor(state.UnitMM),
		compress: true,
		defSize:  A4,
		lMargin:  10,
		tMargin:  10,
		rMargin:  10,
		bMargin:  10,
		cMargin:  2,
		fonts:    resources.NewFontRegistry(),
		images:   resources.NewImageRegistry(),
		lineWidth: 0.2,
	}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

// SetTitle sets the document title metadata.
func (d *Document) SetTitle(s string) { d.title = s }

// SetAuthor sets the document author metadata.
func (d *Document) SetAuthor(s string) { d.author = s }

// SetSubject sets the document subject metadata.
func (d *Document) SetSubject(s string) { d.subject = s }

// SetCreator sets the document creator metadata.
func (d *Document) SetCreator(s string) { d.creator = s }

// SetMargins sets the left, top, and right margins in user units.
func (d *Document) SetMargins(left, top, right float64) {
	d.lMargin = left
	d.tMargin = top
	d.rMargin = right
}

// SetFont sets the current font. The font is auto-registered if needed.
// family: "helvetica", "courier", "times", "arial", "zapfdingbats"
// style: "", "B", "I", "BI"
// size: font size in points
func (d *Document) SetFont(family, style string, size float64) {
	if d.err != nil {
		return
	}
	fe, err := d.fonts.Register(family, style)
	if err != nil {
		d.err = fmt.Errorf("SetFont: %w", err)
		return
	}
	d.fontFamily = family
	d.fontStyle = style
	d.fontSizePt = size
	d.fontEntry = fe

	// If a page is active, emit the font change to its content stream
	if d.currentPage != nil {
		d.currentPage.applyFont(fe, size)
	}
}

// SetDrawColor sets the stroke color using 0-255 RGB values.
func (d *Document) SetDrawColor(r, g, b int) {
	d.drawColor = state.ColorFromRGB(r, g, b)
	if d.currentPage != nil {
		d.currentPage.stream.SetStrokeColorRGB(d.drawColor.R, d.drawColor.G, d.drawColor.B)
	}
}

// SetFillColor sets the fill color using 0-255 RGB values.
func (d *Document) SetFillColor(r, g, b int) {
	d.fillColor = state.ColorFromRGB(r, g, b)
	if d.currentPage != nil {
		d.currentPage.stream.SetFillColorRGB(d.fillColor.R, d.fillColor.G, d.fillColor.B)
	}
}

// SetTextColor sets the text color using 0-255 RGB values.
func (d *Document) SetTextColor(r, g, b int) {
	d.textColor = state.ColorFromRGB(r, g, b)
}

// SetLineWidth sets the line width in user units.
func (d *Document) SetLineWidth(w float64) {
	d.lineWidth = w
	if d.currentPage != nil {
		d.currentPage.stream.SetLineWidth(w * d.k)
	}
}

// AddPage adds a new page with the given size and returns it.
func (d *Document) AddPage(size PageSize) *Page {
	if d.err != nil {
		return &Page{doc: d}
	}

	p := &Page{
		doc:  d,
		size: size,
		w:    size.WidthPt / d.k,
		h:    size.HeightPt / d.k,
		x:    d.lMargin,
		y:    d.tMargin,
	}

	d.pages = append(d.pages, p)
	d.currentPage = p

	// Emit initial page state
	p.stream.SetLineWidth(d.lineWidth * d.k)
	p.stream.SetLineCap(0)
	p.stream.SetLineJoin(0)

	// Restore font if one was set
	if d.fontEntry != nil {
		p.applyFont(d.fontEntry, d.fontSizePt)
		p.fontFamily = d.fontFamily
		p.fontStyle = d.fontStyle
		p.fontSizePt = d.fontSizePt
		p.fontEntry = d.fontEntry
	}

	// Apply colors if non-default
	if !d.drawColor.IsBlack() {
		p.stream.SetStrokeColorRGB(d.drawColor.R, d.drawColor.G, d.drawColor.B)
	}
	if !d.fillColor.IsBlack() {
		p.stream.SetFillColorRGB(d.fillColor.R, d.fillColor.G, d.fillColor.B)
	}

	return p
}

// RegisterImage registers a JPEG image from a reader for later use.
func (d *Document) RegisterImage(name string, r io.Reader) error {
	if d.err != nil {
		return d.err
	}
	_, err := d.images.RegisterJPEG(name, r)
	if err != nil {
		d.err = fmt.Errorf("RegisterImage: %w", err)
	}
	return err
}

// WriteTo serializes the PDF and writes it to w.
func (d *Document) WriteTo(w io.Writer) (int64, error) {
	if d.err != nil {
		return 0, d.err
	}
	pw, err := d.serialize()
	if err != nil {
		return 0, err
	}
	return pw.WriteTo(w)
}

// Save writes the PDF to a file.
func (d *Document) Save(path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = d.WriteTo(f)
	return err
}

// Bytes returns the serialized PDF as a byte slice.
func (d *Document) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	_, err := d.WriteTo(&buf)
	return buf.Bytes(), err
}

// Err returns the first accumulated error, if any.
func (d *Document) Err() error { return d.err }

// pdfDate formats a time.Time as a PDF date string: D:YYYYMMDDHHmmSS
func pdfDate(t time.Time) string {
	return fmt.Sprintf("D:%04d%02d%02d%02d%02d%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
}

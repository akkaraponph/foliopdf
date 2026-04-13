package folio

import "github.com/akkaraponph/folio/internal/state"

// PageSize defines page dimensions in PDF points.
type PageSize struct {
	WidthPt  float64
	HeightPt float64
}

// Landscape returns a landscape variant of the page size by swapping
// width and height.
func (s PageSize) Landscape() PageSize {
	return PageSize{WidthPt: s.HeightPt, HeightPt: s.WidthPt}
}

// Standard page sizes.
var (
	A3     = PageSize{841.89, 1190.55}
	A4     = PageSize{595.28, 841.89}
	A5     = PageSize{420.94, 595.28}
	Letter = PageSize{612, 792}
	Legal  = PageSize{612, 1008}
)

// Landscape page sizes.
var (
	A3Landscape     = A3.Landscape()
	A4Landscape     = A4.Landscape()
	A5Landscape     = A5.Landscape()
	LetterLandscape = Letter.Landscape()
	LegalLandscape  = Legal.Landscape()
)

// Unit type re-exported for public API.
type Unit = state.Unit

// Measurement units.
const (
	UnitMM   = state.UnitMM   // millimeters (default)
	UnitPt   = state.UnitPt   // points (1/72 inch)
	UnitCM   = state.UnitCM   // centimeters
	UnitInch = state.UnitInch // inches
)

// Option configures a Document.
type Option func(*Document)

// WithUnit sets the measurement unit.
func WithUnit(u Unit) Option {
	return func(d *Document) {
		d.unit = u
		d.k = state.ScaleFactor(u)
	}
}

// WithCompression enables or disables zlib compression of content streams.
func WithCompression(on bool) Option {
	return func(d *Document) {
		d.compress = on
	}
}

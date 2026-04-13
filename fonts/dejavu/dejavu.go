// Package dejavu provides the DejaVu Sans Condensed font family embedded via go:embed.
//
// DejaVu Sans Condensed has broad Unicode coverage including Latin, Greek,
// Cyrillic, Armenian, Georgian, and many other scripts. It is based on
// Bitstream Vera and is freely distributable.
//
// Usage:
//
//	import "github.com/akkaraponph/presspdf/fonts/dejavu"
//
//	doc := presspdf.New()
//	dejavu.Register(doc)           // registers all 4 styles
//	doc.SetFont("dejavu", "", 14)  // use it
package dejavu

import (
	_ "embed"

	"github.com/akkaraponph/presspdf"
)

//go:embed DejaVuSansCondensed.ttf
var regular []byte

//go:embed DejaVuSansCondensed-Bold.ttf
var bold []byte

//go:embed DejaVuSansCondensed-Oblique.ttf
var italic []byte

//go:embed DejaVuSansCondensed-BoldOblique.ttf
var boldItalic []byte

// Regular returns the raw TTF bytes for DejaVu Sans Condensed Regular.
func Regular() []byte { return regular }

// Bold returns the raw TTF bytes for DejaVu Sans Condensed Bold.
func Bold() []byte { return bold }

// Italic returns the raw TTF bytes for DejaVu Sans Condensed Oblique.
func Italic() []byte { return italic }

// BoldItalic returns the raw TTF bytes for DejaVu Sans Condensed Bold Oblique.
func BoldItalic() []byte { return boldItalic }

// Register registers all four DejaVu Sans Condensed styles with the document.
// After calling Register, use doc.SetFont("dejavu", style, size) where
// style is "", "B", "I", or "BI".
func Register(doc *presspdf.Document) error {
	if err := doc.AddUTF8Font("dejavu", "", regular); err != nil {
		return err
	}
	if err := doc.AddUTF8Font("dejavu", "B", bold); err != nil {
		return err
	}
	if err := doc.AddUTF8Font("dejavu", "I", italic); err != nil {
		return err
	}
	if err := doc.AddUTF8Font("dejavu", "BI", boldItalic); err != nil {
		return err
	}
	return nil
}

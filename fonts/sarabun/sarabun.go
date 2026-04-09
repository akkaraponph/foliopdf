// Package sarabun provides the Sarabun Thai font family embedded via go:embed.
//
// Sarabun is a Thai national font licensed under the SIL Open Font License.
// It includes Thai script support with proper mark positioning via OpenType
// GPOS/GSUB tables.
//
// Usage:
//
//	import "github.com/akkaraponph/folio/fonts/sarabun"
//
//	doc := folio.New()
//	sarabun.Register(doc)           // registers all 4 styles
//	doc.SetFont("sarabun", "", 14)  // use it
package sarabun

import (
	_ "embed"

	"github.com/akkaraponph/folio"
)

//go:embed Sarabun-Regular.ttf
var regular []byte

//go:embed Sarabun-Bold.ttf
var bold []byte

//go:embed Sarabun-Italic.ttf
var italic []byte

//go:embed Sarabun-BoldItalic.ttf
var boldItalic []byte

// Regular returns the raw TTF bytes for Sarabun Regular.
func Regular() []byte { return regular }

// Bold returns the raw TTF bytes for Sarabun Bold.
func Bold() []byte { return bold }

// Italic returns the raw TTF bytes for Sarabun Italic.
func Italic() []byte { return italic }

// BoldItalic returns the raw TTF bytes for Sarabun Bold Italic.
func BoldItalic() []byte { return boldItalic }

// Register registers all four Sarabun styles with the document.
// After calling Register, use doc.SetFont("sarabun", style, size) where
// style is "", "B", "I", or "BI".
func Register(doc *folio.Document) error {
	if err := doc.AddUTF8Font("sarabun", "", regular); err != nil {
		return err
	}
	if err := doc.AddUTF8Font("sarabun", "B", bold); err != nil {
		return err
	}
	if err := doc.AddUTF8Font("sarabun", "I", italic); err != nil {
		return err
	}
	if err := doc.AddUTF8Font("sarabun", "BI", boldItalic); err != nil {
		return err
	}
	return nil
}

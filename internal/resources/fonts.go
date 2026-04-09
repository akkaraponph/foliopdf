package resources

import (
	"fmt"
	"strings"
)

// FontEntry represents a registered font.
type FontEntry struct {
	Index  string   // sequential label: "1", "2", ... (used as /F1, /F2)
	Key    string   // lookup key: familyStr + styleStr, e.g. "helveticaB"
	Name   string   // PDF BaseFont name, e.g. "Helvetica-Bold"
	Type   string   // "Core" or "TTF"
	Widths [256]int // character widths in 1/1000 text space units (core fonts)
	Up     int      // underline position
	Ut     int      // underline thickness
	ObjNum int      // set during serialization

	// TTF-specific fields (only set when Type == "TTF")
	TTF          *TTFFont       // parsed TTF data
	UsedRunes    map[int]int    // runes used in the document (for subsetting)
	SubsetData   []byte         // subset font bytes (populated during serialization)
	RuneToGlyph  map[int]int    // unicode -> glyph ID (populated during serialization)
	DescObjNum   int            // FontDescriptor object number
	CIDFontObjNum int           // CIDFont object number
	ToUnicodeObjNum int         // ToUnicode CMap object number
	FileFontObjNum  int         // font file stream object number
}

// FontRegistry manages font registrations and deduplication.
type FontRegistry struct {
	fonts   map[string]*FontEntry // key -> entry
	order   []string              // insertion order for stable iteration
	counter int
}

// NewFontRegistry creates an empty font registry.
func NewFontRegistry() *FontRegistry {
	return &FontRegistry{
		fonts: make(map[string]*FontEntry),
	}
}

// Register registers a core font and returns its entry.
// If already registered, returns the existing entry.
// family is case-insensitive; "arial" maps to "helvetica".
// style is "", "B", "I", or "BI".
func (r *FontRegistry) Register(family, style string) (*FontEntry, error) {
	family = strings.ToLower(family)

	// Apply aliases
	if alias, ok := familyAliases[family]; ok {
		family = alias
	}

	// zapfdingbats has no style variants
	if family == "zapfdingbats" {
		style = ""
	}

	// Normalize style: sort to "BI" not "IB"
	style = normalizeStyle(style)
	key := family + style

	// Return existing entry if already registered
	if fe, ok := r.fonts[key]; ok {
		return fe, nil
	}

	// Look up core font
	cf, ok := coreFonts[key]
	if !ok {
		return nil, fmt.Errorf("unknown core font: %s (family=%q style=%q)", key, family, style)
	}

	r.counter++
	fe := &FontEntry{
		Index:  fmt.Sprintf("%d", r.counter),
		Key:    key,
		Name:   cf.Name,
		Type:   "Core",
		Widths: cf.Widths,
		Up:     cf.Up,
		Ut:     cf.Ut,
	}
	r.fonts[key] = fe
	r.order = append(r.order, key)
	return fe, nil
}

// Get retrieves a registered font entry by family and style.
func (r *FontRegistry) Get(family, style string) (*FontEntry, bool) {
	family = strings.ToLower(family)
	if alias, ok := familyAliases[family]; ok {
		family = alias
	}
	if family == "zapfdingbats" {
		style = ""
	}
	style = normalizeStyle(style)
	fe, ok := r.fonts[family+style]
	return fe, ok
}

// All returns all registered fonts in insertion order.
func (r *FontRegistry) All() []*FontEntry {
	result := make([]*FontEntry, len(r.order))
	for i, key := range r.order {
		result[i] = r.fonts[key]
	}
	return result
}

// StringWidth returns the width of s in 1/1000 text space units.
func StringWidth(fe *FontEntry, s string) int {
	w := 0
	for i := 0; i < len(s); i++ {
		w += fe.Widths[s[i]]
	}
	return w
}

// RegisterTTF registers a TrueType font from parsed TTF data.
// family is the user-facing name (case-insensitive). style is "", "B", "I", or "BI".
func (r *FontRegistry) RegisterTTF(family, style string, ttf *TTFFont) (*FontEntry, error) {
	family = strings.ToLower(family)
	style = normalizeStyle(style)
	key := family + style

	if fe, ok := r.fonts[key]; ok {
		return fe, nil
	}

	r.counter++
	fe := &FontEntry{
		Index:     fmt.Sprintf("%d", r.counter),
		Key:       key,
		Name:      family,
		Type:      "TTF",
		TTF:       ttf,
		UsedRunes: make(map[int]int),
		Up:        int(ttf.UnderlinePosition),
		Ut:        int(ttf.UnderlineThickness),
	}
	r.fonts[key] = fe
	r.order = append(r.order, key)
	return fe, nil
}

// StringWidthUTF8 returns the width of a UTF-8 string in 1/1000 text space units
// using the TTF font's character widths.
func StringWidthUTF8(fe *FontEntry, s string) int {
	if fe.TTF == nil {
		return StringWidth(fe, s)
	}
	w := 0
	for _, r := range s {
		if int(r) < len(fe.TTF.CharWidths) {
			w += fe.TTF.CharWidths[int(r)]
		}
	}
	return w
}

// AddUsedRune marks a rune as used in this font (for subsetting).
func (fe *FontEntry) AddUsedRune(r rune) {
	if fe.UsedRunes != nil {
		fe.UsedRunes[int(r)] = int(r)
	}
}

// AddUsedRunes marks all runes in s as used in this font.
func (fe *FontEntry) AddUsedRunes(s string) {
	if fe.UsedRunes == nil {
		return
	}
	for _, r := range s {
		fe.UsedRunes[int(r)] = int(r)
	}
}

// normalizeStyle normalizes font style to "B", "I", "BI", or "".
func normalizeStyle(style string) string {
	style = strings.ToUpper(style)
	hasB := strings.Contains(style, "B")
	hasI := strings.Contains(style, "I")
	switch {
	case hasB && hasI:
		return "BI"
	case hasB:
		return "B"
	case hasI:
		return "I"
	default:
		return ""
	}
}

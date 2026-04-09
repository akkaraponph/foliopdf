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
	Type   string   // "Core" for standard PDF fonts
	Widths [256]int // character widths in 1/1000 text space units
	Up     int      // underline position
	Ut     int      // underline thickness
	ObjNum int      // set during serialization
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

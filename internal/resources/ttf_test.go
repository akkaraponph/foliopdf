package resources

import (
	"os"
	"testing"
)

func loadTestFont(t *testing.T) []byte {
	t.Helper()
	// Use a system TTF font with cmap format 4 (required)
	paths := []string{
		"/System/Library/Fonts/Supplemental/Georgia.ttf",
		"/System/Library/Fonts/Supplemental/Trebuchet MS Italic.ttf",
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
	}
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err == nil {
			return data
		}
	}
	t.Skip("no suitable TTF font found on this system")
	return nil
}

func TestParseTTF(t *testing.T) {
	data := loadTestFont(t)
	f, err := ParseTTF(data)
	if err != nil {
		t.Fatalf("ParseTTF: %v", err)
	}

	// Basic sanity checks
	if f.unitsPerEm == 0 {
		t.Error("unitsPerEm should be non-zero")
	}
	if f.Ascent == 0 {
		t.Error("Ascent should be non-zero")
	}
	if f.Descent >= 0 {
		t.Error("Descent should be negative")
	}
	if f.CapHeight == 0 {
		t.Error("CapHeight should be non-zero")
	}
	if f.StemV == 0 {
		t.Error("StemV should be non-zero")
	}
	if len(f.CharWidths) != 256*256 {
		t.Errorf("CharWidths length = %d, want %d", len(f.CharWidths), 256*256)
	}

	// 'A' (U+0041) should have a non-default width
	wA := f.CharWidths[0x41]
	if wA == 0 {
		t.Error("width of 'A' should not be zero")
	}

	// Check that glyph mapping exists for ASCII
	if _, ok := f.CharToGlyph[0x41]; !ok {
		t.Error("CharToGlyph should have mapping for 'A'")
	}

	t.Logf("unitsPerEm=%d Ascent=%d Descent=%d CapHeight=%d StemV=%d",
		f.unitsPerEm, f.Ascent, f.Descent, f.CapHeight, f.StemV)
	t.Logf("Bbox=%v Flags=%d ItalicAngle=%d", f.Bbox, f.Flags, f.ItalicAngle)
	t.Logf("width('A')=%d width('W')=%d width(' ')=%d",
		f.CharWidths[0x41], f.CharWidths[0x57], f.CharWidths[0x20])
}

func TestParseTTFInvalid(t *testing.T) {
	_, err := ParseTTF([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	if err == nil {
		t.Error("expected error for invalid font data")
	}
}

func TestSubset(t *testing.T) {
	data := loadTestFont(t)
	f, err := ParseTTF(data)
	if err != nil {
		t.Fatalf("ParseTTF: %v", err)
	}

	// Subset with just "Hello"
	usedRunes := map[int]int{
		0: 'H',
		1: 'e',
		2: 'l',
		3: 'o',
	}

	subsetData, runeToGlyph := f.Subset(usedRunes)
	if subsetData == nil {
		t.Fatal("Subset returned nil")
	}
	if len(subsetData) == 0 {
		t.Fatal("Subset returned empty data")
	}
	if len(subsetData) >= len(data) {
		t.Errorf("subset (%d bytes) should be smaller than original (%d bytes)",
			len(subsetData), len(data))
	}

	// Check runeToGlyph has mappings
	for _, ch := range usedRunes {
		if _, ok := runeToGlyph[ch]; !ok {
			t.Errorf("runeToGlyph missing mapping for %c (U+%04X)", rune(ch), ch)
		}
	}

	// Verify subset is valid TTF
	sf, err := ParseTTF(subsetData)
	if err != nil {
		t.Fatalf("subset is not valid TTF: %v", err)
	}
	if sf.unitsPerEm == 0 {
		t.Error("subset unitsPerEm should be non-zero")
	}

	t.Logf("original=%d bytes, subset=%d bytes, ratio=%.1f%%",
		len(data), len(subsetData), float64(len(subsetData))/float64(len(data))*100)
}

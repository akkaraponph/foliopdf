package resources

import (
	"testing"
)

func TestFontRegistryRegister(t *testing.T) {
	r := NewFontRegistry()

	fe, err := r.Register("Helvetica", "")
	if err != nil {
		t.Fatal(err)
	}
	if fe.Name != "Helvetica" {
		t.Fatalf("Name = %q, want Helvetica", fe.Name)
	}
	if fe.Index != "1" {
		t.Fatalf("Index = %q, want 1", fe.Index)
	}
	if fe.Type != "Core" {
		t.Fatalf("Type = %q, want Core", fe.Type)
	}
}

func TestFontRegistryDedup(t *testing.T) {
	r := NewFontRegistry()

	fe1, _ := r.Register("helvetica", "")
	fe2, _ := r.Register("Helvetica", "")
	if fe1 != fe2 {
		t.Fatal("duplicate registration returned different entries")
	}
	if len(r.All()) != 1 {
		t.Fatalf("All() has %d entries, want 1", len(r.All()))
	}
}

func TestFontRegistryArialAlias(t *testing.T) {
	r := NewFontRegistry()

	fe, err := r.Register("arial", "B")
	if err != nil {
		t.Fatal(err)
	}
	if fe.Name != "Helvetica-Bold" {
		t.Fatalf("Name = %q, want Helvetica-Bold", fe.Name)
	}
}

func TestFontRegistryBoldItalic(t *testing.T) {
	r := NewFontRegistry()

	fe, err := r.Register("times", "BI")
	if err != nil {
		t.Fatal(err)
	}
	if fe.Name != "Times-BoldItalic" {
		t.Fatalf("Name = %q, want Times-BoldItalic", fe.Name)
	}
}

func TestFontRegistryGet(t *testing.T) {
	r := NewFontRegistry()
	r.Register("courier", "I")

	fe, ok := r.Get("Courier", "I")
	if !ok {
		t.Fatal("Get returned not found")
	}
	if fe.Name != "Courier-Oblique" {
		t.Fatalf("Name = %q, want Courier-Oblique", fe.Name)
	}
}

func TestFontRegistryUnknown(t *testing.T) {
	r := NewFontRegistry()
	_, err := r.Register("comic sans", "")
	if err == nil {
		t.Fatal("expected error for unknown font")
	}
}

func TestStringWidth(t *testing.T) {
	r := NewFontRegistry()
	fe, _ := r.Register("helvetica", "")

	w := StringWidth(fe, "H")
	// Helvetica "H" (byte 72) should be 722
	if w != 722 {
		t.Fatalf("StringWidth(H) = %d, want 722", w)
	}

	w = StringWidth(fe, "Hello")
	if w == 0 {
		t.Fatal("StringWidth(Hello) = 0")
	}
}

func TestFontRegistryAll(t *testing.T) {
	r := NewFontRegistry()
	r.Register("helvetica", "")
	r.Register("courier", "B")
	r.Register("times", "I")

	all := r.All()
	if len(all) != 3 {
		t.Fatalf("All() has %d entries, want 3", len(all))
	}
	// Check insertion order
	if all[0].Key != "helvetica" {
		t.Fatalf("first entry key = %q, want helvetica", all[0].Key)
	}
	if all[1].Key != "courierB" {
		t.Fatalf("second entry key = %q, want courierB", all[1].Key)
	}
}

func TestNormalizeStyle(t *testing.T) {
	tests := []struct{ in, want string }{
		{"", ""},
		{"B", "B"},
		{"b", "B"},
		{"I", "I"},
		{"i", "I"},
		{"BI", "BI"},
		{"IB", "BI"},
		{"bi", "BI"},
	}
	for _, tt := range tests {
		got := normalizeStyle(tt.in)
		if got != tt.want {
			t.Errorf("normalizeStyle(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

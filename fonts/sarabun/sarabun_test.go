package sarabun

import (
	"bytes"
	"strings"
	"testing"

	"github.com/akkaraponph/foliopdf"
)

func TestRegister(t *testing.T) {
	doc := foliopdf.New(foliopdf.WithCompression(false))
	if err := Register(doc); err != nil {
		t.Fatalf("Register: %v", err)
	}

	doc.SetFont("sarabun", "", 14)
	page := doc.AddPage(foliopdf.A4)
	page.TextAt(20, 30, "สวัสดี Hello")

	var buf bytes.Buffer
	_, err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	s := buf.String()

	if !strings.Contains(s, "/Subtype /CIDFontType2") {
		t.Error("missing CIDFontType2")
	}
	if !strings.Contains(s, "/Encoding /Identity-H") {
		t.Error("missing Identity-H encoding")
	}
	if !strings.Contains(s, "%%EOF") {
		t.Error("missing EOF")
	}
}

func TestAllStyles(t *testing.T) {
	doc := foliopdf.New(foliopdf.WithCompression(false))
	if err := Register(doc); err != nil {
		t.Fatalf("Register: %v", err)
	}

	styles := []struct {
		style string
		text  string
	}{
		{"", "ปกติ Regular"},
		{"B", "ตัวหนา Bold"},
		{"I", "ตัวเอียง Italic"},
		{"BI", "ตัวหนาเอียง BoldItalic"},
	}

	page := doc.AddPage(foliopdf.A4)
	y := 20.0
	for _, st := range styles {
		doc.SetFont("sarabun", st.style, 14)
		page.TextAt(20, y, st.text)
		y += 15
	}

	var buf bytes.Buffer
	_, err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo: %v", err)
	}

	// Should have 4 TTF font objects
	s := buf.String()
	count := strings.Count(s, "/Subtype /CIDFontType2")
	if count != 4 {
		t.Errorf("expected 4 CIDFontType2, got %d", count)
	}
}

func TestEmbeddedDataPresent(t *testing.T) {
	if len(Regular()) == 0 {
		t.Error("Regular font data is empty")
	}
	if len(Bold()) == 0 {
		t.Error("Bold font data is empty")
	}
	if len(Italic()) == 0 {
		t.Error("Italic font data is empty")
	}
	if len(BoldItalic()) == 0 {
		t.Error("BoldItalic font data is empty")
	}
}

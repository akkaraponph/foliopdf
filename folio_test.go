package folio

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"strings"
	"testing"
)

func TestEmptyPage(t *testing.T) {
	doc := New()
	doc.AddPage(A4)

	var buf bytes.Buffer
	_, err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	s := buf.String()

	// Verify PDF structure
	if !strings.HasPrefix(s, "%PDF-1.4") {
		t.Error("missing PDF header")
	}
	if !strings.Contains(s, "/Type /Page") {
		t.Error("missing Page object")
	}
	if !strings.Contains(s, "/Type /Pages") {
		t.Error("missing Pages root")
	}
	if !strings.Contains(s, "/Type /Catalog") {
		t.Error("missing Catalog")
	}
	if !strings.Contains(s, "xref") {
		t.Error("missing xref")
	}
	if !strings.Contains(s, "%%EOF") {
		t.Error("missing EOF marker")
	}
}

func TestTextAt(t *testing.T) {
	doc := New(WithCompression(false))
	doc.SetFont("helvetica", "", 16)
	page := doc.AddPage(A4)
	page.TextAt(40, 60, "Hello Folio")

	var buf bytes.Buffer
	_, err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	s := buf.String()

	// Should contain text operators
	if !strings.Contains(s, "BT") {
		t.Error("missing BT operator")
	}
	if !strings.Contains(s, "Hello Folio") {
		t.Error("missing text content")
	}
	if !strings.Contains(s, "Tj") {
		t.Error("missing Tj operator")
	}
	if !strings.Contains(s, "ET") {
		t.Error("missing ET operator")
	}
}

func TestTargetAPI(t *testing.T) {
	doc := New(WithCompression(false))
	doc.SetTitle("Invoice")
	doc.SetAuthor("Akkarapon")
	doc.SetFont("helvetica", "", 16)

	page := doc.AddPage(A4)
	page.TextAt(40, 60, "Hello Folio")
	page.Line(40, 70, 200, 70)

	var buf bytes.Buffer
	_, err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	s := buf.String()

	if !strings.Contains(s, "/Title (Invoice)") {
		t.Error("missing title")
	}
	if !strings.Contains(s, "/Author (Akkarapon)") {
		t.Error("missing author")
	}
	if !strings.Contains(s, "Hello Folio") {
		t.Error("missing text")
	}
	// Line should have moveto/lineto/stroke
	if !strings.Contains(s, " m\n") {
		t.Error("missing line moveto")
	}
	if !strings.Contains(s, " l\n") {
		t.Error("missing line lineto")
	}
}

func TestMultiplePages(t *testing.T) {
	doc := New(WithCompression(false))
	doc.SetFont("courier", "", 12)

	p1 := doc.AddPage(A4)
	p1.TextAt(20, 20, "Page 1")

	p2 := doc.AddPage(Letter)
	p2.TextAt(20, 20, "Page 2")

	var buf bytes.Buffer
	_, err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	s := buf.String()

	if strings.Count(s, "/Type /Page\n") != 2 {
		t.Errorf("expected 2 Page objects, got %d", strings.Count(s, "/Type /Page\n"))
	}
	if !strings.Contains(s, "/Count 2") {
		t.Error("Pages count should be 2")
	}
	if !strings.Contains(s, "Page 1") || !strings.Contains(s, "Page 2") {
		t.Error("missing page text content")
	}
}

func TestRect(t *testing.T) {
	doc := New(WithCompression(false))
	page := doc.AddPage(A4)
	page.Rect(20, 50, 170, 100, "D")

	var buf bytes.Buffer
	_, err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	s := buf.String()

	if !strings.Contains(s, "re\nS") {
		t.Error("missing rect + stroke operators")
	}
}

func TestRectFill(t *testing.T) {
	doc := New(WithCompression(false))
	doc.SetFillColor(255, 0, 0)
	page := doc.AddPage(A4)
	page.Rect(20, 50, 170, 100, "F")

	var buf bytes.Buffer
	_, err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	s := buf.String()

	if !strings.Contains(s, "re\nf") {
		t.Error("missing rect + fill operators")
	}
}

func TestFontDedup(t *testing.T) {
	doc := New()
	doc.SetFont("helvetica", "", 12)
	doc.SetFont("helvetica", "", 16) // same font, different size
	doc.AddPage(A4)

	var buf bytes.Buffer
	_, err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	s := buf.String()

	// Should only have one font object
	if strings.Count(s, "/BaseFont /Helvetica") != 1 {
		t.Error("font should be deduplicated")
	}
}

func TestNoPages(t *testing.T) {
	doc := New()
	var buf bytes.Buffer
	_, err := doc.WriteTo(&buf)
	if err == nil {
		t.Error("expected error for no pages")
	}
}

func TestTextColor(t *testing.T) {
	doc := New(WithCompression(false))
	doc.SetFont("helvetica", "", 12)
	doc.SetTextColor(255, 0, 0)
	page := doc.AddPage(A4)
	page.TextAt(20, 20, "Red text")

	var buf bytes.Buffer
	_, err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	s := buf.String()

	if !strings.Contains(s, "1.000 0.000 0.000 rg") {
		t.Error("missing red fill color for text")
	}
}

func TestGetStringWidth(t *testing.T) {
	doc := New()
	doc.SetFont("helvetica", "", 12)
	page := doc.AddPage(A4)

	w := page.GetStringWidth("Hello")
	if w <= 0 {
		t.Fatalf("GetStringWidth(Hello) = %f, want > 0", w)
	}
	// Helvetica 12pt "Hello" should be roughly 9-15mm
	if w < 5 || w > 20 {
		t.Errorf("GetStringWidth(Hello) = %f, seems unreasonable", w)
	}
}

func TestEscapeParens(t *testing.T) {
	doc := New(WithCompression(false))
	doc.SetFont("helvetica", "", 12)
	page := doc.AddPage(A4)
	page.TextAt(20, 20, "test (parens) here")

	var buf bytes.Buffer
	_, err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	s := buf.String()

	if !strings.Contains(s, `test \(parens\) here`) {
		t.Error("parentheses not escaped in text")
	}
}

func TestCell(t *testing.T) {
	doc := New(WithCompression(false))
	doc.SetFont("helvetica", "", 12)
	page := doc.AddPage(A4)
	page.SetXY(10, 10)
	page.Cell(50, 10, "Cell 1", "1", "L", false, 0)
	page.Cell(50, 10, "Cell 2", "", "C", false, 1)

	var buf bytes.Buffer
	_, err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	if !strings.Contains(s, "Cell 1") {
		t.Error("missing Cell 1 text")
	}
	if !strings.Contains(s, "Cell 2") {
		t.Error("missing Cell 2 text")
	}
}

func TestMultiCell(t *testing.T) {
	doc := New(WithCompression(false))
	doc.SetFont("helvetica", "", 12)
	page := doc.AddPage(A4)
	page.SetXY(10, 10)
	page.MultiCell(80, 6, "This is a long text that should be wrapped into multiple lines when it exceeds the width.", "", "L", false)

	var buf bytes.Buffer
	_, err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	// Should have produced output with text operators
	s := buf.String()
	if !strings.Contains(s, "Tj") {
		t.Error("missing text output from MultiCell")
	}
}

func TestMultiCellNewlines(t *testing.T) {
	doc := New(WithCompression(false))
	doc.SetFont("helvetica", "", 12)
	page := doc.AddPage(A4)
	page.SetXY(10, 10)
	page.MultiCell(0, 6, "Line 1\nLine 2\nLine 3", "", "L", false)

	var buf bytes.Buffer
	_, err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	if !strings.Contains(s, "Line 1") || !strings.Contains(s, "Line 2") || !strings.Contains(s, "Line 3") {
		t.Error("missing newline-separated text")
	}
}

func testJPEG(w, h int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 128, A: 255})
		}
	}
	var buf bytes.Buffer
	jpeg.Encode(&buf, img, &jpeg.Options{Quality: 50})
	return buf.Bytes()
}

func TestJPEGImage(t *testing.T) {
	doc := New(WithCompression(false))
	doc.SetFont("helvetica", "", 12)

	jpgData := testJPEG(100, 80)
	err := doc.RegisterImage("photo", bytes.NewReader(jpgData))
	if err != nil {
		t.Fatal(err)
	}

	page := doc.AddPage(A4)
	page.TextAt(20, 20, "Image below:")
	page.DrawImageRect("photo", 20, 30, 60, 48)

	var buf bytes.Buffer
	_, err = doc.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	s := buf.String()

	if !strings.Contains(s, "/Type /XObject") {
		t.Error("missing XObject")
	}
	if !strings.Contains(s, "/Subtype /Image") {
		t.Error("missing /Subtype /Image")
	}
	if !strings.Contains(s, "/Filter /DCTDecode") {
		t.Error("missing DCTDecode filter")
	}
	if !strings.Contains(s, "/Im1 Do") {
		t.Error("missing image draw operator")
	}
}

func TestImageDedup(t *testing.T) {
	doc := New(WithCompression(false))

	jpgData := testJPEG(50, 50)
	doc.RegisterImage("img1", bytes.NewReader(jpgData))
	doc.RegisterImage("img2", bytes.NewReader(jpgData)) // same data

	page := doc.AddPage(A4)
	page.DrawImageRect("img1", 10, 10, 30, 30)
	page.DrawImageRect("img2", 50, 10, 30, 30)

	var buf bytes.Buffer
	_, err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	s := buf.String()

	// Should only have one image XObject
	count := strings.Count(s, "/Type /XObject")
	if count != 1 {
		t.Errorf("expected 1 image XObject, got %d", count)
	}
}

func TestImageNotRegistered(t *testing.T) {
	doc := New()
	page := doc.AddPage(A4)
	page.DrawImageRect("nonexistent", 10, 10, 30, 30)

	_, err := doc.WriteTo(&bytes.Buffer{})
	if err == nil {
		t.Error("expected error for unregistered image")
	}
}

func TestBytes(t *testing.T) {
	doc := New()
	doc.AddPage(A4)

	b, err := doc.Bytes()
	if err != nil {
		t.Fatal(err)
	}
	if len(b) == 0 {
		t.Error("Bytes() returned empty")
	}
	if !strings.HasPrefix(string(b), "%PDF") {
		t.Error("Bytes() not a valid PDF")
	}
}

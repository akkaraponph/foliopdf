package folio

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"os"
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

func loadTTFFont(t *testing.T) []byte {
	t.Helper()
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

func TestUTF8Font(t *testing.T) {
	data := loadTTFFont(t)

	doc := New(WithCompression(false))
	err := doc.AddUTF8Font("georgia", "", data)
	if err != nil {
		t.Fatalf("AddUTF8Font: %v", err)
	}

	doc.SetFont("georgia", "", 14)
	page := doc.AddPage(A4)
	page.TextAt(20, 30, "Hello UTF-8 World")

	var buf bytes.Buffer
	_, err = doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	s := buf.String()

	// Verify CIDFont Type2 structure
	if !strings.Contains(s, "/Subtype /Type0") {
		t.Error("missing Type0 font")
	}
	if !strings.Contains(s, "/Subtype /CIDFontType2") {
		t.Error("missing CIDFontType2")
	}
	if !strings.Contains(s, "/Encoding /Identity-H") {
		t.Error("missing Identity-H encoding")
	}
	if !strings.Contains(s, "/Type /FontDescriptor") {
		t.Error("missing FontDescriptor")
	}
	if !strings.Contains(s, "/FontFile2") {
		t.Error("missing FontFile2")
	}
	if !strings.Contains(s, "/CIDToGIDMap") {
		t.Error("missing CIDToGIDMap")
	}
	if !strings.Contains(s, "CIDInit") {
		t.Error("missing ToUnicode CMap")
	}

	// Text should be hex-encoded, not ASCII
	if !strings.Contains(s, "Tj") {
		t.Error("missing Tj operator")
	}
	// Hex string should use angle brackets
	if !strings.Contains(s, "<") {
		t.Error("missing hex string in text output")
	}

	// Verify PDF structure
	if !strings.HasPrefix(s, "%PDF-1.4") {
		t.Error("missing PDF header")
	}
	if !strings.Contains(s, "%%EOF") {
		t.Error("missing EOF marker")
	}
}

func TestUTF8FontWidth(t *testing.T) {
	data := loadTTFFont(t)

	doc := New(WithCompression(false))
	doc.AddUTF8Font("georgia", "", data)
	doc.SetFont("georgia", "", 12)
	page := doc.AddPage(A4)

	w := page.GetStringWidth("Hello")
	if w <= 0 {
		t.Fatalf("GetStringWidth(Hello) = %f, want > 0", w)
	}
	if w < 5 || w > 20 {
		t.Errorf("GetStringWidth(Hello) = %f, seems unreasonable", w)
	}
}

func TestUTF8MultiCell(t *testing.T) {
	data := loadTTFFont(t)

	doc := New(WithCompression(false))
	doc.AddUTF8Font("georgia", "", data)
	doc.SetFont("georgia", "", 10)
	page := doc.AddPage(A4)
	page.SetXY(10, 10)
	page.MultiCell(80, 6, "This is a long UTF-8 text that should wrap into multiple lines when it exceeds the width.", "", "L", false)

	var buf bytes.Buffer
	_, err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	s := buf.String()
	if !strings.Contains(s, "Tj") {
		t.Error("missing text output from UTF-8 MultiCell")
	}
}

func TestMixedFonts(t *testing.T) {
	data := loadTTFFont(t)

	doc := New(WithCompression(false))
	doc.AddUTF8Font("georgia", "", data)

	doc.SetFont("helvetica", "", 12)
	page := doc.AddPage(A4)
	page.TextAt(20, 20, "Core font text")

	doc.SetFont("georgia", "", 12)
	page.TextAt(20, 40, "UTF-8 font text")

	var buf bytes.Buffer
	_, err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	s := buf.String()

	// Should have both core and TTF font objects
	if !strings.Contains(s, "/Subtype /Type1") {
		t.Error("missing Type1 font (core)")
	}
	if !strings.Contains(s, "/Subtype /Type0") {
		t.Error("missing Type0 font (TTF)")
	}
	// Core font text uses parentheses, TTF uses hex
	if !strings.Contains(s, "(Core font text)") {
		t.Error("missing core font text with parenthesis encoding")
	}
}

func TestSetFontSize(t *testing.T) {
	doc := New(WithCompression(false))
	doc.SetFont("helvetica", "", 12)

	p1 := doc.AddPage(A4)
	w1 := p1.GetStringWidth("Hello")

	doc.SetFontSize(24)
	p2 := doc.AddPage(A4)
	w2 := p2.GetStringWidth("Hello")

	if w2 <= w1 {
		t.Errorf("expected larger width at 24pt (%f) than 12pt (%f)", w2, w1)
	}
	if doc.GetFontSize() != 24 {
		t.Errorf("GetFontSize() = %f, want 24", doc.GetFontSize())
	}
}

func TestSetFontStyle(t *testing.T) {
	doc := New(WithCompression(false))
	doc.SetFont("helvetica", "", 12)
	page := doc.AddPage(A4)
	page.TextAt(20, 20, "Normal")

	doc.SetFontStyle("B")
	page.TextAt(20, 40, "Bold")

	if doc.GetFontStyle() != "B" {
		t.Errorf("GetFontStyle() = %q, want %q", doc.GetFontStyle(), "B")
	}
	if doc.GetFontFamily() != "helvetica" {
		t.Errorf("GetFontFamily() = %q, want %q", doc.GetFontFamily(), "helvetica")
	}

	var buf bytes.Buffer
	_, err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	if !strings.Contains(s, "(Normal)") {
		t.Error("missing Normal text")
	}
	if !strings.Contains(s, "(Bold)") {
		t.Error("missing Bold text")
	}
	// Should have both Helvetica and Helvetica-Bold
	if !strings.Contains(s, "/BaseFont /Helvetica\n") {
		t.Error("missing Helvetica font")
	}
	if !strings.Contains(s, "/BaseFont /Helvetica-Bold") {
		t.Error("missing Helvetica-Bold font")
	}
}

func TestPageSetFontSize(t *testing.T) {
	doc := New(WithCompression(false))
	doc.SetFont("helvetica", "", 12)
	page := doc.AddPage(A4)

	page.SetFontSize(20)
	if page.GetFontSize() != 20 {
		t.Errorf("page.GetFontSize() = %f, want 20", page.GetFontSize())
	}
	// Document font size should be unchanged
	if doc.GetFontSize() != 12 {
		t.Errorf("doc.GetFontSize() = %f, want 12 (unchanged)", doc.GetFontSize())
	}
}

func TestPageSetFontStyle(t *testing.T) {
	doc := New(WithCompression(false))
	doc.SetFont("helvetica", "", 12)
	page := doc.AddPage(A4)

	page.TextAt(20, 20, "Regular")
	page.SetFontStyle("I")
	page.TextAt(20, 40, "Italic")

	if page.GetFontStyle() != "I" {
		t.Errorf("page.GetFontStyle() = %q, want %q", page.GetFontStyle(), "I")
	}
	if page.GetFontFamily() != "helvetica" {
		t.Errorf("page.GetFontFamily() = %q, want %q", page.GetFontFamily(), "helvetica")
	}

	var buf bytes.Buffer
	_, err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	if !strings.Contains(s, "(Regular)") {
		t.Error("missing Regular text")
	}
	if !strings.Contains(s, "(Italic)") {
		t.Error("missing Italic text")
	}
}

func TestSetFontSizeNoFont(t *testing.T) {
	doc := New()
	doc.SetFontSize(12)
	if doc.Err() == nil {
		t.Error("expected error when setting size with no font")
	}
}

func TestSetFontStyleNoFont(t *testing.T) {
	doc := New()
	doc.SetFontStyle("B")
	if doc.Err() == nil {
		t.Error("expected error when setting style with no font")
	}
}

func TestAddUTF8FontFromFile(t *testing.T) {
	paths := []string{
		"/System/Library/Fonts/Supplemental/Georgia.ttf",
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
	}
	var fontPath string
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			fontPath = p
			break
		}
	}
	if fontPath == "" {
		t.Skip("no suitable TTF font found on this system")
	}

	doc := New(WithCompression(false))
	err := doc.AddUTF8FontFromFile("testfont", "", fontPath)
	if err != nil {
		t.Fatalf("AddUTF8FontFromFile: %v", err)
	}
	doc.SetFont("testfont", "", 14)
	page := doc.AddPage(A4)
	page.TextAt(20, 30, "From file")

	var buf bytes.Buffer
	_, err = doc.WriteTo(&buf)
	if err != nil {
		t.Fatalf("WriteTo: %v", err)
	}
	if !strings.Contains(buf.String(), "/Subtype /Type0") {
		t.Error("missing Type0 font")
	}
}

// --- Auto page break tests ---

func TestAutoPageBreak(t *testing.T) {
	doc := New(WithCompression(false))
	doc.SetFont("helvetica", "", 12)
	doc.SetAutoPageBreak(true, 15)
	page := doc.AddPage(A4)

	// A4 = 297mm height. Top margin 10mm, bottom margin 15mm.
	// Available: 272mm. Cell height: 10mm. 27 cells fit on page 1.
	// Cell 28 triggers a break → page 2.
	for i := 0; i < 30; i++ {
		page.Cell(0, 10, fmt.Sprintf("Line %d", i+1), "", "L", false, 1)
	}

	if n := doc.PageCount(); n != 2 {
		t.Errorf("expected 2 pages, got %d", n)
	}

	var buf bytes.Buffer
	_, err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	s := buf.String()
	if !strings.Contains(s, "/Count 2") {
		t.Error("PDF should have 2 pages")
	}
}

func TestAutoPageBreakMultiplePages(t *testing.T) {
	doc := New(WithCompression(false))
	doc.SetFont("helvetica", "", 12)
	doc.SetAutoPageBreak(true, 15)
	page := doc.AddPage(A4)

	// 100 cells × 10mm. ~27 cells/page → 4 pages.
	for i := 0; i < 100; i++ {
		page.Cell(0, 10, "X", "", "L", false, 1)
	}

	if n := doc.PageCount(); n != 4 {
		t.Errorf("expected 4 pages, got %d", n)
	}

	var buf bytes.Buffer
	_, err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAutoPageBreakForwarding(t *testing.T) {
	doc := New()
	doc.SetFont("helvetica", "", 12)
	doc.SetAutoPageBreak(true, 15)
	page := doc.AddPage(A4)

	// Fill page 1 to trigger a break.
	for i := 0; i < 30; i++ {
		page.Cell(0, 10, "X", "", "L", false, 1)
	}

	// The original page variable should forward to the active page.
	// After 30 cells (27 on page 1, 3 on page 2), cursor should be
	// near the top of page 2.
	y := page.GetY()
	if y < 20 || y > 50 {
		t.Errorf("expected cursor near top of page 2, got y=%f", y)
	}

	// CurrentPage should be page 2.
	if doc.CurrentPage() == nil {
		t.Fatal("CurrentPage should not be nil")
	}
	if doc.PageCount() != 2 {
		t.Errorf("expected 2 pages, got %d", doc.PageCount())
	}
}

func TestAutoPageBreakDisabled(t *testing.T) {
	doc := New(WithCompression(false))
	doc.SetFont("helvetica", "", 12)
	// Auto page break NOT enabled (default).
	page := doc.AddPage(A4)

	for i := 0; i < 50; i++ {
		page.Cell(0, 10, "X", "", "L", false, 1)
	}

	// All content on one page — no automatic breaks.
	if n := doc.PageCount(); n != 1 {
		t.Errorf("expected 1 page (no auto break), got %d", n)
	}
}

func TestAutoPageBreakMultiCell(t *testing.T) {
	doc := New(WithCompression(false))
	doc.SetFont("helvetica", "", 12)
	doc.SetAutoPageBreak(true, 15)
	page := doc.AddPage(A4)

	// Write a very long text via MultiCell. Each line is ~6mm tall.
	// With 272mm available, ~45 lines fit per page.
	long := strings.Repeat("This is a line of text that will be wrapped. ", 120)
	page.MultiCell(0, 6, long, "", "L", false)

	if n := doc.PageCount(); n < 2 {
		t.Errorf("expected at least 2 pages from long MultiCell, got %d", n)
	}

	var buf bytes.Buffer
	_, err := doc.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPageBreakTrigger(t *testing.T) {
	doc := New()
	doc.SetAutoPageBreak(true, 20)
	page := doc.AddPage(A4)

	// A4 height in mm ≈ 297. Trigger = 297 - 20 = 277.
	trigger := page.PageBreakTrigger()
	if trigger < 276 || trigger > 278 {
		t.Errorf("expected trigger ~277, got %f", trigger)
	}
}

func TestCurrentPageAndPageCount(t *testing.T) {
	doc := New()
	if doc.PageCount() != 0 {
		t.Errorf("expected 0 pages initially, got %d", doc.PageCount())
	}

	p1 := doc.AddPage(A4)
	if doc.PageCount() != 1 {
		t.Errorf("expected 1 page, got %d", doc.PageCount())
	}
	if doc.CurrentPage() != p1 {
		t.Error("CurrentPage should be p1")
	}

	p2 := doc.AddPage(A4)
	if doc.PageCount() != 2 {
		t.Errorf("expected 2 pages, got %d", doc.PageCount())
	}
	if doc.CurrentPage() != p2 {
		t.Error("CurrentPage should be p2")
	}
}

package presspdf_test

import (
	"bytes"
	"fmt"

	"github.com/akkaraponph/presspdf"
)

func Example() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 16)

	page := doc.AddPage(presspdf.A4)
	page.TextAt(20, 20, "Hello, PressPDF!")

	var buf bytes.Buffer
	_, err := doc.WriteTo(&buf)
	if err != nil {
		panic(err)
	}
	fmt.Println("PDF size:", buf.Len(), "bytes")
	// Output: PDF size: 873 bytes
}

func ExampleNew() {
	// Create a document with default settings (mm, A4, compression on).
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 12)
	doc.AddPage(presspdf.A4)

	if err := doc.Err(); err != nil {
		panic(err)
	}
	fmt.Println("ok")
	// Output: ok
}

func ExampleNew_withOptions() {
	// Create a document with custom units and no compression.
	doc := presspdf.New(
		presspdf.WithUnit(presspdf.UnitPt),
		presspdf.WithCompression(false),
	)
	doc.SetFont("helvetica", "", 12)
	doc.AddPage(presspdf.A4)

	if err := doc.Err(); err != nil {
		panic(err)
	}
	fmt.Println("ok")
	// Output: ok
}

func ExampleDocument_SetFont() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 12)     // regular
	doc.SetFont("helvetica", "B", 14)    // bold
	doc.SetFont("helvetica", "BI", 10)   // bold italic
	doc.SetFont("times", "I", 12)        // Times italic
	doc.SetFont("courier", "", 10)       // monospace

	fmt.Println(doc.GetFontFamily())
	// Output: courier
}

func ExampleDocument_SetMargins() {
	doc := presspdf.New()
	doc.SetMargins(15, 20, 15) // left, top, right

	l, t, r, _ := doc.GetMargins()
	fmt.Printf("margins: %.0f %.0f %.0f\n", l, t, r)
	// Output: margins: 15 20 15
}

func ExampleDocument_SetHeaderFunc() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 12)

	doc.SetHeaderFunc(func(p *presspdf.Page) {
		p.SetFont("helvetica", "B", 10)
		p.Cell(0, 8, "Report Header", "", "C", false, 0)
		p.Ln(10)
	})

	doc.AddPage(presspdf.A4)
	if err := doc.Err(); err != nil {
		panic(err)
	}
	fmt.Println("ok")
	// Output: ok
}

func ExampleDocument_AddBookmark() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 14)

	doc.AddPage(presspdf.A4)
	doc.AddBookmark("Chapter 1", 0)

	doc.AddPage(presspdf.A4)
	doc.AddBookmark("Chapter 2", 0)

	fmt.Println("bookmarks added")
	// Output: bookmarks added
}

func ExampleDocument_SetProtection() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 12)
	doc.AddPage(presspdf.A4)
	doc.SetProtection("", "owner123", presspdf.PermPrint|presspdf.PermCopy)

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		panic(err)
	}
	fmt.Println("encrypted PDF:", buf.Len() > 0)
	// Output: encrypted PDF: true
}

func ExamplePage_TextAt() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 16)
	page := doc.AddPage(presspdf.A4)

	page.TextAt(20, 30, "Hello World")

	if err := doc.Err(); err != nil {
		panic(err)
	}
	fmt.Println("ok")
	// Output: ok
}

func ExamplePage_Cell() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 12)
	page := doc.AddPage(presspdf.A4)

	// Cell(width, height, text, border, align, fill, lineBreak)
	page.Cell(60, 10, "Left aligned", "1", "L", false, 0)
	page.Cell(60, 10, "Centered", "1", "C", false, 0)
	page.Cell(60, 10, "Right aligned", "1", "R", false, 1)

	fmt.Println("ok")
	// Output: ok
}

func ExamplePage_MultiCell() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 12)
	page := doc.AddPage(presspdf.A4)

	text := "This is a long paragraph that will automatically wrap " +
		"within the given width. MultiCell handles line breaking."
	page.MultiCell(100, 6, text, "1", "J", false)

	fmt.Println("ok")
	// Output: ok
}

func ExamplePage_Write() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 12)
	page := doc.AddPage(presspdf.A4)

	// Write allows inline mixed formatting.
	page.Write(6, "This is normal. ")
	doc.SetFontStyle("B")
	page.Write(6, "This is bold. ")
	doc.SetFontStyle("")
	page.Write(6, "Back to normal.")

	fmt.Println("ok")
	// Output: ok
}

func ExamplePage_Line() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 12)
	page := doc.AddPage(presspdf.A4)

	doc.SetDrawColor(255, 0, 0) // red
	page.Line(20, 50, 190, 50)

	fmt.Println("ok")
	// Output: ok
}

func ExamplePage_Rect() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 12)
	page := doc.AddPage(presspdf.A4)

	doc.SetFillColor(200, 220, 255)
	page.Rect(20, 20, 80, 40, "DF") // draw and fill

	fmt.Println("ok")
	// Output: ok
}

func ExamplePage_Circle() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 12)
	page := doc.AddPage(presspdf.A4)

	page.Circle(100, 100, 30, "D") // stroke only

	fmt.Println("ok")
	// Output: ok
}

func ExamplePage_Polygon() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 12)
	page := doc.AddPage(presspdf.A4)

	// Draw a triangle.
	page.Polygon([]presspdf.Point{
		presspdf.Pt(100, 30),
		presspdf.Pt(60, 90),
		presspdf.Pt(140, 90),
	}, "D")

	fmt.Println("ok")
	// Output: ok
}

func ExamplePage_MoveTo() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 12)
	page := doc.AddPage(presspdf.A4)

	// Build a custom path.
	page.MoveTo(20, 20)
	page.LineTo(80, 20)
	page.LineTo(50, 60)
	page.ClosePath()
	page.DrawPath("DF")

	fmt.Println("ok")
	// Output: ok
}

func ExampleNewTable() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 10)
	page := doc.AddPage(presspdf.A4)

	tbl := presspdf.NewTable(doc, page)
	tbl.SetWidths(60, 40, 30)
	tbl.SetHeaderStyle(presspdf.CellStyle{
		FillColor: [3]int{41, 128, 185},
		TextColor: [3]int{255, 255, 255},
		Fill:      true,
	})

	tbl.Header("Product", "Qty", "Price")
	tbl.Row("Widget A", "100", "$5.00")
	tbl.Row("Widget B", "50", "$8.50")

	fmt.Println("ok")
	// Output: ok
}

func ExamplePage_Barcode128() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 12)
	page := doc.AddPage(presspdf.A4)

	page.Barcode128(20, 20, 80, 30, "PRESSPDF-001")

	fmt.Println("ok")
	// Output: ok
}

func ExamplePage_QRCode() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 12)
	page := doc.AddPage(presspdf.A4)

	page.QRCode(20, 20, 50, "https://github.com/akkaraponph/presspdf", presspdf.ECMedium)

	fmt.Println("ok")
	// Output: ok
}

func ExamplePage_LinearGradient() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 12)
	page := doc.AddPage(presspdf.A4)

	// Horizontal gradient from red to blue.
	page.LinearGradient(20, 20, 170, 60, 0, 0, 1, 0,
		presspdf.GradientStop(0, 255, 0, 0),
		presspdf.GradientStop(1, 0, 0, 255),
	)

	fmt.Println("ok")
	// Output: ok
}

func ExamplePage_ClipCircle() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 12)
	page := doc.AddPage(presspdf.A4)

	// Clip to a circle, then draw inside it.
	page.ClipCircle(100, 100, 40)
	doc.SetFillColor(255, 200, 200)
	page.Rect(60, 60, 80, 80, "F") // only visible inside circle
	page.ClipEnd()

	fmt.Println("ok")
	// Output: ok
}

func ExamplePage_Rotate() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 14)
	page := doc.AddPage(presspdf.A4)

	page.TransformBegin()
	page.Rotate(30, 100, 100)
	page.TextAt(100, 100, "Rotated 30 degrees")
	page.TransformEnd()

	fmt.Println("ok")
	// Output: ok
}

func ExampleDocument_SetAlpha() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 12)
	page := doc.AddPage(presspdf.A4)

	doc.SetAlpha(0.5) // 50% transparent
	doc.SetFillColor(255, 0, 0)
	page.Rect(20, 20, 80, 40, "F")

	doc.SetAlpha(1.0) // restore full opacity
	page.TextAt(30, 35, "Semi-transparent box behind this text")

	fmt.Println("ok")
	// Output: ok
}

func ExampleDocument_AddLink() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 12)

	// Create a forward link before the target exists.
	linkID := doc.AddLink()

	page1 := doc.AddPage(presspdf.A4)
	page1.WriteLinkID(6, "Go to page 2", linkID)

	page2 := doc.AddPage(presspdf.A4)
	doc.SetLink(linkID, 0, 2) // target: page 2, top
	page2.TextAt(20, 20, "You are on page 2")

	fmt.Println("ok")
	// Output: ok
}

func ExamplePage_Ln() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 12)
	page := doc.AddPage(presspdf.A4)

	page.Cell(0, 10, "Line 1", "", "L", false, 0)
	page.Ln(10) // advance to next line
	page.Cell(0, 10, "Line 2", "", "L", false, 0)

	fmt.Println("ok")
	// Output: ok
}

func ExampleDocument_AliasNbPages() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 12)
	doc.AliasNbPages("{nb}") // placeholder for total pages

	doc.SetFooterFunc(func(p *presspdf.Page) {
		p.SetY(-15)
		p.Cell(0, 10, fmt.Sprintf("Page %d of {nb}", doc.PageNo()), "", "C", false, 0)
	})

	doc.AddPage(presspdf.A4)
	doc.AddPage(presspdf.A4)

	fmt.Println("ok")
	// Output: ok
}

func ExamplePage_Paragraph() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 11)
	page := doc.AddPage(presspdf.A4)

	page.Paragraph("This is a nicely formatted paragraph with justified "+
		"alignment, first-line indent, and spacing before and after.",
		presspdf.ParagraphAlign("J"),
		presspdf.ParagraphIndent(10),
		presspdf.ParagraphSpaceBefore(5),
		presspdf.ParagraphSpaceAfter(5),
	)

	fmt.Println("ok")
	// Output: ok
}

func ExamplePage_SVGPath() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 12)
	page := doc.AddPage(presspdf.A4)

	// Draw a heart shape using SVG path data.
	page.SVGPath(50, 50, 0.5,
		"M 10,30 A 20,20 0,0,1 50,30 A 20,20 0,0,1 90,30 "+
			"Q 90,60 50,90 Q 10,60 10,30 Z",
		"F")

	fmt.Println("ok")
	// Output: ok
}

func ExampleNewColumnLayout() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 10)
	page := doc.AddPage(presspdf.A4)

	cols := presspdf.NewColumnLayout(doc, page, 2, 10) // 2 columns, 10mm gutter
	cols.Begin()
	page.Write(5, "Left column text...")
	cols.NextColumn()
	page.Write(5, "Right column text...")
	cols.End()

	fmt.Println("ok")
	// Output: ok
}

func ExampleNewTOC() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 12)

	toc := presspdf.NewTOC(doc)

	page1 := doc.AddPage(presspdf.A4)
	toc.Add("Introduction", 0, page1, page1.GetY())
	page1.TextAt(20, 30, "Introduction content...")

	page2 := doc.AddPage(presspdf.A4)
	toc.Add("Chapter 1", 0, page2, page2.GetY())
	page2.TextAt(20, 30, "Chapter 1 content...")

	fmt.Println("ok")
	// Output: ok
}

func ExamplePage_FormTextField() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 12)
	page := doc.AddPage(presspdf.A4)

	page.TextAt(20, 20, "Name:")
	page.FormTextField("name", 50, 15, 80, 12)

	page.TextAt(20, 40, "Email:")
	page.FormTextField("email", 50, 35, 80, 12,
		presspdf.WithDefaultValue("user@example.com"))

	fmt.Println("ok")
	// Output: ok
}

func ExamplePage_RichText() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 12)
	page := doc.AddPage(presspdf.A4)

	page.RichText(6, "This is <b>bold</b>, <i>italic</i>, and <b><i>both</i></b>.")

	fmt.Println("ok")
	// Output: ok
}

func ExamplePage_HTML() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 12)
	page := doc.AddPage(presspdf.A4)

	page.HTML("<h1>Title</h1><p>Paragraph with <b>bold</b> text.</p>")

	fmt.Println("ok")
	// Output: ok
}

func ExamplePage_Markdown() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 12)
	page := doc.AddPage(presspdf.A4)

	page.Markdown("## Heading\n\nParagraph with **bold** and *italic*.")

	fmt.Println("ok")
	// Output: ok
}

func ExampleDocument_BeginTemplate() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 12)

	// Create a reusable template.
	tpl := doc.BeginTemplate(presspdf.A4)
	tpl.TextAt(10, 10, "Letterhead")
	name := doc.EndTemplate()

	page := doc.AddPage(presspdf.A4)
	page.UseTemplate(name, 0, 0, 210, 297)

	fmt.Println("ok")
	// Output: ok
}

func ExampleDocument_AddLayer() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 12)
	page := doc.AddPage(presspdf.A4)

	layerID := doc.AddLayer("Notes", true)
	doc.BeginLayer(layerID)
	page.TextAt(20, 20, "This text is on a toggleable layer")
	doc.EndLayer()

	fmt.Println("ok")
	// Output: ok
}

func ExamplePage_Spacer() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 12)
	page := doc.AddPage(presspdf.A4)

	page.Cell(0, 10, "Before spacer", "", "L", false, 1)
	page.Spacer(20)
	page.Cell(0, 10, "After spacer", "", "L", false, 0)

	fmt.Println("ok")
	// Output: ok
}

func ExamplePage_Ellipse() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 12)
	page := doc.AddPage(presspdf.A4)

	page.Ellipse(100, 100, 40, 25, "D")

	fmt.Println("ok")
	// Output: ok
}

func ExampleDocument_SetAlpha_transparency() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 14)
	page := doc.AddPage(presspdf.A4)

	// Draw overlapping semi-transparent squares.
	doc.SetAlpha(0.5)
	doc.SetFillColor(255, 0, 0)
	page.Rect(20, 20, 60, 60, "F")
	doc.SetFillColor(0, 0, 255)
	page.Rect(50, 50, 60, 60, "F")
	doc.SetAlpha(1.0)

	fmt.Println("ok")
	// Output: ok
}

func ExampleDocument_AddSpotColor() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 12)
	page := doc.AddPage(presspdf.A4)

	doc.AddSpotColor("Pantone 186 C", 0, 100, 81, 4)
	doc.SetFillSpotColor("Pantone 186 C", 100)
	page.Rect(20, 20, 80, 40, "F")

	fmt.Println("ok")
	// Output: ok
}

func ExampleAutoTableFromStructs() {
	type Product struct {
		Name  string  `pdf:"Product"`
		Qty   int     `pdf:"Qty"`
		Price float64 `pdf:"Price"`
	}

	doc := presspdf.New()
	doc.SetFont("helvetica", "", 10)
	page := doc.AddPage(presspdf.A4)

	items := []Product{
		{"Widget A", 100, 5.00},
		{"Widget B", 50, 8.50},
	}
	at := presspdf.AutoTableFromStructs(doc, page, items)
	at.Render()

	fmt.Println("ok")
	// Output: ok
}

func ExampleDocument_SetProtectionAES256() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 12)
	doc.AddPage(presspdf.A4)

	doc.SetProtectionAES256("", "owner123", presspdf.PermAll)

	var buf bytes.Buffer
	if _, err := doc.WriteTo(&buf); err != nil {
		panic(err)
	}
	fmt.Println("AES-256 encrypted:", buf.Len() > 0)
	// Output: AES-256 encrypted: true
}

func ExamplePage_WriteLinkString() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 12)
	page := doc.AddPage(presspdf.A4)

	page.Write(6, "Visit ")
	doc.SetTextColor(0, 0, 255)
	page.WriteLinkString(6, "PressPDF on GitHub", "https://github.com/akkaraponph/presspdf")

	fmt.Println("ok")
	// Output: ok
}

func ExampleDocument_SetTagged() {
	doc := presspdf.New()
	doc.SetTagged(true)
	doc.SetFont("helvetica", "", 12)

	page := doc.AddPage(presspdf.A4)
	page.BeginTag("P")
	page.Write(6, "This paragraph is tagged for accessibility.")
	page.EndTag()

	fmt.Println("ok")
	// Output: ok
}

func ExamplePage_KeepTogether() {
	doc := presspdf.New()
	doc.SetFont("helvetica", "", 12)
	doc.SetAutoPageBreak(true, 15)
	page := doc.AddPage(presspdf.A4)

	// This block will not be split across pages.
	page.KeepTogether(func() {
		page.Cell(0, 10, "Title", "", "L", false, 1)
		page.Cell(0, 10, "Subtitle", "", "L", false, 1)
		page.Cell(0, 10, "Body text", "", "L", false, 1)
	})

	fmt.Println("ok")
	// Output: ok
}

// Package presspdf is a PDF generation library for Go.
//
// PressPDF provides a clean, idiomatic API for creating PDF documents
// programmatically. It handles text, images, shapes, tables, barcodes,
// forms, and more — with built-in support for UTF-8 fonts, Thai text
// segmentation, encryption, digital signatures, tagged PDF, and PDF/A.
//
// # Quick Start
//
//	doc := presspdf.New()
//	doc.SetFont("helvetica", "", 16)
//
//	page := doc.AddPage(presspdf.A4)
//	page.TextAt(20, 20, "Hello, PressPDF!")
//
//	doc.Save("hello.pdf")
//
// # Creating Documents
//
// A [Document] is created with [New] and configured with functional options.
// Pages are added with [Document.AddPage]. The document is written to a file
// or any [io.Writer]:
//
//	doc := presspdf.New(
//	    presspdf.WithUnit(presspdf.UnitMM),
//	    presspdf.WithCompression(true),
//	)
//
//	doc.SetTitle("Invoice #1234")
//	doc.SetAuthor("Acme Corp")
//	doc.SetFont("helvetica", "", 12)
//
//	page := doc.AddPage(presspdf.A4)
//	// ... draw content ...
//
//	doc.Save("invoice.pdf")           // save to file
//	doc.WriteTo(os.Stdout)            // write to io.Writer
//	data, err := doc.Bytes()          // get as []byte
//
// Errors are accumulated inside [Document] — call [Document.Err] at the
// end instead of checking every method call.
//
// Available page sizes: [A4], [A3], [A5], [Letter], [Legal], and their
// landscape variants. Use [PageSize] to define custom sizes.
//
// # Text
//
// Several methods cover different text layout needs:
//
//	page.TextAt(x, y, "Positioned text")              // absolute position
//	page.Cell(w, h, "Cell", "1", "C", true, 0)        // table-like cell
//	page.MultiCell(w, h, text, "1", "J", false)        // auto-wrapping cell
//	page.Write(lineHeight, "Inline flowing text...")    // inline flow
//	page.Paragraph(text,                               // formatted paragraph
//	    presspdf.ParagraphAlign("J"),
//	    presspdf.ParagraphIndent(10),
//	    presspdf.ParagraphSpaceBefore(5),
//	)
//
// Rich inline formatting is available through markup:
//
//	page.RichText(6, "This is <b>bold</b> and <i>italic</i>.")
//	page.HTML("<h1>Title</h1><p>Paragraph with <b>bold</b>.</p>")
//	page.Markdown("## Heading\n\nParagraph with **bold**.")
//
// Text appearance is controlled at the document level:
//
//	doc.SetTextColor(0, 0, 128)          // navy text
//	doc.SetUnderline(true)               // underline
//	doc.SetStrikethrough(true)           // strikethrough
//	doc.SetCharSpacing(0.5)              // character spacing
//	doc.SetWordSpacing(2)                // word spacing
//	doc.SetTextRise(3)                   // superscript/subscript
//	doc.SetTextRenderingMode(1)          // stroke text
//
// A fluent [TextBuilder] API is also available:
//
//	page.Text("Hello").At(20, 30).Bold().Size(18).Draw()
//
// # Fonts
//
// Core PDF fonts (Helvetica, Times, Courier, Symbol, ZapfDingbats) are
// available by name. TrueType fonts with full Unicode support can be
// loaded from files or byte slices:
//
//	doc.SetFont("helvetica", "B", 14)                          // core font
//	doc.AddUTF8FontFromFile("noto", "", "NotoSans-Regular.ttf") // TTF file
//	doc.AddUTF8Font("noto", "", fontBytes)                      // TTF bytes
//	doc.SetFont("noto", "", 12)
//
// Font style string: "" (regular), "B" (bold), "I" (italic), "BI" (bold italic).
//
// Embedded font packages are available for immediate use:
//
//	import "github.com/akkaraponph/presspdf/fonts/sarabun"
//	import "github.com/akkaraponph/presspdf/fonts/dejavu"
//
// # Drawing
//
// Pages support lines, rectangles, circles, ellipses, arcs, polygons,
// and curves. The style parameter controls rendering: "D" (stroke),
// "F" (fill), "DF" (both).
//
//	page.Line(x1, y1, x2, y2)
//	page.Rect(x, y, w, h, "DF")
//	page.Circle(x, y, r, "F")
//	page.Ellipse(x, y, rx, ry, "D")
//	page.EllipseRotated(x, y, rx, ry, 45, "DF")
//	page.Arc(x, y, rx, ry, degStart, degEnd, "D")
//	page.Polygon([]presspdf.Point{...}, "D")
//	page.Curve(x0, y0, cx, cy, x1, y1, "D")
//	page.CurveBezier(x0, y0, cx0, cy0, cx1, cy1, x1, y1, "D")
//
// Arbitrary paths can be built step by step:
//
//	page.MoveTo(10, 10)
//	page.LineTo(50, 10)
//	page.CurveTo(cx0, cy0, cx1, cy1, x, y)
//	page.ClosePath()
//	page.DrawPath("DF")
//
// SVG path data is supported directly:
//
//	page.SVGPath(x, y, scale, "M 10,30 L 50,30 Z", "F")
//
// A fluent [ShapeBuilder] API is also available:
//
//	page.Shape().Rect(20, 20, 80, 40).Fill(200, 220, 255).Draw()
//
// Line appearance is controlled at the document level:
//
//	doc.SetDrawColor(255, 0, 0)          // red stroke
//	doc.SetLineWidth(0.5)
//	doc.SetLineCapStyle("round")         // butt, round, square
//	doc.SetLineJoinStyle("bevel")        // miter, round, bevel
//	doc.SetDashPattern([]float64{3, 2})  // dashed
//
// # Images
//
// JPEG and PNG images are registered once and drawn on any page:
//
//	doc.RegisterImageFromFile("logo", "logo.png")
//	doc.RegisterImage("photo", reader)
//	page.DrawImageRect("logo", x, y, w, h)
//
// # Tables
//
// The simple [Table] API draws rows immediately:
//
//	tbl := presspdf.NewTable(doc, page)
//	tbl.SetWidths(60, 40, 30)
//	tbl.SetHeaderStyle(presspdf.CellStyle{...})
//	tbl.Header("Product", "Qty", "Price")
//	tbl.Row("Widget", "10", "$5.00")
//
// The complex API supports colspan, rowspan, and per-cell styling with
// [TableCell]:
//
//	tbl.AddHeader(presspdf.TableCell{Text: "Report", Colspan: 3})
//	tbl.AddRow(
//	    presspdf.TableCell{Text: "A"},
//	    presspdf.TableCell{Text: "B", Rowspan: 2},
//	)
//	tbl.Render()
//
// [AutoTableFromStructs] and [AutoTableFromJSON] generate tables from Go data:
//
//	type Item struct {
//	    Name  string  `pdf:"Product"`
//	    Price float64 `pdf:"Price"`
//	}
//	at := presspdf.AutoTableFromStructs(doc, page, items)
//	at.Render()
//
// # Layout
//
// Layout helpers manage vertical flow and multi-column arrangements:
//
//	page.Ln(10)                                   // line break
//	page.Spacer(5)                                // vertical space
//	page.PageBreakIfNeeded(50)                    // break if < 50 units left
//	page.KeepTogether(func() { ... })             // prevent splitting
//	page.Stack(block1, block2, block3)            // stack blocks vertically
//
// Multi-column layout:
//
//	cols := presspdf.NewColumnLayout(doc, page, 2, 10)
//	cols.Begin()
//	page.Write(5, "Left column...")
//	cols.NextColumn()
//	page.Write(5, "Right column...")
//	cols.End()
//
// Headers, footers, and automatic page breaks:
//
//	doc.SetHeaderFunc(func(p *presspdf.Page) { ... })
//	doc.SetFooterFunc(func(p *presspdf.Page) { ... })
//	doc.SetFooterFuncLpi(func(p *presspdf.Page, lastPage bool) { ... })
//	doc.SetAutoPageBreak(true, 15)
//	doc.AliasNbPages("{nb}")
//
// # Barcodes and QR Codes
//
// Code 128, EAN-13, and QR codes are built in:
//
//	page.Barcode128(x, y, w, h, "ABC-123")
//	page.Barcode128WithText(x, y, w, h, "ABC-123", 8)
//	page.BarcodeEAN13(x, y, w, h, "4006381333931")
//	page.BarcodeEAN13WithText(x, y, w, h, "4006381333931", 8)
//	page.QRCode(x, y, size, "https://example.com", presspdf.ECMedium)
//
// Error correction levels: [ECLow], [ECMedium], [ECQuartile], [ECHigh].
//
// # Transforms and Effects
//
// Rotation, scaling, skewing, and translation. Always wrap in
// [Page.TransformBegin] / [Page.TransformEnd]:
//
//	page.TransformBegin()
//	page.Rotate(45, cx, cy)
//	page.Scale(1.5, 1.5, cx, cy)
//	page.Skew(10, 0, cx, cy)
//	page.TextAt(cx, cy, "Transformed")
//	page.TransformEnd()
//
// Transparency:
//
//	doc.SetAlpha(0.5)                // 50% opacity
//	// ... draw semi-transparent content ...
//	doc.SetAlpha(1.0)                // restore
//
// Gradients:
//
//	page.LinearGradient(x, y, w, h, x1, y1, x2, y2,
//	    presspdf.GradientStop(0, 255, 0, 0),   // red
//	    presspdf.GradientStop(1, 0, 0, 255),   // blue
//	)
//	page.RadialGradient(x, y, w, h, cx, cy, rx, ry,
//	    presspdf.GradientStop(0, 255, 255, 255),
//	    presspdf.GradientStop(1, 0, 0, 0),
//	)
//
// Clipping restricts drawing to a region:
//
//	page.ClipRect(x, y, w, h)
//	page.ClipCircle(cx, cy, r)
//	page.ClipEllipse(cx, cy, rx, ry)
//	page.ClipText(x, y, "CLIP")
//	// ... draw inside the clip ...
//	page.ClipEnd()
//
// # Navigation
//
// Bookmarks appear in the PDF viewer's outline panel:
//
//	doc.AddBookmark("Chapter 1", 0)     // level 0 = top
//	doc.AddBookmark("Section 1.1", 1)   // level 1 = nested
//
// Table of contents with automatic page numbers:
//
//	toc := presspdf.NewTOC(doc)
//	toc.Add("Introduction", 0, page, page.GetY())
//
// Links — both URL and internal:
//
//	page.LinkURL(x, y, w, h, "https://example.com")
//	page.WriteLinkString(6, "click here", "https://example.com")
//
//	// Anchor-based internal links
//	page.AddAnchor("section1")
//	page.LinkAnchor(x, y, w, h, "section1")
//
//	// Integer-based internal links
//	linkID := doc.AddLink()
//	page1.WriteLinkID(6, "Go to page 2", linkID)
//	doc.SetLink(linkID, 0, 2)
//
// # Interactive Forms
//
// PDF forms with text fields, checkboxes, and dropdowns:
//
//	page.FormTextField("name", x, y, w, h)
//	page.FormTextField("email", x, y, w, h,
//	    presspdf.WithDefaultValue("user@example.com"),
//	    presspdf.WithMaxLen(100),
//	)
//	page.FormCheckbox("agree", x, y, size, false)
//	page.FormDropdown("country", x, y, w, h, []string{"US", "UK", "JP"})
//
// # Templates
//
// Reusable content blocks that can be placed on multiple pages:
//
//	tpl := doc.BeginTemplate(presspdf.A4)
//	tpl.TextAt(10, 10, "Template content")
//	name := doc.EndTemplate()
//
//	page.UseTemplate(name, x, y, w, h)
//
// # Layers
//
// Optional content groups (layers) that viewers can toggle:
//
//	layerID := doc.AddLayer("Annotations", true)
//	doc.BeginLayer(layerID)
//	page.TextAt(20, 20, "This is on a layer")
//	doc.EndLayer()
//	doc.OpenLayerPane()
//
// # Security
//
// Password protection with RC4 or AES-256 encryption:
//
//	doc.SetProtection("user", "owner", presspdf.PermPrint|presspdf.PermCopy)
//	doc.SetProtectionAES256("user", "owner", presspdf.PermAll)
//
// Permission flags: [PermPrint], [PermModify], [PermCopy], [PermAnnotate], [PermAll].
//
// Digital signatures with X.509 certificates:
//
//	doc.Sign(cert, key, page, x, y, w, h, presspdf.SignOptions{
//	    Name:   "Signer",
//	    Reason: "Approval",
//	})
//
// # Accessibility and Compliance
//
// Tagged PDF for screen readers:
//
//	doc.SetTagged(true)
//	page.BeginTag("P")
//	page.Write(6, "Accessible paragraph text")
//	page.EndTag()
//
// PDF/A archival compliance:
//
//	doc := presspdf.New(presspdf.WithPDFA("1b"))
//
// # PDF Utilities
//
// Standalone functions for working with existing PDF files:
//
//	presspdf.MergePDF("out.pdf", "a.pdf", "b.pdf")
//	presspdf.SplitPDF("in.pdf", "outdir/")
//	presspdf.CompressPDF("in.pdf", "out.pdf")
//	presspdf.WatermarkPDF("in.pdf", "out.pdf",
//	    presspdf.WatermarkText("DRAFT"),
//	    presspdf.WatermarkOpacity(0.3),
//	    presspdf.WatermarkRotation(45),
//	)
//	presspdf.DecryptPDF("in.pdf", "out.pdf", "password")
//
// Text extraction:
//
//	pages, _ := presspdf.ExtractText("in.pdf")
//	pages, _ := presspdf.ExtractTextFromBytes(data)
//
// Conversion:
//
//	presspdf.ConvertToImages("in.pdf", "outdir/")
//	presspdf.ConvertPage("in.pdf", 1)
//	presspdf.ImagesToPDF("out.pdf", []string{"1.png", "2.jpg"})
//
// # Spot Colors
//
// CMYK spot colors for professional printing:
//
//	doc.AddSpotColor("Pantone 186 C", 0, 100, 81, 4)
//	doc.SetFillSpotColor("Pantone 186 C", 100)
//	doc.SetDrawSpotColor("Pantone 186 C", 50)
//	doc.SetTextSpotColor("Pantone 186 C", 100)
//
// # Coordinate System
//
// All coordinates use a top-left origin. Units are configurable:
//
//	presspdf.New(presspdf.WithUnit(presspdf.UnitMM))   // millimeters (default)
//	presspdf.New(presspdf.WithUnit(presspdf.UnitPt))   // points (1/72 inch)
//	presspdf.New(presspdf.WithUnit(presspdf.UnitInch)) // inches
//
// Colors are specified as 0–255 RGB integers:
//
//	doc.SetDrawColor(255, 0, 0)   // red stroke
//	doc.SetFillColor(0, 0, 255)   // blue fill
//	doc.SetTextColor(0, 128, 0)   // green text
//
// # Thai Language Support
//
// The [github.com/akkaraponph/presspdf/thai] subpackage provides Thai word
// segmentation and font setup:
//
//	import "github.com/akkaraponph/presspdf/thai"
//
//	thai.Setup(doc) // registers Sarabun font + word breaker
package presspdf

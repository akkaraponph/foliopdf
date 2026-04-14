<p align="center">
  <img src="docs/assets/banner-presspdf.png" alt="PressPDF" width="600">
</p>

<p align="center">
  A full-featured PDF generation library for Go.
</p>

<p align="center">
  <a href="#install">Install</a> &middot;
  <a href="#quick-start">Quick Start</a> &middot;
  <a href="#features">Features</a> &middot;
  <a href="#documentation">Documentation</a> &middot;
  <a href="#examples">Examples</a>
</p>

---

## Why PressPDF?

- **Clean 4-layer architecture** — easy to understand, easy to extend
- **Full Unicode** — TrueType font embedding with CIDFont Type 2
- **Thai language built-in** — dictionary-based word segmentation (~15K words)
- **Rich feature set** — tables, barcodes, forms, signatures, HTML/Markdown, PDF/A
- **PDF tools included** — merge, split, compress, watermark, decrypt, and convert existing PDFs

## Install

```bash
go get github.com/akkaraponph/presspdf
```

Requires Go 1.26+.

## Quick Start

```go
package main

import "github.com/akkaraponph/presspdf"

func main() {
    doc := presspdf.New()
    doc.SetFont("helvetica", "B", 20)

    page := doc.AddPage(presspdf.A4)
    page.TextAt(40, 30, "Hello, PressPDF!")

    doc.Save("hello.pdf")
}
```

## Features

### Text & Typography

```go
// Core fonts (built-in, no files needed)
doc.SetFont("helvetica", "B", 16)
page.TextAt(20, 20, "Bold Helvetica")

// TrueType fonts with full Unicode
doc.AddUTF8Font("sarabun", "", sarabunData)
doc.SetFont("sarabun", "", 14)
page.TextAt(20, 40, "สวัสดีครับ")

// Text decoration
doc.SetUnderline(true)
doc.SetStrikethrough(true)

// Fluent builder API
page.Text("Styled text").At(20, 60).Bold().Color(255, 0, 0).Size(18).Draw()
```

Core fonts: Helvetica, Times, Courier, Arial, ZapfDingbats — each with regular, bold, italic, and bold-italic variants.

### Drawing Primitives

```go
// Lines, rectangles, circles, ellipses, arcs
page.Line(20, 50, 190, 50)
page.Rect(20, 60, 170, 40, "DF") // "D"=stroke, "F"=fill, "DF"=both
page.Circle(105, 150, 30, "D")
page.Ellipse(105, 150, 40, 25, "DF")

// SVG paths
page.SVGPath(10, 10, 0.5, "M 0 0 L 100 0 L 100 100 Z", "D")

// Dash patterns
page.SetDashPattern([]float64{3, 2}, 0)
```

### Colors & Transparency

```go
doc.SetDrawColor(0, 0, 255)       // stroke color (RGB 0-255)
doc.SetFillColor(200, 220, 255)   // fill color
doc.SetTextColor(100, 0, 0)       // text color
doc.SetAlpha(0.5)                 // 50% opacity
```

### Gradients

```go
// Linear gradient
page.LinearGradient(20, 20, 170, 50, 20, 20, 190, 20,
    presspdf.ColorStop(0, 255, 0, 0),
    presspdf.ColorStop(1, 0, 0, 255),
)

// Radial gradient
page.RadialGradient(20, 80, 170, 50, 105, 105, 60,
    presspdf.ColorStop(0, 255, 255, 0),
    presspdf.ColorStop(1, 0, 128, 0),
)
```

### Transforms

```go
page.TransformBegin()
page.Rotate(45, 105, 148)    // rotate 45° around point
page.Scale(1.5, 1.5, 50, 50) // scale 150%
page.Skew(10, 0, 50, 50)     // horizontal shear
page.Translate(20, 0)         // shift right
page.TransformEnd()

// Convenience: rotated text
page.TextRotatedAt(100, 200, 30, "Angled text")
```

### Clipping

```go
page.TransformBegin()
page.ClipRect(20, 20, 100, 80)
page.ClipCircle(70, 60, 40)
// ... draw inside the clipped region ...
page.TransformEnd()
```

### Text Layout

```go
// Positioned cursor
page.SetXY(20, 100)

// Single-line cell with optional border and fill
page.Cell(60, 10, "Name", "1", "L", true)

// Multi-line with automatic word wrapping
page.MultiCell(170, 6, longText, "1", "J", false)

// Inline text flow (like Write in fpdf)
page.Write(6, "This flows ")
page.Write(6, "continuously.")
```

### Tables

**Simple API** — draws rows immediately:

```go
tbl := presspdf.NewTable(doc, page, 20, []float64{20, 80, 40}, 8)
tbl.SetHeaderStyle(presspdf.CellStyle{FontStyle: "B", Fill: true, FillColor: [3]int{200, 220, 255}})
tbl.Header("#", "Product", "Price")
tbl.Row("1", "Widget", "$9.99")
tbl.Row("2", "Gadget", "$19.99")
```

**Complex API** — buffered rendering with colspan, rowspan, multi-line cells:

```go
tbl := presspdf.NewTable(doc, page, 20, []float64{30, 60, 50, 50}, 8)
tbl.AddHeader(presspdf.TableCell{Text: "Quarterly Report", ColSpan: 4, Align: "C"})
tbl.AddRow(
    presspdf.TableCell{Text: "Region", RowSpan: 2},
    presspdf.TableCell{Text: "Q1"},
    presspdf.TableCell{Text: "Q2"},
    presspdf.TableCell{Text: "Total"},
)
tbl.Render()
```

**Auto-tables** — generate from structs or JSON:

```go
type Product struct {
    Name  string
    Price float64
}
products := []Product{{"Widget", 9.99}, {"Gadget", 19.99}}
at := presspdf.AutoTableFromStructs(doc, page, products)
at.Render()
```

### Images

```go
doc.RegisterImage("photo", jpegBytes, 800, 600)
page.DrawImageRect("photo", 20, 20, 80, 60)
```

### Barcodes

```go
page.Barcode128(20, 200, 100, 30, "PRESSPDF-2024")
page.BarcodeEAN13(20, 240, 80, 30, "590123412345")
page.QRCode(140, 200, 40, "https://example.com")
```

### HTML & Markdown

```go
// Render HTML subset
page.HTML(`
    <h1>Title</h1>
    <p>Paragraph with <b>bold</b> and <i>italic</i>.</p>
    <ul><li>Item 1</li><li>Item 2</li></ul>
    <table><tr><td>A</td><td>B</td></tr></table>
`)

// Render Markdown subset
page.Markdown(`
# Heading
**Bold** and *italic* text.
- List item
[Link](https://example.com)
`, presspdf.WithBookmarks())
```

### Page Layout

```go
// Page sizes and orientation
page := doc.AddPage(presspdf.A4)
page := doc.AddPage(presspdf.A4.Landscape())
page := doc.AddPage(presspdf.Letter)

// Margins
doc.SetMargins(15, 15, 15)
doc.SetAutoPageBreak(true, 20)

// Headers and footers
doc.SetHeaderFunc(func(p *presspdf.Page) {
    p.TextAt(20, 10, "My Document")
    p.Line(20, 14, 190, 14)
})
doc.SetFooterFunc(func(p *presspdf.Page) {
    p.SetY(-15)
    p.Cell(0, 10, fmt.Sprintf("Page %d", doc.PageNo()), "", "C", false)
})

// Multi-column layout
cols := presspdf.NewColumnLayout(doc, page, 2, 5)
cols.Begin()
page.MultiCell(0, 5, leftText, "", "L", false)
cols.NextColumn()
page.MultiCell(0, 5, rightText, "", "L", false)
cols.End()
```

### Layout Helpers

```go
page.Spacer(10)              // vertical gap (triggers page break if needed)
page.PageBreakIfNeeded(50)   // ensure 50mm of space, or break

// Keep content together on one page
page.KeepTogether(func() {
    page.Cell(0, 8, "Title", "", "L", false)
    page.MultiCell(0, 5, body, "", "L", false)
})
```

### Bookmarks & Table of Contents

```go
// Bookmarks (PDF outline sidebar)
doc.AddBookmark("Chapter 1", 0, page, 20)
doc.AddBookmark("Section 1.1", 1, page, 60)

// Table of contents with dot leaders and page numbers
toc := presspdf.NewTOC(doc)
toc.Add("Chapter 1", 0, chapterPage, 20)
toc.Add("Section 1.1", 1, sectionPage, 60)
toc.Render(tocPage, 6)
```

### Links

```go
page.LinkURL(20, 100, 80, 10, "https://example.com")
doc.AddAnchor("chapter1", page, 20)
page.LinkAnchor(20, 120, 80, 10, "chapter1")
```

### Interactive Forms (AcroForms)

```go
page.FormTextField("name", 60, 50, 100, 12)
page.FormCheckbox("agree", 60, 70, 10, false)
page.FormDropdown("color", 60, 90, 100, 12, []string{"Red", "Green", "Blue"})
```

### Page Templates

```go
tpl := doc.BeginTemplate(presspdf.A4)
tpl.TextAt(20, 10, "CONFIDENTIAL")
tpl.Line(20, 14, 190, 14)
doc.EndTemplate(tpl)

page.UseTemplate(tpl, 0, 0, 1.0) // stamp onto any page
```

### Digital Signatures

```go
doc.Sign(cert, privateKey, page, 20, 250, 80, 30, presspdf.SignOptions{
    Name:     "John Doe",
    Reason:   "Approval",
    Location: "Bangkok",
})
```

### Encryption

```go
doc.SetProtection("userpass", "ownerpass", presspdf.PermPrint|presspdf.PermCopy)
```

### PDF/A Compliance

```go
doc := presspdf.New(presspdf.WithPDFA("1b")) // PDF/A-1b
doc := presspdf.New(presspdf.WithPDFA("2b")) // PDF/A-2b
```

### PDF Tools

Work with any existing PDF — not just files created by PressPDF.

```go
// Merge multiple PDFs
presspdf.MergePDF("combined.pdf", "doc1.pdf", "doc2.pdf", "doc3.pdf")

// Split into individual pages
presspdf.SplitPDF("report.pdf", "output/")

// Split by page ranges
presspdf.SplitPDF("report.pdf", "output/",
    presspdf.WithRanges(presspdf.PageRange{From: 1, To: 5}),
)

// Compress an existing PDF
presspdf.CompressPDF("input.pdf", "smaller.pdf")

// Add watermark
presspdf.WatermarkPDF("input.pdf", "output.pdf",
    presspdf.WatermarkText("DRAFT"),
    presspdf.WatermarkOpacity(0.3),
    presspdf.WatermarkRotation(45),
)

// Remove password protection
presspdf.DecryptPDF("encrypted.pdf", "decrypted.pdf", "password")

// Convert images to PDF
presspdf.ImagesToPDF("album.pdf", "photo1.jpg", "photo2.jpg")

// Convert PDF pages to images
presspdf.ConvertPDF("document.pdf", "output/", presspdf.WithFormat(presspdf.PNG))
```

### Thai Language Support

```go
import (
    "github.com/akkaraponph/presspdf/fonts/sarabun"
    "github.com/akkaraponph/presspdf/thai"
)

doc.AddUTF8Font("sarabun", "", sarabun.Regular())
doc.AddUTF8Font("sarabun", "B", sarabun.Bold())
doc.SetFont("sarabun", "", 14)
doc.SetWordBreaker(thai.WordBreaker)

page.MultiCell(170, 6, "ภาษาไทยตัดคำอัตโนมัติ", "", "L", false)
```

### Measurement Units

```go
presspdf.New(presspdf.WithUnit(presspdf.UnitMM))   // millimeters (default)
presspdf.New(presspdf.WithUnit(presspdf.UnitPt))   // points (1/72 inch)
presspdf.New(presspdf.WithUnit(presspdf.UnitCM))   // centimeters
presspdf.New(presspdf.WithUnit(presspdf.UnitInch)) // inches
```

### Output

```go
doc.Save("output.pdf")              // write to file
doc.WriteTo(w)                       // write to any io.Writer
buf, err := doc.Bytes()              // get as byte slice
```

## Documentation

| Guide | What you'll learn |
|-------|------------------|
| [Getting Started](docs/getting-started.md) | Your first PDF in 5 minutes |
| [Text & Fonts](docs/text-and-fonts.md) | Core fonts, TrueType, Thai, typography |
| [Layout](docs/layout.md) | Pages, margins, page breaks, headers, columns |
| [Drawing](docs/drawing.md) | Shapes, SVG paths, transforms, gradients |
| [Tables](docs/tables.md) | Styled tables, auto-tables from structs/JSON |
| [Rich Content](docs/rich-content.md) | HTML, Markdown, fluent builder API |
| [Images & Barcodes](docs/images-and-barcodes.md) | JPEG images, Code 128, EAN-13, QR codes |
| [Security & Compliance](docs/security.md) | Encryption, signatures, forms, PDF/A |
| [PDF Tools](docs/tools.md) | Merge, split, compress, watermark, decrypt, convert |
| [Architecture](docs/architecture.md) | 4-layer internals, how to extend |
| [API Reference](docs/api.md) | Complete function reference |

## Examples

```bash
go run ./cmd/demo              # minimal getting-started
go run ./examples/thai         # Thai language
go run ./examples/invoice      # bilingual business invoice
go run ./examples/tables       # table features showcase
go run ./examples/article      # long-form article with TOC
go run ./examples/resume       # resume/CV layout
go run ./examples/ratesheet    # financial rate sheet
go run ./examples/showcase     # 15 feature demos
```

**PDF tools:**

```bash
go run ./examples/mergepdf     # merge multiple PDFs
go run ./examples/splitpdf     # split PDF by pages
go run ./examples/compresspdf  # compress existing PDF
go run ./examples/watermark    # add watermark overlay
go run ./examples/decryptpdf   # remove password protection
go run ./examples/jpg2pdf      # convert images to PDF
go run ./examples/pdf2jpg      # convert PDF to images
```

See [`examples/`](examples/) for all runnable examples.

## Architecture

PressPDF uses a clean 4-layer architecture where each layer has a single responsibility and depends only on layers below it:

```
┌─────────────────────────────────────────────┐
│  presspdf (public API)                      │
├─────────────────────────────────────────────┤
│  internal/state      │ Unit conversion,     │
│                      │ color math           │
├──────────────────────┼──────────────────────┤
│  internal/resources  │ Font/image registry, │
│                      │ deduplication        │
├──────────────────────┼──────────────────────┤
│  internal/content    │ PDF content stream   │
│                      │ operators            │
├──────────────────────┼──────────────────────┤
│  internal/pdfcore    │ Raw PDF syntax,      │
│                      │ objects, xref        │
└─────────────────────────────────────────────┘
```

See [Architecture](docs/architecture.md) for details on internals and extending PressPDF.

## License

MIT

<p align="center">
  <img src="assets/logo-presspdf.png" alt="PressPDF" width="120">
</p>

# Text & Fonts

## Core fonts

PressPDF includes metrics for 13 standard PDF fonts. These are built into every PDF viewer — no embedding needed.

| Family | Styles | Notes |
|--------|--------|-------|
| `helvetica` | Regular, B, I, BI | Sans-serif. `arial` is an alias. |
| `times` | Regular, B, I, BI | Serif. |
| `courier` | Regular, B, I, BI | Monospace. |
| `zapfdingbats` | Regular | Symbols. `symbol` is an alias. |

```go
doc.SetFont("helvetica", "", 12)    // regular
doc.SetFont("helvetica", "B", 14)   // bold
doc.SetFont("times", "I", 11)       // italic
doc.SetFont("courier", "BI", 10)    // bold italic
```

Core fonts only support Latin characters (WinAnsiEncoding). For Unicode text, use TrueType fonts.

## TrueType fonts

Load any `.ttf` file to get full Unicode support:

```go
// From file
doc.AddUTF8FontFromFile("noto", "", "NotoSans-Regular.ttf")
doc.AddUTF8FontFromFile("noto", "B", "NotoSans-Bold.ttf")

// From embedded bytes
doc.AddUTF8Font("myfont", "", fontBytes)
```

TrueType fonts are embedded in the PDF as CIDFont Type 2 with 2-byte character encoding, supporting any Unicode text.

## Bundled fonts

PressPDF ships two font families ready to use:

### Sarabun (Thai + Latin)

```go
import "github.com/akkaraponph/presspdf/fonts/sarabun"

sarabun.Register(doc) // registers Regular, Bold, Italic, BoldItalic
doc.SetFont("sarabun", "B", 14)
```

### DejaVu Sans Condensed (Latin + extended)

```go
import "github.com/akkaraponph/presspdf/fonts/dejavu"

dejavu.Register(doc)
doc.SetFont("dejavu", "", 12)
```

## Thai language support

Thai text needs word segmentation for proper line wrapping (Thai doesn't use spaces between words). PressPDF includes a built-in dictionary-based segmenter:

```go
import (
    "github.com/akkaraponph/presspdf/fonts/sarabun"
    "github.com/akkaraponph/presspdf/thai"
)

doc := presspdf.New()
sarabun.Register(doc)
thai.Setup(doc)  // install word breaker

doc.SetFont("sarabun", "", 14)
page := doc.AddPage(presspdf.A4)
page.SetXY(20, 20)
page.MultiCell(170, 7,
    "สวัสดีครับ นี่คือตัวอย่างภาษาไทยใน PressPDF",
    "", "L", false)
```

Without `thai.Setup()`, Thai text wraps only at spaces — which means it won't wrap within Thai phrases. The segmenter uses a ~15K word dictionary and a shortest-path algorithm.

You can also use the segmenter directly:

```go
tokens := thai.Segment("สวัสดีครับ")
// ["สวัสดี", "ครับ"]
```

### Custom word breakers

For other languages without spaces (Chinese, Japanese, Khmer, etc.), supply your own:

```go
doc.SetWordBreaker(func(paragraph string) []string {
    return mySegmenter.Split(paragraph)
})
```

## Typography settings

### Character and word spacing

```go
doc.SetCharSpacing(0.5)   // extra space between each character (PDF Tc operator)
doc.SetWordSpacing(2.0)   // extra space at each space character (PDF Tw operator)
```

### Text rise

Shift the baseline up or down. Useful for superscript/subscript effects:

```go
doc.SetFont("helvetica", "", 12)
page.Write(6, "H")
doc.SetTextRise(-2)
doc.SetFont("helvetica", "", 8)
page.Write(6, "2")
doc.SetTextRise(0)
doc.SetFont("helvetica", "", 12)
page.Write(6, "O")
```

### Underline and strikethrough

```go
doc.SetUnderline(true)
page.TextAt(20, 30, "This text is underlined")
doc.SetUnderline(false)

doc.SetStrikethrough(true)
page.TextAt(20, 40, "This text has a line through it")
doc.SetStrikethrough(false)
```

### Configurable underline thickness

Control the underline weight with a multiplier (default 1.0):

```go
doc.SetUnderlineThickness(2.0) // double thickness
doc.SetUnderline(true)
page.Cell(80, 10, "Thick underline", "", "L", false, 0)
doc.SetUnderlineThickness(0.5) // half thickness
```

### Text rendering mode

Control how text glyphs are painted — fill, stroke, or both:

```go
doc.SetTextRenderingMode(1) // 0=fill, 1=stroke, 2=fill+stroke, 3=invisible
page.TextAt(20, 50, "Outlined text")
doc.SetTextRenderingMode(0) // restore fill
```

Modes 4-7 add clipping behavior (fill+clip, stroke+clip, fill+stroke+clip, clip only).

### Measuring text

Get the width of a string in current units, using the current font:

```go
w := page.GetStringWidth("Hello, World!")
// Use to center text, align elements, etc.
```

### Splitting text for layout

Split text into lines that fit within a given width, useful for calculating vertical space:

```go
lines := page.SplitText("Long paragraph text...", 170)
totalHeight := float64(len(lines)) * lineHeight
```

## Line feed

`Ln` moves the cursor to the left margin and advances vertically:

```go
page.Cell(80, 10, "Line 1", "", "L", false, 0)
page.Ln(10)   // advance 10 user units
page.Cell(80, 10, "Line 2", "", "L", false, 0)
page.Ln(-1)   // advance by the height of the last Cell
```

## Printf-style formatting

Convenience wrappers with `fmt.Sprintf` built in:

```go
page.Cellf(80, 10, "Total: $%.2f", 99.95)
page.Writef(6, "Item %d: %s", 1, "Widget")
```

## Inline hyperlinks

Write text that opens a URL when clicked:

```go
page.WriteLinkString(6, "Visit our website", "https://example.com")
```

## Aligned multi-line text

`WriteAligned` writes word-wrapped text with horizontal alignment:

```go
page.WriteAligned(0, 6, "Centered paragraph", "C")  // 0 = full page width
page.Ln(6)
page.WriteAligned(100, 6, "Right-aligned text", "R") // 100mm wide box
```

Alignment options: `"L"` (left), `"C"` (center), `"R"` (right).

## Superscript and subscript

`SubWrite` shifts text vertically with a different font size:

```go
// Superscript: E = mc²
page.Write(6, "E = mc")
page.SubWrite(6, "2", 8, 4)  // fontSize=8pt, offset=+4pt (up)

// Subscript: H₂O
page.Write(6, "H")
page.SubWrite(6, "2", 8, -3) // fontSize=8pt, offset=-3pt (down)
page.Write(6, "O")
```

## Rotated text

```go
page.TextRotatedAt(105, 148, 45, "Rotated 45 degrees")
```

The text rotates around the given point.

## Rich inline text

Mix formatting inline without switching fonts manually:

```go
page.RichText(6, "Normal <b>bold</b> and <i>italic</i> text")
```

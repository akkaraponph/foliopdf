# Folio Implementation Guide

## Architecture

Folio is a layered PDF generation library. Each layer has a single responsibility and depends only on layers below it.

```
┌──────────────────────────────────────────────┐
│  folio (public API)                          │
│  Document, Page, options, serialization      │
├──────────────────────────────────────────────┤
│  internal/state      │  internal/resources   │
│  units, colors,      │  fonts, images,       │
│  coordinate xform    │  deduplication        │
├──────────────────────┴───────────────────────┤
│  internal/content                            │
│  PDF content stream operators (BT, m, re...) │
├──────────────────────────────────────────────┤
│  internal/pdfcore                            │
│  raw PDF objects, xref, trailer, streams     │
└──────────────────────────────────────────────┘
```

### Layer rules

- **pdfcore** knows nothing about fonts, images, pages, or coordinates. It writes raw PDF syntax.
- **content** knows nothing about object IDs or document structure. It outputs PDF operators into a buffer.
- **resources** knows nothing about serialization order. It stores font/image data and handles dedup.
- **state** knows nothing about PDF syntax. It converts user units to points and manages colors.
- **folio** (root) orchestrates all layers. It is the only package users import.

---

## File Map

```
folio/
├── doc.go              # package documentation
├── options.go          # PageSize constants, Option funcs
├── document.go         # Document struct and methods
├── page.go             # Page struct: TextAt, Cell, MultiCell, Line, Rect, DrawImageRect
├── serialize.go        # PDF serialization pipeline (putPages, putFonts, putImages, etc.)
├── escape.go           # PDF string escaping
├── folio_test.go       # integration tests
│
├── internal/
│   ├── pdfcore/
│   │   └── writer.go       # Writer: NewObj, EndObj, PutStream, WriteXref, WriteTrailer
│   ├── content/
│   │   └── stream.go       # Stream: BeginText, SetFont, MoveTo, Rect, Stroke, DrawImage
│   ├── resources/
│   │   ├── fonts.go         # FontRegistry, FontEntry, StringWidth
│   │   ├── fonts_core.go    # 13 core PDF font width tables
│   │   └── images.go        # ImageRegistry, ImageEntry (JPEG, SHA-1 dedup)
│   └── state/
│       ├── state.go         # Color, ColorFromRGB
│       └── units.go         # Unit enum, ScaleFactor, ToPointsX, ToPointsY
│
└── cmd/demo/
    └── main.go          # runnable demo
```

---

## How PDF Generation Works

### 1. User builds the document model

```go
doc := folio.New()
doc.SetFont("helvetica", "", 16)
page := doc.AddPage(folio.A4)
page.TextAt(40, 60, "Hello")
```

Each drawing call (`TextAt`, `Line`, `Rect`, `Cell`, `DrawImageRect`) appends PDF operators to the page's `content.Stream` buffer. Coordinates are converted from user units (mm, top-left origin) to PDF points (bottom-left origin) at call time.

### 2. Serialization pipeline

When `WriteTo` or `Save` is called, `serialize()` in `serialize.go` runs:

```
WriteHeader("1.4")
  │
  ├─ putPages()         writes Page dict + content stream for each page
  │                     then writes Pages root at object 1
  │
  ├─ putFonts()         writes /Type /Font for each registered core font
  │
  ├─ putImages()        writes /Type /XObject /Subtype /Image for each JPEG
  │
  ├─ putResourceDict()  writes shared resource dict at object 2
  │                     references all fonts (/F1, /F2...) and images (/Im1...)
  │
  ├─ putInfo()          writes metadata (title, author, creation date)
  │
  └─ putCatalog()       writes /Type /Catalog pointing to Pages root
      │
      ├─ WriteXref()     cross-reference table with byte offsets
      ├─ WriteTrailer()  /Root and /Info references
      └─ WriteStartXref() + %%EOF
```

### 3. Object numbering

Objects 1 and 2 are reserved:
- **Object 1**: Pages root (`/Type /Pages`, `/Kids [...]`, `/Count N`)
- **Object 2**: Shared resource dictionary (`/Font <<...>>`, `/XObject <<...>>`)

These are written out of order using `Writer.SetOffset()`. All other objects get sequential IDs via `Writer.NewObj()`.

Example for 1 page, 1 font, 1 image:

| Obj | Content | Written by |
|-----|---------|------------|
| 1 | Pages root | putPages (deferred) |
| 2 | Resource dict | putResourceDict (deferred) |
| 3 | Page dict | putPages |
| 4 | Content stream | putPages |
| 5 | Font: Helvetica | putFonts |
| 6 | Image XObject | putImages |
| 7 | Info dict | putInfo |
| 8 | Catalog | putCatalog |

### 4. Coordinate system

Users work in **top-left origin** with configurable units (default: mm).

Internally, PDF uses **bottom-left origin** in points (1/72 inch).

Conversion (`state/units.go`):
```
x_pdf = x_user * k
y_pdf = (pageHeight_user - y_user) * k
```

where `k = ScaleFactor(unit)` (for mm: `72/25.4 ≈ 2.8346`).

---

## How to Add a New Feature

### Adding a new drawing primitive

1. **Add the PDF operator to `internal/content/stream.go`**

   Example: adding a circle requires cubic bezier approximation:
   ```go
   func (s *Stream) Circle(cx, cy, r float64) {
       k := 0.5522847498 // bezier approximation of quarter-circle
       s.MoveTo(cx+r, cy)
       s.CurveTo(cx+r, cy+r*k, cx+r*k, cy+r, cx, cy+r)
       s.CurveTo(cx-r*k, cy+r, cx-r, cy+r*k, cx-r, cy)
       s.CurveTo(cx-r, cy-r*k, cx-r*k, cy-r, cx, cy-r)
       s.CurveTo(cx+r*k, cy-r, cx+r, cy-r*k, cx+r, cy)
   }
   ```

2. **Add the public method to `page.go`**

   Convert user coordinates to points, call the content operator:
   ```go
   func (p *Page) Circle(cx, cy, r float64, style string) {
       k := p.doc.k
       p.stream.Circle(
           state.ToPointsX(cx, k),
           state.ToPointsY(cy, p.h, k),
           r*k,
       )
       // stroke/fill based on style
   }
   ```

3. **Add a test to `folio_test.go`**

4. **No changes needed** in pdfcore, resources, state, or serialize.go.

### Adding a new resource type (e.g. PNG images)

1. **Add parsing in `internal/resources/images.go`**

   Create `RegisterPNG(name, io.Reader)` that:
   - Reads raw bytes
   - Parses PNG header for dimensions, color type, bit depth
   - Decompresses IDAT chunks (PNG uses zlib internally)
   - Extracts alpha channel as separate soft mask data
   - Sets `Filter: "FlateDecode"` (re-compress the raw pixel data)
   - Computes SHA-1 for dedup

2. **Update `putImages` in `serialize.go`**

   Handle the new filter and add soft mask object:
   ```go
   if ie.Filter == "FlateDecode" {
       // Write compressed pixel data
       // If SMask data exists, write a separate SMask XObject
   }
   ```

3. **Update `putResourceDict`** if needed (usually no change).

4. **Update `RegisterImage` in `document.go`**

   Detect image format and route to `RegisterJPEG` or `RegisterPNG`.

### Adding TTF font support

1. **Create `internal/resources/ttf.go`**

   Parse TrueType file:
   - Read `cmap` table for Unicode-to-glyph mapping
   - Read `hmtx` table for glyph widths
   - Read `head` table for units-per-em
   - Read `OS/2` table for ascent/descent
   - Scale widths: `width_1000 = glyphWidth * 1000 / unitsPerEm`

2. **Extend `FontEntry`**

   Add fields for TTF data:
   ```go
   type FontEntry struct {
       // ... existing fields ...
       TTFData    []byte   // raw TTF file bytes (for embedding)
       CIDWidths  map[rune]int  // Unicode → width
       ToUnicode  []byte   // CMap for text extraction
   }
   ```

3. **Add `RegisterTTF` to `FontRegistry`**

4. **Update `putFonts` in `serialize.go`**

   TTF fonts require more objects than core fonts:
   - Font dictionary (`/Subtype /TrueType` or `/Type0` for CID)
   - Font descriptor (`/FontDescriptor`)
   - Embedded font stream (the TTF file, possibly subset)
   - CIDSystemInfo, W array (widths), ToUnicode CMap

5. **Update text encoding**

   Core fonts use WinAnsiEncoding (single byte). TTF/UTF-8 requires:
   - Encoding text as CID values (2-byte glyph IDs)
   - Writing hex strings `<XXXX>` instead of literal strings `(text)`

### Adding auto page breaks

1. **Add to `Document`**:
   ```go
   autoPageBreak    bool
   pageBreakTrigger float64 // y position (in user units) where break triggers
   ```

2. **Add check in `Cell` and `MultiCell`**:
   ```go
   if d.autoPageBreak && p.y+h > d.pageBreakTrigger {
       d.AddPage(p.size)
   }
   ```

3. **Add `SetAutoPageBreak(enabled bool, margin float64)` to Document**.

### Adding headers and footers

1. **Add callback fields to `Document`**:
   ```go
   headerFunc func(*Page)
   footerFunc func(*Page)
   ```

2. **Call `headerFunc` at the start of `AddPage`** (after creating the page).

3. **Call `footerFunc` at the end of a page** (before creating next page, or during serialize).

4. **Track `inHeader`/`inFooter` flags** to prevent recursive page breaks.

---

## How the Serialization Pipeline Maps to PDF Structure

A PDF file has this physical structure:

```
%PDF-1.4                          ← header
3 0 obj ... endobj                ← body objects
4 0 obj ... stream ... endstream endobj
1 0 obj ... endobj                ← (Pages root, written last)
5 0 obj ... endobj
2 0 obj ... endobj                ← (Resource dict, written last)
xref                              ← cross-reference table
0 9
0000000000 65535 f
0000000XXX 00000 n                ← byte offset of each object
...
trailer                           ← trailer
<< /Size 9 /Root 8 0 R /Info 7 0 R >>
startxref
XXXX                              ← byte offset of xref
%%EOF
```

The `pdfcore.Writer` handles this automatically:
- `NewObj()` records the byte offset in `offsets[objNum]`
- `SetOffset(1)` / `SetOffset(2)` record offsets for reserved objects
- `WriteXref()` iterates `offsets[1..n]` and formats the table
- `WriteTrailer()` references Root (Catalog) and Info objects

---

## Error Handling

Folio uses **error accumulation** (like gofpdf):

```go
doc := folio.New()
doc.SetFont("nonexistent", "", 12)  // sets internal error
page := doc.AddPage(folio.A4)       // returns dummy page
page.TextAt(10, 10, "test")         // silently no-ops

_, err := doc.WriteTo(w)            // returns the accumulated error
// or check at any point:
if doc.Err() != nil { ... }
```

Every method checks `d.err != nil` at the top and returns early. This lets users write imperative code without error checks on every line.

---

## Content Stream Operators Reference

These are the PDF operators emitted by `internal/content/stream.go`:

| Operator | Method | Effect |
|----------|--------|--------|
| `q` | `SaveState()` | Push graphics state |
| `Q` | `RestoreState()` | Pop graphics state |
| `w` | `SetLineWidth(w)` | Set line width |
| `J` | `SetLineCap(s)` | Set line cap style |
| `j` | `SetLineJoin(s)` | Set line join style |
| `RG` | `SetStrokeColorRGB(r,g,b)` | Set stroke color |
| `rg` | `SetFillColorRGB(r,g,b)` | Set fill color |
| `G` | `SetStrokeGray(g)` | Set stroke gray |
| `g` | `SetFillGray(g)` | Set fill gray |
| `BT` | `BeginText()` | Begin text object |
| `ET` | `EndText()` | End text object |
| `Tf` | `SetFont(name, size)` | Set font and size |
| `Td` | `MoveText(tx, ty)` | Move text position |
| `Tj` | `ShowText(s)` | Show text string |
| `TL` | `SetTextLeading(v)` | Set text leading |
| `T*` | `NextLine()` | Move to next line |
| `m` | `MoveTo(x, y)` | Begin subpath |
| `l` | `LineTo(x, y)` | Line segment |
| `c` | `CurveTo(...)` | Cubic bezier curve |
| `re` | `Rect(x, y, w, h)` | Rectangle |
| `h` | `ClosePath()` | Close subpath |
| `S` | `Stroke()` | Stroke path |
| `f` | `Fill()` | Fill path |
| `B` | `FillStroke()` | Fill and stroke |
| `cm`+`Do` | `DrawImage(...)` | Place image XObject |

---

## Core Fonts

Folio embeds metrics for 13 standard PDF core fonts. These fonts are built into every PDF viewer and require no embedding.

| Key | PDF BaseFont Name |
|-----|-------------------|
| `helvetica` | Helvetica |
| `helveticaB` | Helvetica-Bold |
| `helveticaI` | Helvetica-Oblique |
| `helveticaBI` | Helvetica-BoldOblique |
| `courier` | Courier |
| `courierB` | Courier-Bold |
| `courierI` | Courier-Oblique |
| `courierBI` | Courier-BoldOblique |
| `times` | Times-Roman |
| `timesB` | Times-Bold |
| `timesI` | Times-Italic |
| `timesBI` | Times-BoldItalic |
| `zapfdingbats` | ZapfDingbats |

Aliases: `"arial"` maps to `"helvetica"`, `"symbol"` maps to `"zapfdingbats"`.

Width data is stored in `internal/resources/fonts_core.go` as `[256]int` arrays (WinAnsiEncoding, 1/1000 text units).

---

## Page Sizes

Defined in `options.go` as `PageSize{WidthPt, HeightPt}`:

| Constant | Width (pt) | Height (pt) | mm |
|----------|-----------|-------------|-----|
| `A3` | 841.89 | 1190.55 | 297 x 420 |
| `A4` | 595.28 | 841.89 | 210 x 297 |
| `A5` | 420.94 | 595.28 | 148 x 210 |
| `Letter` | 612 | 792 | 216 x 279 |
| `Legal` | 612 | 1008 | 216 x 356 |

Custom sizes: `folio.PageSize{WidthPt: 400, HeightPt: 600}`.

---

## Feature Roadmap

### v0.1 (current)
- [x] Pages, text, lines, rectangles
- [x] Core PDF fonts (Helvetica, Courier, Times, ZapfDingbats)
- [x] Cell and MultiCell with word wrapping
- [x] JPEG image embedding with deduplication
- [x] RGB colors (stroke, fill, text)
- [x] Document metadata (title, author, subject)
- [x] WriteTo/Save/Bytes output
- [x] Zlib compression of content streams
- [x] Configurable units (mm, pt, cm, inch)

### v0.2 (layout)
- [ ] Auto page breaks
- [ ] Headers and footers
- [ ] SetAutoPageBreak(enabled, margin)
- [ ] Cursor flow: Write() for inline text
- [ ] Page numbering alias

### v0.3 (resources)
- [ ] PNG image support (with alpha/soft mask)
- [ ] TTF font loading and embedding
- [ ] UTF-8 text support (CIDFont Type2)
- [ ] Font subsetting
- [ ] Resource dedup for TTF fonts

### v0.4 (typography)
- [ ] Text decoration (underline, strikethrough)
- [ ] Character spacing, word spacing
- [ ] Text rise
- [ ] Justified text alignment in MultiCell

### v0.5 (advanced drawing)
- [ ] Circle, Ellipse, Arc
- [ ] Bezier curves (public API)
- [ ] Clipping paths
- [ ] Transforms (rotate, scale, skew)
- [ ] Alpha transparency (ExtGState)
- [ ] Dash patterns

### v0.6 (document features)
- [ ] Internal and external links
- [ ] Bookmarks / outlines
- [ ] Table helper API
- [ ] Gradients (linear, radial)

### v1.0
- [ ] Layers (Optional Content Groups)
- [ ] Password protection
- [ ] Templates
- [ ] SVG path parsing
- [ ] Barcode generation (contrib)

---

## Running Tests

```bash
# All tests
go test ./...

# Verbose
go test ./... -v

# Specific package
go test ./internal/pdfcore/ -v

# Generate a demo PDF
go run ./cmd/demo/
open /tmp/folio_demo.pdf
```

---

## Quick API Reference

### Document

```go
doc := folio.New(opts ...Option)         // create document
doc.SetTitle(s)                          // metadata
doc.SetAuthor(s)
doc.SetSubject(s)
doc.SetCreator(s)
doc.SetMargins(left, top, right)         // user units
doc.SetFont(family, style, size)         // auto-registers
doc.SetDrawColor(r, g, b)               // 0-255
doc.SetFillColor(r, g, b)
doc.SetTextColor(r, g, b)
doc.SetLineWidth(w)                      // user units
doc.RegisterImage(name, io.Reader)       // JPEG
page := doc.AddPage(folio.A4)            // returns *Page
doc.WriteTo(w io.Writer)                 // serialize
doc.Save(path)                           // file output
doc.Bytes()                              // []byte output
doc.Err()                                // check error
```

### Page

```go
page.TextAt(x, y, text)                                   // absolute text
page.Cell(w, h, text, border, align, fill, ln)             // single-line cell
page.MultiCell(w, h, text, border, align, fill)            // word-wrapped
page.Line(x1, y1, x2, y2)                                 // line segment
page.Rect(x, y, w, h, style)                              // "D"/"F"/"DF"
page.DrawImageRect(name, x, y, w, h)                      // registered image
page.GetStringWidth(s)                                     // text measurement
page.SetXY(x, y) / SetX(x) / SetY(y)                     // cursor
page.GetX() / GetY()                                       // cursor position
page.SetFont(family, style, size)                          // page-level font
```

### Options

```go
folio.WithUnit(state.UnitPt)             // points
folio.WithUnit(state.UnitMM)             // millimeters (default)
folio.WithUnit(state.UnitCM)             // centimeters
folio.WithUnit(state.UnitInch)           // inches
folio.WithCompression(false)             // disable zlib
```

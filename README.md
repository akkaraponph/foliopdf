<p align="center">
  <img src="docs/assets/banner-folio.png" alt="Folio" width="600">
</p>

<p align="center">
  A layered PDF generation library for Go. Zero external dependencies.
</p>

```go
doc := folio.New()
doc.SetFont("helvetica", "B", 20)

page := doc.AddPage(folio.A4)
page.TextAt(40, 30, "Hello, Folio!")

doc.Save("hello.pdf")
```

## Why Folio?

- **Zero dependencies** — everything built from scratch on the Go standard library
- **Clean 4-layer architecture** — easy to understand, easy to extend
- **Full Unicode** — TrueType font embedding with CIDFont Type 2
- **Thai language built-in** — dictionary-based word segmentation (~15K words)
- **Rich feature set** — tables, barcodes, forms, signatures, HTML/Markdown, PDF/A

## Install

```bash
go get github.com/akkaraponph/folio
```

Requires Go 1.26+.

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
| [Architecture](docs/architecture.md) | 4-layer internals, how to extend |
| [API Reference](docs/api.md) | Complete function reference |

## Examples

```bash
go run ./cmd/demo              # minimal getting-started
go run ./examples/thai         # Thai language
go run ./examples/invoice      # bilingual business invoice
go run ./examples/showcase     # 15 feature demos
```

See [`examples/`](examples/) for all runnable examples.

## License

MIT


## Fact
I try to build this package by AI Driven Development
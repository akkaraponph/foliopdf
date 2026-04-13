package foliopdf

import (
	"strings"
	"unicode"
)

// MarkdownOption configures the markdown renderer.
type MarkdownOption func(*mdRenderer)

// WithBookmarks enables automatic bookmark generation for headings.
func WithBookmarks() MarkdownOption {
	return func(r *mdRenderer) { r.bookmarks = true }
}

// WithLineHeight sets the line height for paragraph text.
func WithLineHeight(h float64) MarkdownOption {
	return func(r *mdRenderer) { r.lineHeight = h }
}

// Markdown renders a subset of Markdown onto the page starting at the
// current cursor position.
//
// Supported syntax: # headings (1-6), **bold**, *italic*, `code`,
// - unordered lists, 1. ordered lists, --- horizontal rules,
// [text](url) links.
func (p *Page) Markdown(md string, opts ...MarkdownOption) {
	p = p.active()
	r := &mdRenderer{
		page:       p,
		lineHeight: 6,
	}
	for _, opt := range opts {
		opt(r)
	}

	blocks := parseMDBlocks(md)
	r.render(blocks)
}

// mdBlock represents a parsed markdown block.
type mdBlock struct {
	kind    string // "h1"-"h6", "p", "ul", "ol", "hr"
	content string // raw inline content (for h/p/list items)
	items   []string // list items (for ul/ol)
}

// mdInline represents a parsed inline element.
type mdInline struct {
	kind string // "text", "bold", "italic", "code", "link"
	text string
	url  string // for links
}

// mdRenderer holds rendering state.
type mdRenderer struct {
	page       *Page
	bookmarks  bool
	lineHeight float64
}

// parseMDBlocks parses markdown text into a sequence of blocks.
func parseMDBlocks(md string) []mdBlock {
	lines := strings.Split(strings.ReplaceAll(md, "\r\n", "\n"), "\n")
	var blocks []mdBlock
	var paraLines []string
	var listItems []string
	listKind := ""

	flushPara := func() {
		if len(paraLines) > 0 {
			blocks = append(blocks, mdBlock{kind: "p", content: strings.Join(paraLines, " ")})
			paraLines = nil
		}
	}
	flushList := func() {
		if len(listItems) > 0 {
			blocks = append(blocks, mdBlock{kind: listKind, items: listItems})
			listItems = nil
			listKind = ""
		}
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Blank line: flush current paragraph/list.
		if trimmed == "" {
			flushPara()
			flushList()
			continue
		}

		// Horizontal rule: ---, ***, ___
		if isHRule(trimmed) {
			flushPara()
			flushList()
			blocks = append(blocks, mdBlock{kind: "hr"})
			continue
		}

		// Heading: # ... ######
		if strings.HasPrefix(trimmed, "#") {
			flushPara()
			flushList()
			level := 0
			for level < len(trimmed) && trimmed[level] == '#' {
				level++
			}
			if level <= 6 && level < len(trimmed) && trimmed[level] == ' ' {
				blocks = append(blocks, mdBlock{
					kind:    "h" + string(rune('0'+level)),
					content: strings.TrimSpace(trimmed[level+1:]),
				})
				continue
			}
		}

		// Unordered list: - item or * item
		if (strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ")) && len(trimmed) > 2 {
			flushPara()
			if listKind != "" && listKind != "ul" {
				flushList()
			}
			listKind = "ul"
			listItems = append(listItems, strings.TrimSpace(trimmed[2:]))
			continue
		}

		// Ordered list: 1. item, 2. item, etc.
		if idx := strings.Index(trimmed, ". "); idx > 0 && idx <= 3 {
			prefix := trimmed[:idx]
			allDigits := true
			for _, c := range prefix {
				if !unicode.IsDigit(c) {
					allDigits = false
					break
				}
			}
			if allDigits {
				flushPara()
				if listKind != "" && listKind != "ol" {
					flushList()
				}
				listKind = "ol"
				listItems = append(listItems, strings.TrimSpace(trimmed[idx+2:]))
				continue
			}
		}

		// Otherwise: paragraph line.
		flushList()
		paraLines = append(paraLines, trimmed)
	}
	flushPara()
	flushList()
	return blocks
}

// isHRule checks if a line is a horizontal rule (3+ of -, *, or _).
func isHRule(s string) bool {
	if len(s) < 3 {
		return false
	}
	ch := s[0]
	if ch != '-' && ch != '*' && ch != '_' {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] != ch && s[i] != ' ' {
			return false
		}
	}
	return true
}

// parseInline parses inline markdown elements from text.
func parseInline(text string) []mdInline {
	var result []mdInline
	runes := []rune(text)
	n := len(runes)
	i := 0

	flushText := func(s string) {
		if s != "" {
			result = append(result, mdInline{kind: "text", text: s})
		}
	}

	var buf []rune
	for i < n {
		// Link: [text](url)
		if runes[i] == '[' {
			flushText(string(buf))
			buf = nil
			end := findLinkEnd(runes, i)
			if end > 0 {
				linkText, linkURL := extractLink(runes[i:end])
				result = append(result, mdInline{kind: "link", text: linkText, url: linkURL})
				i = end
				continue
			}
		}

		// Code: `text`
		if runes[i] == '`' {
			flushText(string(buf))
			buf = nil
			end := indexRune(runes, '`', i+1)
			if end > 0 {
				result = append(result, mdInline{kind: "code", text: string(runes[i+1 : end])})
				i = end + 1
				continue
			}
		}

		// Bold: **text**
		if i+1 < n && runes[i] == '*' && runes[i+1] == '*' {
			flushText(string(buf))
			buf = nil
			end := indexDoubleRune(runes, '*', i+2)
			if end > 0 {
				// Recursively parse inline content within bold.
				result = append(result, mdInline{kind: "bold", text: string(runes[i+2 : end])})
				i = end + 2
				continue
			}
		}

		// Italic: *text*
		if runes[i] == '*' && (i+1 < n && runes[i+1] != '*') {
			flushText(string(buf))
			buf = nil
			end := indexRune(runes, '*', i+1)
			if end > 0 {
				result = append(result, mdInline{kind: "italic", text: string(runes[i+1 : end])})
				i = end + 1
				continue
			}
		}

		buf = append(buf, runes[i])
		i++
	}
	flushText(string(buf))
	return result
}

// indexRune finds the next occurrence of r in runes starting at start.
func indexRune(runes []rune, r rune, start int) int {
	for i := start; i < len(runes); i++ {
		if runes[i] == r {
			return i
		}
	}
	return -1
}

// indexDoubleRune finds the next occurrence of two consecutive r runes.
func indexDoubleRune(runes []rune, r rune, start int) int {
	for i := start; i+1 < len(runes); i++ {
		if runes[i] == r && runes[i+1] == r {
			return i
		}
	}
	return -1
}

// findLinkEnd finds the closing ) of a [text](url) link.
func findLinkEnd(runes []rune, start int) int {
	// Find ]
	closeBracket := indexRune(runes, ']', start+1)
	if closeBracket < 0 || closeBracket+1 >= len(runes) || runes[closeBracket+1] != '(' {
		return -1
	}
	closeParen := indexRune(runes, ')', closeBracket+2)
	if closeParen < 0 {
		return -1
	}
	return closeParen + 1
}

// extractLink extracts text and URL from [text](url) runes.
func extractLink(runes []rune) (string, string) {
	closeBracket := indexRune(runes, ']', 1)
	text := string(runes[1:closeBracket])
	url := string(runes[closeBracket+2 : len(runes)-1])
	return text, url
}

// render draws all blocks onto the page.
func (r *mdRenderer) render(blocks []mdBlock) {
	for _, b := range blocks {
		switch b.kind {
		case "h1", "h2", "h3", "h4", "h5", "h6":
			r.renderHeading(b)
		case "p":
			r.renderParagraph(b)
		case "ul":
			r.renderUnorderedList(b)
		case "ol":
			r.renderOrderedList(b)
		case "hr":
			r.renderHRule()
		}
	}
}

// headingSizes maps heading level to font size in points.
var headingSizes = map[string]float64{
	"h1": 24, "h2": 20, "h3": 16, "h4": 14, "h5": 12, "h6": 10,
}

func (r *mdRenderer) renderHeading(b mdBlock) {
	p := r.page.active()
	doc := p.doc

	size := headingSizes[b.kind]
	lh := size / doc.k * 1.4

	savedFamily := doc.fontFamily
	savedStyle := doc.fontStyle
	savedSize := doc.fontSizePt

	doc.SetFont(doc.fontFamily, "B", size)

	if r.bookmarks {
		level := int(b.kind[1]-'0') - 1
		doc.AddBookmark(b.content, level)
	}

	p = r.page.active()
	p.Write(lh, b.content)
	p = r.page.active()
	p.x = doc.lMargin
	p.y += lh // advance past the heading line

	doc.SetFont(savedFamily, savedStyle, savedSize)
}

func (r *mdRenderer) renderParagraph(b mdBlock) {
	inlines := parseInline(b.content)
	r.renderInlines(inlines)

	p := r.page.active()
	p.x = p.doc.lMargin
	p.y += r.lineHeight // advance past the last text line
}

func (r *mdRenderer) renderUnorderedList(b mdBlock) {
	p := r.page.active()
	doc := p.doc
	indent := 5.0

	for _, item := range b.items {
		p = r.page.active()
		p.x = doc.lMargin + indent

		// Bullet (ASCII-safe for core fonts)
		p.Write(r.lineHeight, "- ")

		// Item content with inline formatting
		inlines := parseInline(item)
		r.renderInlines(inlines)

		p = r.page.active()
		p.x = doc.lMargin
		p.y += r.lineHeight
	}
	p = r.page.active()
	p.y += r.lineHeight * 0.3
}

func (r *mdRenderer) renderOrderedList(b mdBlock) {
	p := r.page.active()
	doc := p.doc
	indent := 5.0

	for i, item := range b.items {
		p = r.page.active()
		p.x = doc.lMargin + indent

		// Number
		p.Write(r.lineHeight, string(rune('1'+i))+". ")

		inlines := parseInline(item)
		r.renderInlines(inlines)

		p = r.page.active()
		p.x = doc.lMargin
		p.y += r.lineHeight
	}
	p = r.page.active()
	p.y += r.lineHeight * 0.3
}

func (r *mdRenderer) renderHRule() {
	p := r.page.active()
	doc := p.doc
	y := p.y + r.lineHeight*0.5
	p.Line(doc.lMargin, y, p.w-doc.rMargin, y)
	p.y = y + r.lineHeight*0.5
}

// renderInlines renders a sequence of inline elements.
func (r *mdRenderer) renderInlines(inlines []mdInline) {
	p := r.page.active()
	doc := p.doc

	for _, inl := range inlines {
		p = r.page.active()
		switch inl.kind {
		case "text":
			p.Write(r.lineHeight, inl.text)

		case "bold":
			savedStyle := doc.fontStyle
			doc.SetFontStyle("B")
			p = r.page.active()
			p.Write(r.lineHeight, inl.text)
			doc.SetFontStyle(savedStyle)

		case "italic":
			savedStyle := doc.fontStyle
			doc.SetFontStyle("I")
			p = r.page.active()
			p.Write(r.lineHeight, inl.text)
			doc.SetFontStyle(savedStyle)

		case "code":
			savedFamily := doc.fontFamily
			savedStyle := doc.fontStyle
			doc.SetFont("courier", "", doc.fontSizePt)
			p = r.page.active()
			p.Write(r.lineHeight, inl.text)
			doc.SetFont(savedFamily, savedStyle, doc.fontSizePt)

		case "link":
			// Save cursor X for link annotation rect
			startX := p.x
			startY := p.y
			// Draw link text in blue
			savedColor := doc.textColor
			doc.SetTextColor(0, 0, 200)
			p.Write(r.lineHeight, inl.text)
			doc.textColor = savedColor
			// Add clickable link
			endX := r.page.active().x
			p.LinkURL(startX, startY, endX-startX, r.lineHeight, inl.url)
		}
	}
}

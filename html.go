package folio

import (
	"strconv"
	"strings"
)

// HTMLOption configures the HTML renderer.
type HTMLOption func(*htmlRenderer)

// HTML renders a subset of HTML onto the page starting at the current
// cursor position.
//
// Supported tags: <h1>-<h6>, <p>, <b>, <i>, <u>, <br>, <hr>,
// <ul>/<ol>/<li>, <table>/<tr>/<td>, <a href="...">.
//
// Supported inline CSS (via style attribute): color, font-size,
// text-align, background-color.
func (p *Page) HTML(html string, opts ...HTMLOption) {
	p = p.active()
	r := &htmlRenderer{
		page:       p,
		lineHeight: 6,
	}
	for _, opt := range opts {
		opt(r)
	}

	nodes := parseHTML(html)
	r.renderNodes(nodes)
}

// htmlNode represents a parsed HTML element or text node.
type htmlNode struct {
	tag      string
	attrs    map[string]string
	children []*htmlNode
	text     string // non-empty for text nodes
}

// htmlRenderer holds rendering state.
type htmlRenderer struct {
	page       *Page
	lineHeight float64
	listIdx    int // current ordered list index
}

// parseHTML parses an HTML string into a tree of nodes.
// This is a minimal parser supporting the subset of tags we render.
func parseHTML(s string) []*htmlNode {
	p := &htmlParser{input: s}
	return p.parseNodes()
}

type htmlParser struct {
	input string
	pos   int
}

func (p *htmlParser) parseNodes() []*htmlNode {
	var nodes []*htmlNode
	for p.pos < len(p.input) {
		if p.input[p.pos] == '<' {
			// Check for closing tag.
			if p.pos+1 < len(p.input) && p.input[p.pos+1] == '/' {
				break // let parent handle the closing tag
			}
			node := p.parseElement()
			if node != nil {
				nodes = append(nodes, node)
			}
		} else {
			text := p.parseText()
			if text != "" {
				nodes = append(nodes, &htmlNode{text: text})
			}
		}
	}
	return nodes
}

func (p *htmlParser) parseText() string {
	start := p.pos
	for p.pos < len(p.input) && p.input[p.pos] != '<' {
		p.pos++
	}
	text := p.input[start:p.pos]
	// Decode basic HTML entities.
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	return text
}

func (p *htmlParser) parseElement() *htmlNode {
	if p.pos >= len(p.input) || p.input[p.pos] != '<' {
		return nil
	}
	p.pos++ // skip '<'

	// Read tag name.
	tag := p.readWord()
	tag = strings.ToLower(tag)

	// Read attributes.
	attrs := p.readAttrs()

	// Self-closing tag?
	selfClose := false
	p.skipWhitespace()
	if p.pos < len(p.input) && p.input[p.pos] == '/' {
		selfClose = true
		p.pos++
	}

	// Skip '>'.
	if p.pos < len(p.input) && p.input[p.pos] == '>' {
		p.pos++
	}

	node := &htmlNode{tag: tag, attrs: attrs}

	// Void elements (no closing tag).
	if selfClose || isVoidElement(tag) {
		return node
	}

	// Parse children until closing tag.
	node.children = p.parseNodes()

	// Consume closing tag </tag>.
	p.consumeClosingTag(tag)

	return node
}

func (p *htmlParser) readWord() string {
	start := p.pos
	for p.pos < len(p.input) && !isHTMLDelim(p.input[p.pos]) {
		p.pos++
	}
	return p.input[start:p.pos]
}

func (p *htmlParser) readAttrs() map[string]string {
	attrs := make(map[string]string)
	for {
		p.skipWhitespace()
		if p.pos >= len(p.input) || p.input[p.pos] == '>' || p.input[p.pos] == '/' {
			break
		}
		name := p.readWord()
		if name == "" {
			break
		}
		name = strings.ToLower(name)

		p.skipWhitespace()
		if p.pos < len(p.input) && p.input[p.pos] == '=' {
			p.pos++ // skip '='
			p.skipWhitespace()
			val := p.readAttrValue()
			attrs[name] = val
		} else {
			attrs[name] = name // boolean attribute
		}
	}
	return attrs
}

func (p *htmlParser) readAttrValue() string {
	if p.pos >= len(p.input) {
		return ""
	}
	quote := p.input[p.pos]
	if quote == '"' || quote == '\'' {
		p.pos++ // skip opening quote
		start := p.pos
		for p.pos < len(p.input) && p.input[p.pos] != quote {
			p.pos++
		}
		val := p.input[start:p.pos]
		if p.pos < len(p.input) {
			p.pos++ // skip closing quote
		}
		return val
	}
	// Unquoted value.
	return p.readWord()
}

func (p *htmlParser) consumeClosingTag(tag string) {
	if p.pos+2 >= len(p.input) || p.input[p.pos] != '<' || p.input[p.pos+1] != '/' {
		return
	}
	// Skip past </tag>
	end := strings.IndexByte(p.input[p.pos:], '>')
	if end >= 0 {
		p.pos += end + 1
	}
}

func (p *htmlParser) skipWhitespace() {
	for p.pos < len(p.input) && (p.input[p.pos] == ' ' || p.input[p.pos] == '\t' || p.input[p.pos] == '\n' || p.input[p.pos] == '\r') {
		p.pos++
	}
}

func isHTMLDelim(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r' || b == '>' || b == '/' || b == '='
}

func isVoidElement(tag string) bool {
	switch tag {
	case "br", "hr", "img":
		return true
	}
	return false
}

// renderNodes renders a list of HTML nodes.
func (r *htmlRenderer) renderNodes(nodes []*htmlNode) {
	for _, n := range nodes {
		if n.text != "" {
			r.renderText(n.text)
			continue
		}
		r.renderElement(n)
	}
}

func (r *htmlRenderer) renderText(text string) {
	// Collapse whitespace like a browser.
	text = collapseWhitespace(text)
	if text == "" {
		return
	}
	p := r.page.active()
	p.Write(r.lineHeight, text)
}

func collapseWhitespace(s string) string {
	var buf strings.Builder
	lastWasSpace := false
	for _, r := range s {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			if !lastWasSpace {
				buf.WriteByte(' ')
				lastWasSpace = true
			}
		} else {
			buf.WriteRune(r)
			lastWasSpace = false
		}
	}
	return buf.String()
}

// renderElement dispatches rendering based on tag name.
func (r *htmlRenderer) renderElement(n *htmlNode) {
	p := r.page.active()
	doc := p.doc

	// Apply inline style before rendering.
	savedColor := [3]int{
		int(doc.textColor.R * 255),
		int(doc.textColor.G * 255),
		int(doc.textColor.B * 255),
	}
	savedSize := doc.fontSizePt
	style := parseInlineStyle(n.attrs["style"])
	r.applyStyle(style)

	switch n.tag {
	case "h1", "h2", "h3", "h4", "h5", "h6":
		r.renderHeading(n)
	case "p":
		r.renderBlock(n)
	case "b", "strong":
		r.withFontStyle("B", func() { r.renderNodes(n.children) })
	case "i", "em":
		r.withFontStyle("I", func() { r.renderNodes(n.children) })
	case "u":
		doc.SetUnderline(true)
		r.renderNodes(n.children)
		doc.SetUnderline(false)
	case "br":
		p = r.page.active()
		p.x = doc.lMargin
		p.y += r.lineHeight
	case "hr":
		p = r.page.active()
		y := p.y + r.lineHeight*0.5
		p.Line(doc.lMargin, y, p.w-doc.rMargin, y)
		p.y = y + r.lineHeight*0.5
	case "ul":
		r.renderUL(n)
	case "ol":
		r.renderOL(n)
	case "li":
		r.renderNodes(n.children)
	case "table":
		r.renderTable(n)
	case "a":
		r.renderLink(n)
	default:
		// Unknown tag: just render children.
		r.renderNodes(n.children)
	}

	// Restore style.
	r.restoreStyle(style, savedColor, savedSize)
}

// htmlHeadingSizes maps heading tag to font size.
var htmlHeadingSizes = map[string]float64{
	"h1": 24, "h2": 20, "h3": 16, "h4": 14, "h5": 12, "h6": 10,
}

func (r *htmlRenderer) renderHeading(n *htmlNode) {
	p := r.page.active()
	doc := p.doc
	size := htmlHeadingSizes[n.tag]
	lh := size / doc.k * 1.4

	savedFamily := doc.fontFamily
	savedStyle := doc.fontStyle
	savedSize := doc.fontSizePt

	doc.SetFont(doc.fontFamily, "B", size)
	p = r.page.active()

	// Render children as inline content.
	r.lineHeight = lh
	r.renderNodes(n.children)
	r.lineHeight = 6

	p = r.page.active()
	p.x = doc.lMargin
	p.y += lh // advance past the heading line

	doc.SetFont(savedFamily, savedStyle, savedSize)
}

func (r *htmlRenderer) renderBlock(n *htmlNode) {
	r.renderNodes(n.children)
	p := r.page.active()
	p.x = p.doc.lMargin
	p.y += r.lineHeight // advance past the last text line
}

func (r *htmlRenderer) renderUL(n *htmlNode) {
	p := r.page.active()
	doc := p.doc
	indent := 5.0

	for _, child := range n.children {
		if child.tag != "li" {
			continue
		}
		p = r.page.active()
		p.x = doc.lMargin + indent
		p.Write(r.lineHeight, "- ")
		r.renderNodes(child.children)
		p = r.page.active()
		p.x = doc.lMargin
		p.y += r.lineHeight
	}
	p = r.page.active()
	p.y += r.lineHeight * 0.3
}

func (r *htmlRenderer) renderOL(n *htmlNode) {
	p := r.page.active()
	doc := p.doc
	indent := 5.0
	idx := 1

	for _, child := range n.children {
		if child.tag != "li" {
			continue
		}
		p = r.page.active()
		p.x = doc.lMargin + indent
		p.Write(r.lineHeight, strconv.Itoa(idx)+". ")
		r.renderNodes(child.children)
		p = r.page.active()
		p.x = doc.lMargin
		p.y += r.lineHeight
		idx++
	}
	p = r.page.active()
	p.y += r.lineHeight * 0.3
}

func (r *htmlRenderer) renderLink(n *htmlNode) {
	p := r.page.active()
	doc := p.doc
	href := n.attrs["href"]

	startX := p.x
	startY := p.y
	savedColor := doc.textColor
	doc.SetTextColor(0, 0, 200)

	r.renderNodes(n.children)

	doc.textColor = savedColor
	endX := r.page.active().x

	if href != "" {
		p.LinkURL(startX, startY, endX-startX, r.lineHeight, href)
	}
}

// renderTable renders an HTML <table> using the existing Table helper.
func (r *htmlRenderer) renderTable(n *htmlNode) {
	p := r.page.active()
	doc := p.doc

	// Collect rows.
	var rows [][]string
	for _, child := range n.children {
		switch child.tag {
		case "tr":
			var cells []string
			for _, td := range child.children {
				if td.tag == "td" || td.tag == "th" {
					cells = append(cells, extractText(td))
				}
			}
			if len(cells) > 0 {
				rows = append(rows, cells)
			}
		case "thead", "tbody", "tfoot":
			for _, tr := range child.children {
				if tr.tag == "tr" {
					var cells []string
					for _, td := range tr.children {
						if td.tag == "td" || td.tag == "th" {
							cells = append(cells, extractText(td))
						}
					}
					if len(cells) > 0 {
						rows = append(rows, cells)
					}
				}
			}
		}
	}

	if len(rows) == 0 {
		return
	}

	// Determine column count.
	maxCols := 0
	for _, row := range rows {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}

	// Calculate equal column widths.
	availW := p.w - doc.lMargin - doc.rMargin
	colW := availW / float64(maxCols)
	widths := make([]float64, maxCols)
	for i := range widths {
		widths[i] = colW
	}

	tbl := NewTable(doc, p)
	tbl.SetWidths(widths...)

	for i, row := range rows {
		// Pad row to maxCols.
		for len(row) < maxCols {
			row = append(row, "")
		}
		if i == 0 {
			tbl.Header(row...)
		} else {
			tbl.Row(row...)
		}
	}
}

// extractText recursively extracts all text content from an HTML node tree.
func extractText(n *htmlNode) string {
	if n.text != "" {
		return collapseWhitespace(n.text)
	}
	var sb strings.Builder
	for _, child := range n.children {
		sb.WriteString(extractText(child))
	}
	return strings.TrimSpace(sb.String())
}

// withFontStyle temporarily changes font style, runs fn, then restores.
func (r *htmlRenderer) withFontStyle(style string, fn func()) {
	doc := r.page.doc
	saved := doc.fontStyle
	doc.SetFontStyle(style)
	fn()
	doc.SetFontStyle(saved)
}

// cssStyle holds parsed inline CSS properties.
type cssStyle struct {
	color    [3]int
	hasColor bool
	fontSize float64
}

// parseInlineStyle parses a CSS style attribute into a cssStyle.
func parseInlineStyle(s string) cssStyle {
	var cs cssStyle
	if s == "" {
		return cs
	}

	for _, decl := range strings.Split(s, ";") {
		decl = strings.TrimSpace(decl)
		parts := strings.SplitN(decl, ":", 2)
		if len(parts) != 2 {
			continue
		}
		prop := strings.TrimSpace(strings.ToLower(parts[0]))
		val := strings.TrimSpace(parts[1])

		switch prop {
		case "color":
			if r, g, b, ok := parseCSSColor(val); ok {
				cs.color = [3]int{r, g, b}
				cs.hasColor = true
			}
		case "font-size":
			if size := parseCSSFontSize(val); size > 0 {
				cs.fontSize = size
			}
		}
	}
	return cs
}

// parseCSSColor parses a simple CSS color: #rrggbb, #rgb, or named colors.
func parseCSSColor(s string) (int, int, int, bool) {
	s = strings.TrimSpace(strings.ToLower(s))

	// Named colors.
	switch s {
	case "red":
		return 255, 0, 0, true
	case "green":
		return 0, 128, 0, true
	case "blue":
		return 0, 0, 255, true
	case "black":
		return 0, 0, 0, true
	case "white":
		return 255, 255, 255, true
	case "gray", "grey":
		return 128, 128, 128, true
	}

	if !strings.HasPrefix(s, "#") {
		return 0, 0, 0, false
	}
	hex := s[1:]

	if len(hex) == 3 {
		r, _ := strconv.ParseUint(string(hex[0])+string(hex[0]), 16, 8)
		g, _ := strconv.ParseUint(string(hex[1])+string(hex[1]), 16, 8)
		b, _ := strconv.ParseUint(string(hex[2])+string(hex[2]), 16, 8)
		return int(r), int(g), int(b), true
	}
	if len(hex) == 6 {
		r, _ := strconv.ParseUint(hex[0:2], 16, 8)
		g, _ := strconv.ParseUint(hex[2:4], 16, 8)
		b, _ := strconv.ParseUint(hex[4:6], 16, 8)
		return int(r), int(g), int(b), true
	}
	return 0, 0, 0, false
}

// parseCSSFontSize parses "12pt", "16px", or plain numbers.
func parseCSSFontSize(s string) float64 {
	s = strings.TrimSpace(strings.ToLower(s))
	s = strings.TrimSuffix(s, "pt")
	s = strings.TrimSuffix(s, "px")
	v, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return 0
	}
	return v
}

func (r *htmlRenderer) applyStyle(cs cssStyle) {
	doc := r.page.doc
	if cs.hasColor {
		doc.SetTextColor(cs.color[0], cs.color[1], cs.color[2])
	}
	if cs.fontSize > 0 {
		doc.SetFontSize(cs.fontSize)
	}
}

func (r *htmlRenderer) restoreStyle(cs cssStyle, savedColor [3]int, savedSize float64) {
	doc := r.page.doc
	if cs.hasColor {
		doc.SetTextColor(savedColor[0], savedColor[1], savedColor[2])
	}
	if cs.fontSize > 0 {
		doc.SetFontSize(savedSize)
	}
}

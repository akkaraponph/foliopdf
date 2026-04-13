package folio

// Layout orchestration helpers for vertical flow, pagination, and
// block composition. These sit on top of the existing drawing
// primitives (Cell, MultiCell, Write) and page-break machinery.

// ---- Spacer ----------------------------------------------------------------

// Spacer advances the cursor by h user units vertically. If the gap
// would overflow past the bottom margin, a page break is triggered
// first. No PDF operators are emitted — it is purely cursor movement.
func (p *Page) Spacer(h float64) {
	p = p.active()
	if p.doc.err != nil {
		return
	}
	p = p.checkPageBreak(h)
	p.y += h
	p.x = p.doc.lMargin
}

// ---- PageBreakIfNeeded -----------------------------------------------------

// PageBreakIfNeeded checks whether content of height h (in user units)
// would overflow past the bottom margin. If so, a new page is created
// and the method returns true. The caller's *Page reference remains
// valid thanks to the forwarding-pointer chain.
func (p *Page) PageBreakIfNeeded(h float64) bool {
	p = p.active()
	if p.doc.err != nil {
		return false
	}
	np := p.checkPageBreak(h)
	return np != p
}

// ---- KeepTogether ----------------------------------------------------------

// KeepTogether executes fn and guarantees that the content it draws is
// not split across pages. It does a measurement pass to determine the
// height, then either draws on the current page (if it fits) or forces
// a page break first.
//
// fn must only use normal drawing methods (Cell, MultiCell, Write,
// Spacer, Paragraph, etc.). It must not call doc.AddPage directly.
func (p *Page) KeepTogether(fn func()) {
	p = p.active()
	d := p.doc
	if d.err != nil {
		return
	}

	// Save full state for rollback.
	savedStreamLen := p.stream.Len()
	savedX, savedY := p.x, p.y
	savedPageFont := p.fontFamily
	savedPageStyle := p.fontStyle
	savedPageSizePt := p.fontSizePt
	savedPageFontEntry := p.fontEntry
	savedDocState := d.saveDocState()
	savedAutoBreak := d.autoPageBreak
	savedNext := p.next

	// Measurement pass: suppress page breaks so content stays on this page.
	d.autoPageBreak = false
	fn()
	measuredH := p.active().y - savedY

	// Rollback: truncate stream, restore cursor and state.
	p.stream.Truncate(savedStreamLen)
	p.x, p.y = savedX, savedY
	p.next = savedNext
	p.fontFamily = savedPageFont
	p.fontStyle = savedPageStyle
	p.fontSizePt = savedPageSizePt
	p.fontEntry = savedPageFontEntry
	d.restoreDocState(savedDocState)
	d.autoPageBreak = savedAutoBreak

	// Re-emit the current font so the PDF stream state is consistent
	// after truncation.
	if fe := p.effectiveFontEntry(); fe != nil {
		p.applyFont(fe, p.effectiveFontSizePt())
	}

	// Decision: fits on current page, or already at top (can't avoid overflow)?
	if savedY+measuredH <= p.h-d.bMargin || savedY <= d.tMargin {
		fn()
	} else {
		np := d.AddPage(p.size)
		p.next = np
		fn()
	}
}

// ---- Paragraph -------------------------------------------------------------

// ParagraphOption configures a Paragraph call.
type ParagraphOption func(*paragraphConfig)

type paragraphConfig struct {
	fontFamily  string
	fontStyle   string
	fontSize    float64
	align       string
	lineHeight  float64
	spaceBefore float64
	spaceAfter  float64
	textColor   [3]int
	hasColor    bool
	indent      float64
}

// ParagraphFont sets the font for the paragraph.
func ParagraphFont(family, style string, size float64) ParagraphOption {
	return func(c *paragraphConfig) {
		c.fontFamily = family
		c.fontStyle = style
		c.fontSize = size
	}
}

// ParagraphAlign sets text alignment: "L", "C", "R", or "J" (justified).
func ParagraphAlign(align string) ParagraphOption {
	return func(c *paragraphConfig) { c.align = align }
}

// ParagraphLineHeight sets the line height in user units.
func ParagraphLineHeight(h float64) ParagraphOption {
	return func(c *paragraphConfig) { c.lineHeight = h }
}

// ParagraphSpaceBefore adds vertical space before the paragraph.
func ParagraphSpaceBefore(h float64) ParagraphOption {
	return func(c *paragraphConfig) { c.spaceBefore = h }
}

// ParagraphSpaceAfter adds vertical space after the paragraph.
func ParagraphSpaceAfter(h float64) ParagraphOption {
	return func(c *paragraphConfig) { c.spaceAfter = h }
}

// ParagraphTextColor sets the text color (0-255 RGB).
func ParagraphTextColor(r, g, b int) ParagraphOption {
	return func(c *paragraphConfig) {
		c.textColor = [3]int{r, g, b}
		c.hasColor = true
	}
}

// ParagraphIndent sets first-line indent in user units.
func ParagraphIndent(indent float64) ParagraphOption {
	return func(c *paragraphConfig) { c.indent = indent }
}

// Paragraph draws a word-wrapped text block with configurable style,
// alignment, line height, and spacing. It is a higher-level wrapper
// around MultiCell.
func (p *Page) Paragraph(text string, opts ...ParagraphOption) {
	p = p.active()
	d := p.doc
	if d.err != nil {
		return
	}

	cfg := paragraphConfig{
		align:      "L",
		lineHeight: 5, // sensible default
	}
	for _, o := range opts {
		o(&cfg)
	}

	// Space before.
	if cfg.spaceBefore > 0 {
		p.Spacer(cfg.spaceBefore)
		p = p.active()
	}

	// Save state.
	savedFont := d.GetFontFamily()
	savedStyle := d.GetFontStyle()
	savedSize := d.GetFontSize()
	savedTC := d.textColor

	// Apply paragraph style.
	if cfg.fontFamily != "" {
		d.SetFont(cfg.fontFamily, cfg.fontStyle, cfg.fontSize)
	}
	if cfg.hasColor {
		d.SetTextColor(cfg.textColor[0], cfg.textColor[1], cfg.textColor[2])
	}

	// Compute line height from font if not explicitly set.
	if cfg.lineHeight <= 0 {
		cfg.lineHeight = 5
	}

	// First-line indent.
	if cfg.indent > 0 {
		p = p.active()
		p.x = d.lMargin + cfg.indent
	}

	p = p.active()
	p.MultiCell(0, cfg.lineHeight, text, "", cfg.align, false)

	// Restore state.
	if cfg.fontFamily != "" && savedFont != "" {
		d.SetFont(savedFont, savedStyle, savedSize)
	}
	if cfg.hasColor {
		tc := savedTC
		d.SetTextColor(int(tc.R*255+0.5), int(tc.G*255+0.5), int(tc.B*255+0.5))
	}

	// Space after.
	if cfg.spaceAfter > 0 {
		p = p.active()
		p.Spacer(cfg.spaceAfter)
	}
}

// ---- Stack -----------------------------------------------------------------

// Stack executes drawing functions sequentially, resetting X to the
// left margin between each block. This provides a simple vertical
// flow for composing heterogeneous content (paragraphs, tables,
// spacers, images, etc.).
func (p *Page) Stack(blocks ...func()) {
	p = p.active()
	d := p.doc
	if d.err != nil {
		return
	}
	for _, block := range blocks {
		if d.err != nil {
			return
		}
		block()
		q := p.active()
		q.x = d.lMargin
	}
}

package presspdf

// Tagged PDF (structure tree) support for accessibility.
//
// Usage:
//
//	doc := presspdf.New()
//	doc.SetTagged(true)
//	page := doc.AddPage(presspdf.A4)
//
//	page.BeginTag("H1")
//	page.TextAt(20, 20, "Chapter 1")
//	page.EndTag()
//
//	page.BeginTag("P")
//	page.MultiCell(170, 6, "Paragraph text...", "", "L", false)
//	page.EndTag()

// structElement represents a node in the PDF structure tree.
type structElement struct {
	tag      string // PDF structure type: P, H1, H2, Span, Table, TR, TD, Figure, etc.
	page     *Page
	mcid     int    // marked content identifier (per page)
	children []*structElement
	altText  string // alternative text for Figure elements
	parent   *structElement
	objNum   int    // PDF object number, set during serialization
}

// SetTagged enables or disables tagged PDF output. When enabled, the PDF
// includes a structure tree for accessibility (screen readers, reflow).
// Must be called before adding any pages.
func (d *Document) SetTagged(tagged bool) {
	if d.err != nil {
		return
	}
	d.tagged = tagged
	if tagged && d.structRoot == nil {
		d.structRoot = &structElement{tag: "Document"}
	}
}

// BeginTag starts a tagged content section on the page. The tag should be
// a standard PDF structure type:
//   - "H1" through "H6" for headings
//   - "P" for paragraphs
//   - "Span" for inline text
//   - "Table", "TR", "TH", "TD" for tables
//   - "Figure" for images/graphics
//   - "L", "LI" for lists
//   - "Link" for hyperlinks
//
// Tags can be nested. Each BeginTag must be paired with EndTag.
func (p *Page) BeginTag(tag string) {
	p = p.active()
	if p.doc.err != nil || !p.doc.tagged {
		return
	}

	mcid := p.nextMCID
	p.nextMCID++

	elem := &structElement{
		tag:  tag,
		page: p,
		mcid: mcid,
	}

	// Attach to tree.
	parent := p.currentTag
	if parent == nil {
		parent = p.doc.structRoot
	}
	elem.parent = parent
	parent.children = append(parent.children, elem)
	p.currentTag = elem

	// Track this page's marked content for the StructParents mapping.
	p.structElements = append(p.structElements, elem)

	// Emit BDC operator in the content stream.
	p.stream.BeginMarkedContentProp(tag, mcid)
}

// BeginTagAlt starts a tagged content section with alternative text.
// Primarily used for Figure elements to provide descriptive text for
// screen readers.
func (p *Page) BeginTagAlt(tag, altText string) {
	p = p.active()
	if p.doc.err != nil || !p.doc.tagged {
		return
	}

	mcid := p.nextMCID
	p.nextMCID++

	elem := &structElement{
		tag:     tag,
		page:    p,
		mcid:    mcid,
		altText: altText,
	}

	parent := p.currentTag
	if parent == nil {
		parent = p.doc.structRoot
	}
	elem.parent = parent
	parent.children = append(parent.children, elem)
	p.currentTag = elem
	p.structElements = append(p.structElements, elem)

	p.stream.BeginMarkedContentProp(tag, mcid)
}

// EndTag ends the current tagged content section.
func (p *Page) EndTag() {
	p = p.active()
	if p.doc.err != nil || !p.doc.tagged {
		return
	}

	if p.currentTag != nil {
		p.currentTag = p.currentTag.parent
		// If we've popped back to the document root, set nil
		if p.currentTag == p.doc.structRoot {
			p.currentTag = nil
		}
	}

	p.stream.EndMarkedContent()
}

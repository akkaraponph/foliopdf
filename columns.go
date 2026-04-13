package presspdf

// ColumnLayout provides a multi-column layout mode. While active,
// existing text methods (Cell, MultiCell, Write) automatically flow
// within the current column because they respect the document margins,
// which ColumnLayout temporarily adjusts.
//
// Usage:
//
//	cols := presspdf.NewColumnLayout(doc, page, 2, 5)
//	cols.Begin()
//	page.MultiCell(0, 5, leftText, "", "L", false)
//	cols.NextColumn()
//	page.MultiCell(0, 5, rightText, "", "L", false)
//	cols.End()
type ColumnLayout struct {
	doc  *Document
	page *Page

	numCols int     // number of columns
	gutter  float64 // gap between columns in user units
	colW    float64 // computed width of each column

	curCol int     // 0-based current column index
	startY float64 // Y position where columns begin

	// saved margins to restore on End
	savedLMargin float64
	savedRMargin float64
}

// NewColumnLayout creates a multi-column layout. numCols is the number
// of columns, gutter is the horizontal gap between columns in user units.
// The layout uses the available width between the document's left and
// right margins.
func NewColumnLayout(doc *Document, page *Page, numCols int, gutter float64) *ColumnLayout {
	if numCols < 1 {
		numCols = 1
	}
	return &ColumnLayout{
		doc:     doc,
		page:    page,
		numCols: numCols,
		gutter:  gutter,
	}
}

// Begin activates multi-column mode. It saves the current margins and
// cursor position, computes column widths, and constrains the margins
// to the first column.
func (cl *ColumnLayout) Begin() {
	d := cl.doc
	p := cl.page.active()

	cl.savedLMargin = d.lMargin
	cl.savedRMargin = d.rMargin
	cl.startY = p.y

	// Compute column width from available page width.
	avail := p.w - d.lMargin - d.rMargin
	totalGutter := cl.gutter * float64(cl.numCols-1)
	cl.colW = (avail - totalGutter) / float64(cl.numCols)

	cl.curCol = 0
	cl.applyColumnMargins()
}

// NextColumn advances to the next column. If the current column is the
// last one, a new page is created and columns restart from the first.
func (cl *ColumnLayout) NextColumn() {
	cl.curCol++
	if cl.curCol >= cl.numCols {
		// Wrap to first column on a new page.
		cl.curCol = 0
		np := cl.doc.AddPage(cl.page.active().size)
		cl.page = np
		cl.startY = cl.doc.tMargin
	}

	p := cl.page.active()
	p.y = cl.startY
	cl.applyColumnMargins()
}

// End deactivates multi-column mode and restores the original margins.
func (cl *ColumnLayout) End() {
	cl.doc.lMargin = cl.savedLMargin
	cl.doc.rMargin = cl.savedRMargin
	p := cl.page.active()
	p.x = cl.savedLMargin
}

// CurrentColumn returns the 0-based index of the active column.
func (cl *ColumnLayout) CurrentColumn() int {
	return cl.curCol
}

// Page returns the current active page (which may differ from the
// initial page if columns wrapped onto a new page).
func (cl *ColumnLayout) Page() *Page {
	return cl.page.active()
}

// applyColumnMargins sets the document margins so that drawing is
// confined to the current column.
func (cl *ColumnLayout) applyColumnMargins() {
	d := cl.doc
	p := cl.page.active()

	colX := cl.savedLMargin + float64(cl.curCol)*(cl.colW+cl.gutter)
	d.lMargin = colX
	d.rMargin = p.w - colX - cl.colW
	p.x = colX
}

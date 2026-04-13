package presspdf

import (
	"fmt"
	"strings"
)

// TOC builds a table of contents with clickable entries that link to
// bookmarked positions in the document. Each entry displays a title,
// a dot leader, and the target page number.
//
// Usage:
//
//	toc := folio.NewTOC(doc)
//	toc.Add("Chapter 1", 0, chapterPage, chapterY)
//	toc.Add("Section 1.1", 1, sectionPage, sectionY)
//	toc.Render(tocPage, 6)
type TOC struct {
	doc     *Document
	entries []tocEntry
}

// tocEntry represents a single TOC line.
type tocEntry struct {
	title string
	level int
	page  *Page
	y     float64
}

// NewTOC creates a new table of contents builder for the document.
func NewTOC(doc *Document) *TOC {
	return &TOC{doc: doc}
}

// Add registers a TOC entry. title is the display text, level controls
// indentation (0 = top-level), page and y specify the link destination.
func (toc *TOC) Add(title string, level int, page *Page, y float64) {
	toc.entries = append(toc.entries, tocEntry{
		title: title,
		level: level,
		page:  page,
		y:     y,
	})
}

// Render draws the table of contents onto the given page. lineHeight is the
// height of each TOC line in user units. Each entry gets:
//   - Indentation based on level (10 user units per level)
//   - Title text
//   - Dot leader ("....")
//   - Right-aligned page number
//   - Clickable link to the target position
//   - A corresponding bookmark in the outline sidebar
//
// If the TOC overflows the page, Render uses auto page break if enabled.
func (toc *TOC) Render(page *Page, lineHeight float64) {
	toc.renderInternal(page, lineHeight, 0)
}

// RenderWithPageNums is like Render but adds offset to all displayed page
// numbers. This is useful when the TOC itself occupies pages that shift
// the numbering of subsequent content.
func (toc *TOC) RenderWithPageNums(page *Page, lineHeight float64, offset int) {
	toc.renderInternal(page, lineHeight, offset)
}

func (toc *TOC) renderInternal(page *Page, lineHeight float64, pageOffset int) {
	if toc.doc.err != nil {
		return
	}

	d := toc.doc
	indent := 10.0 // user units per level

	// Compute the page number for each entry (1-based).
	pageNums := make([]int, len(toc.entries))
	for i, e := range toc.entries {
		for idx, p := range d.pages {
			if p == e.page {
				pageNums[i] = idx + 1 + pageOffset
				break
			}
		}
	}

	savedCurrent := d.currentPage

	for i, e := range toc.entries {
		p := page.active()
		p = p.checkPageBreak(lineHeight)
		if p != page {
			page = p
		}

		if p.effectiveFontEntry() == nil {
			d.err = fmt.Errorf("TOC.Render: no font set")
			return
		}

		x := d.lMargin + float64(e.level)*indent
		w := p.w - d.rMargin - x

		// Page number string.
		pageStr := fmt.Sprintf("%d", pageNums[i])
		pageW := p.GetStringWidth(pageStr)

		// Title width limit: total width minus page number and some space.
		dotW := p.GetStringWidth(".")
		gap := 2.0 // small gap before page number
		titleMaxW := w - pageW - gap

		// Truncate title if needed.
		title := e.title
		titleW := p.GetStringWidth(title)
		if titleW > titleMaxW {
			for len(title) > 0 && p.GetStringWidth(title+"...") > titleMaxW {
				title = title[:len(title)-1]
			}
			title += "..."
			titleW = p.GetStringWidth(title)
		}

		// Draw title.
		p.SetX(x)
		p.Cell(titleW, lineHeight, title, "", "L", false, 0)

		// Dot leader: fill remaining space between title and page number.
		leaderStart := x + titleW
		leaderEnd := x + w - pageW - gap
		if dotW > 0 && leaderEnd > leaderStart+dotW {
			numDots := int((leaderEnd - leaderStart) / dotW)
			dots := strings.Repeat(".", numDots)
			dotsW := p.GetStringWidth(dots)
			p.SetX(leaderStart)
			p.Cell(dotsW, lineHeight, dots, "", "L", false, 0)
		}

		// Page number (right-aligned).
		p.SetX(x + w - pageW)
		p.Cell(pageW, lineHeight, pageStr, "", "R", false, 0)

		// Internal link: anchor name derived from entry index.
		anchorName := fmt.Sprintf("_toc_%d", i)
		e.page.active().SetXY(0, e.y)
		e.page.active().AddAnchor(anchorName)
		p.LinkAnchor(x, p.y, w, lineHeight, anchorName)

		// Also add bookmark for outline sidebar.
		d.currentPage = p
		d.AddBookmark(e.title, e.level)

		// Move cursor to next line.
		p.SetXY(d.lMargin, p.y+lineHeight)
		page = p
	}

	d.currentPage = savedCurrent
}

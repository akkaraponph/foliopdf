// Example: Lending Rate Sheet
//
// Demonstrates a complex multi-table layout with:
//   - Multiple independent tables positioned side-by-side
//   - Multi-level headers with colspan
//   - Rowspan for merged cells
//   - Dark header rows with white text (per-cell styling)
//   - Small font sizes with dense data grids
//   - Bullet points and multi-line text in cells
//
// Run from the repo root:
//
//	go run ./examples/ratesheet
package main

import (
	"fmt"
	"os"

	"github.com/akkaraponph/folio"
)

// --- Layout constants ---
// Letter Landscape: 279.4mm x 215.9mm, margins 5mm -> usable 269.4 x 205.9

const (
	marginL = 5.0
	marginT = 5.0

	// Column layout: left(147) + gap(2) + mid(44) + gap(2) + right(74) = 269
	col1X = marginL           // 5
	col1W = 147.0             //
	col2X = col1X + col1W + 2 // 154
	col2W = 44.0              //
	col3X = col2X + col2W + 2 // 200
	col3W = 74.0              //

	rowH    = 4.5 // standard row height
	lineH   = 3.0 // line height for multi-line text
	padding = 0.5 // cell padding
)

// Styles
var (
	darkHdr = &folio.CellStyle{
		FontFamily: "helvetica", FontStyle: "B", FontSize: 5.5,
		TextColor: [3]int{255, 255, 255},
		FillColor: [3]int{40, 40, 40},
		Fill:      true,
	}
	bodyReg = &folio.CellStyle{
		FontFamily: "helvetica", FontStyle: "", FontSize: 5,
	}
	bodyB = &folio.CellStyle{
		FontFamily: "helvetica", FontStyle: "B", FontSize: 5,
	}
)

func main() {
	doc := folio.New(folio.WithUnit(folio.UnitMM))
	doc.SetTitle("Commercial Lending Rate Sheet")
	doc.SetMargins(marginL, marginT, marginL)

	page := doc.AddPage(folio.LetterLandscape)

	// --- Top row ---
	ltvBottom := drawLTVTable(doc, page, col1X, marginT)
	drawOverlays(doc, page, col2X, marginT)
	drawProgramRequirements(doc, page, col3X, marginT)

	// --- Middle row ---
	midY := ltvBottom + 2
	incomeBottom := drawIncomeTable(doc, page, col1X, midY)
	drawARMFeatures(doc, page, col2X, midY+8)

	// --- Bottom row ---
	bottomY := incomeBottom + 3
	drawGuidelinesTable(doc, page, col1X, bottomY)

	out := "/tmp/folio_ratesheet.pdf"
	if err := doc.Save(out); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Rate sheet PDF saved to %s\n", out)
}

func newTbl(doc *folio.Document, page *folio.Page) *folio.Table {
	tbl := folio.NewTable(doc, page)
	tbl.SetRowHeight(rowH)
	tbl.SetCellPadding(padding)
	tbl.SetLineHeight(lineH)
	tbl.SetBorder("1")
	return tbl
}

// drawLTVTable returns the Y position of its bottom edge.
func drawLTVTable(doc *folio.Document, page *folio.Page, x, y float64) float64 {
	page.SetXY(x, y)
	tbl := newTbl(doc, page)

	// 11 columns totaling 147mm: 25 + 16 + 10 + 8*12 = 147
	d := 12.0 // data column width
	tbl.SetWidths(25, 16, 10, d, d, d, d, d, d, d, d)

	// Header row 0
	tbl.AddHeader(
		folio.TableCell{Text: "Max LTV by Tier", ColSpan: 3, Align: "C", Style: darkHdr},
		folio.TableCell{Text: "Owner Occupied", ColSpan: 4, Align: "C", Style: darkHdr},
		folio.TableCell{Text: "Non-Owner & Mixed Use", ColSpan: 4, Align: "C", Style: darkHdr},
	)
	// Header row 1
	tbl.AddHeader(
		folio.TableCell{Text: "", ColSpan: 3, Style: darkHdr},
		folio.TableCell{Text: "Standard", ColSpan: 2, Align: "C", Style: darkHdr},
		folio.TableCell{Text: "Reduced Doc", ColSpan: 2, Align: "C", Style: darkHdr},
		folio.TableCell{Text: "Standard", ColSpan: 2, Align: "C", Style: darkHdr},
		folio.TableCell{Text: "Reduced Doc", ColSpan: 2, Align: "C", Style: darkHdr},
	)
	// Header row 2
	tbl.AddHeader(
		folio.TableCell{Text: "Loan Amount", Style: darkHdr, Align: "C"},
		folio.TableCell{Text: "Liquidity", Style: darkHdr, Align: "C"},
		folio.TableCell{Text: "DSCR", Style: darkHdr, Align: "C"},
		folio.TableCell{Text: "Purch. &\nRefi", Style: darkHdr, Align: "C"},
		folio.TableCell{Text: "C/O", Style: darkHdr, Align: "C"},
		folio.TableCell{Text: "Purch. &\nRefi", Style: darkHdr, Align: "C"},
		folio.TableCell{Text: "C/O", Style: darkHdr, Align: "C"},
		folio.TableCell{Text: "Purch. &\nRefi", Style: darkHdr, Align: "C"},
		folio.TableCell{Text: "C/O", Style: darkHdr, Align: "C"},
		folio.TableCell{Text: "Purch. &\nRefi", Style: darkHdr, Align: "C"},
		folio.TableCell{Text: "C/O", Style: darkHdr, Align: "C"},
	)

	type row struct {
		amount, liquidity string
		span              int
		dscr              string
		vals              [8]string
	}
	data := []row{
		{"<= $750,000", "3 Months", 3, "1.25", [8]string{"80%", "70%", "80%", "70%", "75%", "65%", "75%", "65%"}},
		{"", "", 0, "1.15", [8]string{"75%", "65%", "75%", "65%", "70%", "60%", "70%", "60%"}},
		{"", "", 0, "1.00", [8]string{"70%", "60%", "70%", "60%", "65%", "55%", "65%", "55%"}},
		{"<= $1,500,000", "6 Months", 2, "1.25", [8]string{"75%", "65%", "75%", "65%", "70%", "60%", "70%", "60%"}},
		{"", "", 0, "1.15", [8]string{"70%", "60%", "70%", "60%", "65%", "55%", "65%", "55%"}},
		{"<= $2,500,000", "9 Months", 2, "1.25", [8]string{"70%", "60%", "70%", "60%", "65%", "55%", "65%", "55%"}},
		{"", "", 0, "1.15", [8]string{"65%", "55%", "65%", "55%", "60%", "50%", "60%", "50%"}},
		{"<= $5,000,000", "12 Months", 2, "1.25", [8]string{"65%", "55%", "65%", "55%", "60%", "50%", "60%", "50%"}},
		{"", "", 0, "1.15", [8]string{"60%", "50%", "60%", "50%", "55%", "N/A", "55%", "N/A"}},
		{"<= $7,500,000", "18 Months", 1, "1.25", [8]string{"60%", "N/A", "60%", "N/A", "N/A", "N/A", "N/A", "N/A"}},
	}

	for _, r := range data {
		var cells []folio.TableCell
		if r.amount != "" {
			cells = append(cells,
				folio.TableCell{Text: r.amount, RowSpan: r.span, Style: bodyB, Align: "L"},
				folio.TableCell{Text: r.liquidity, RowSpan: r.span, Align: "C"},
			)
		}
		cells = append(cells, folio.TableCell{Text: r.dscr, Align: "C"})
		for _, v := range r.vals {
			cells = append(cells, folio.TableCell{Text: v, Align: "C"})
		}
		tbl.AddRow(cells...)
	}

	tbl.Render()
	return page.GetY()
}

func drawOverlays(doc *folio.Document, page *folio.Page, x, y float64) {
	page.SetXY(x, y)
	tbl := newTbl(doc, page)
	tbl.SetWidths(16, 28)

	tbl.AddHeader(folio.TableCell{Text: "Special Rules", ColSpan: 2, Style: darkHdr, Align: "C"})
	tbl.AddHeader(folio.TableCell{Text: "", ColSpan: 2, Style: darkHdr})
	tbl.AddHeader(folio.TableCell{Text: "Overlays", ColSpan: 2, Style: darkHdr, Align: "C"})

	tbl.AddRow(
		folio.TableCell{Text: "IO Only:", Style: bodyB},
		folio.TableCell{Text: "- Max 75% LTV", Style: bodyReg},
	)
	tbl.AddRow(
		folio.TableCell{Text: "Bridge:", Style: bodyB},
		folio.TableCell{Text: "- 12 mo term max", Style: bodyReg},
	)
	tbl.AddRow(
		folio.TableCell{Text: "", RowSpan: 2},
		folio.TableCell{Text: "- Max 70% LTV (Purch)", Style: bodyReg},
	)
	tbl.AddRow(
		folio.TableCell{Text: "- Max 60% LTV (Refi)", Style: bodyReg},
	)
	tbl.AddRow(
		folio.TableCell{Text: "Mixed Use", Style: bodyB, RowSpan: 3},
		folio.TableCell{Text: "- No subordinate liens", Style: bodyReg},
	)
	tbl.AddRow(folio.TableCell{Text: "- Prepayment penalty", Style: bodyReg})
	tbl.AddRow(folio.TableCell{Text: "  may apply", Style: bodyReg})

	tbl.Render()
}

func drawProgramRequirements(doc *folio.Document, page *folio.Page, x, y float64) {
	// Limits
	page.SetXY(x, y)
	tbl := newTbl(doc, page)
	tbl.SetWidths(40, 34)

	tbl.AddHeader(folio.TableCell{Text: "Program Parameters", ColSpan: 2, Style: darkHdr, Align: "C"})
	tbl.AddHeader(folio.TableCell{Text: "Limits", ColSpan: 2, Style: darkHdr, Align: "C"})

	limits := [][2]string{
		{"Minimum Loan Amount", "$150,000"},
		{"Maximum Loan Amount", "$7,500,000"},
		{"Maximum Cash Out", "$2,000,000"},
		{"Max Cash Out (Non-Owner)", "$1,500,000"},
		{"Minimum DSCR", "1.00x"},
		{"FC/DIL Seasoning", "36 Months"},
		{"BK Seasoning", "36 Months"},
		{"Minimum Liquidity", "$25,000"},
		{"Max Debt Ratio", "55%"},
	}
	for _, r := range limits {
		tbl.AddRow(
			folio.TableCell{Text: r[0], Style: bodyReg},
			folio.TableCell{Text: r[1], Style: bodyReg, Align: "R"},
		)
	}
	tbl.Render()

	// Products
	page.SetXY(x, page.GetY())
	tbl2 := newTbl(doc, page)
	tbl2.SetWidths(col3W)
	tbl2.AddHeader(folio.TableCell{Text: "Products", Style: darkHdr, Align: "C"})
	tbl2.AddRow(folio.TableCell{Text: "5Y Fixed  7/1 ARM  10/1 ARM  15Y Fixed  30Y Fixed", Style: bodyReg, Align: "C"})
	tbl2.Render()

	// Property Type
	page.SetXY(x, page.GetY())
	tbl3 := newTbl(doc, page)
	tbl3.SetWidths(28, 23, 23)
	tbl3.AddHeader(
		folio.TableCell{Text: "Property Type", Style: darkHdr, Align: "C"},
		folio.TableCell{Text: "LTV Max", Style: darkHdr, Align: "C"},
		folio.TableCell{Text: "Notes", Style: darkHdr, Align: "C"},
	)
	for _, r := range [][3]string{
		{"Office / Retail", "75%", "-"},
		{"Industrial", "70%", "-"},
		{"Mixed Use", "65%", "Max 40% res."},
		{"Multi-Family 5+", "80%", "-"},
	} {
		tbl3.AddRow(
			folio.TableCell{Text: r[0], Style: bodyReg},
			folio.TableCell{Text: r[1], Style: bodyReg, Align: "C"},
			folio.TableCell{Text: r[2], Style: bodyReg, Align: "C"},
		)
	}
	tbl3.Render()

	// State Overlays
	page.SetXY(x, page.GetY())
	tbl4 := newTbl(doc, page)
	tbl4.SetWidths(22, 52)
	tbl4.AddHeader(folio.TableCell{Text: "State Restrictions", ColSpan: 2, Style: darkHdr, Align: "C"})
	tbl4.AddRow(
		folio.TableCell{Text: "New York", Style: bodyB},
		folio.TableCell{Text: "Commercial rent control overlay", Style: bodyReg},
	)
	tbl4.AddRow(
		folio.TableCell{Text: "California", Style: bodyB},
		folio.TableCell{Text: "Seismic report required > $2M", Style: bodyReg},
	)
	tbl4.Render()
}

func drawIncomeTable(doc *folio.Document, page *folio.Page, x, y float64) float64 {
	page.SetXY(x, y)
	tbl := newTbl(doc, page)
	tbl.SetWidths(28, col1W-28)

	tbl.AddHeader(folio.TableCell{Text: "Income Verification", ColSpan: 2, Style: darkHdr, Align: "L"})
	tbl.AddRow(
		folio.TableCell{Text: "Full Documentation", Style: bodyB, RowSpan: 2},
		folio.TableCell{Text: "2 Years Business Tax Returns + K-1 / Schedule E", Style: bodyReg},
	)
	tbl.AddRow(folio.TableCell{Text: "Current YTD Profit & Loss Statement (CPA prepared)", Style: bodyReg})
	tbl.AddRow(
		folio.TableCell{Text: "DSCR Only", Style: bodyB},
		folio.TableCell{Text: "Qualify on property cash flow; No personal income docs required", Style: bodyReg},
	)
	tbl.AddRow(
		folio.TableCell{Text: "Reduced Doc\n(Owner-Occ Only)\n(Min 2 yrs in business)", Style: bodyB, RowSpan: 3},
		folio.TableCell{Text: "12 or 24 Months Business Bank Statements", Style: bodyReg},
	)
	tbl.AddRow(folio.TableCell{Text: "12 or 24 Months 1099 Income", Style: bodyReg})
	tbl.AddRow(folio.TableCell{Text: "CPA Letter + 12 Month P&L", Style: bodyReg})

	tbl.Render()
	return page.GetY()
}

func drawARMFeatures(doc *folio.Document, page *folio.Page, x, y float64) {
	page.SetXY(x, y)
	tbl := newTbl(doc, page)
	w := col2W / 3
	tbl.SetWidths(w, w, w)

	tbl.AddHeader(folio.TableCell{Text: "ARM Rate Caps", ColSpan: 3, Style: darkHdr, Align: "C"})
	tbl.AddHeader(
		folio.TableCell{Text: "Product", Style: darkHdr, Align: "C"},
		folio.TableCell{Text: "Adjust", Style: darkHdr, Align: "C"},
		folio.TableCell{Text: "Life Cap", Style: darkHdr, Align: "C"},
	)
	tbl.AddRow(
		folio.TableCell{Text: "7/1 ARM", Align: "C"},
		folio.TableCell{Text: "2/2/5", Align: "C"},
		folio.TableCell{Text: "5%", Align: "C"},
	)
	tbl.AddRow(
		folio.TableCell{Text: "10/1 ARM", Align: "C"},
		folio.TableCell{Text: "2/2/5", Align: "C"},
		folio.TableCell{Text: "5%", Align: "C"},
	)

	tbl.Render()
}

func drawGuidelinesTable(doc *folio.Document, page *folio.Page, x, y float64) {
	page.SetXY(x, y)
	tbl := newTbl(doc, page)
	bodyW := col1W + 2 + col2W + 2 + col3W // full width
	tbl.SetWidths(28, bodyW-28)

	tbl.AddHeader(folio.TableCell{Text: "Guidelines", ColSpan: 2, Style: darkHdr, Align: "L"})

	rows := [][2]string{
		{"Occupancy", "Owner-Occupied, Non-Owner Occupied, Investment"},
		{"Property Types", "Office, Retail, Industrial, Warehouse, Mixed-Use, Multi-Family (5+ units)"},
		{"Cash Out", "Max Cash-Out = $2,000,000; Cash-Out > $1,000,000 requires DSCR >= 1.20 & LTV <= 55%"},
		{"Declining Markets", "If property is in a declining market per appraisal, Max LTV reduced by 5%"},
		{"Subordinate Liens", "Max CLTV = Grid Max LTV (institutional only); No private seconds"},
		{"Borrower Entity", "LLC, Corporation, or Trust required; personal guaranty from 25%+ owners"},
	}
	for _, r := range rows {
		tbl.AddRow(
			folio.TableCell{Text: r[0], Style: bodyB},
			folio.TableCell{Text: r[1], Style: bodyReg},
		)
	}

	// Appraisal — rowspan 2
	tbl.AddRow(
		folio.TableCell{Text: "Appraisal", Style: bodyB, RowSpan: 2},
		folio.TableCell{Text: "Full commercial appraisal with income approach required; MAI designated appraiser for loans > $1M", Style: bodyReg},
	)
	tbl.AddRow(folio.TableCell{Text: "Phase I Environmental required for all transactions; Phase II if recommended", Style: bodyReg})

	tbl.AddRow(
		folio.TableCell{Text: "Insurance", Style: bodyB},
		folio.TableCell{Text: "Hazard, liability, and business income coverage required; flood insurance if in SFHA", Style: bodyReg},
	)

	// Reserves — rowspan 2
	tbl.AddRow(
		folio.TableCell{Text: "Reserves", Style: bodyB, RowSpan: 2},
		folio.TableCell{Text: "6 months PITIA required for loans <= $2.5M; 12 months for loans > $2.5M; 18 months for non-owner occupied > $5M", Style: bodyReg},
	)
	tbl.AddRow(folio.TableCell{Text: "Replacement reserves may be required based on property condition report findings", Style: bodyReg})

	// Compliance — rowspan 2
	tbl.AddRow(
		folio.TableCell{Text: "Compliance", Style: bodyB, RowSpan: 2},
		folio.TableCell{Text: "All loans must comply with applicable federal, state, and local regulations including CRA requirements", Style: bodyReg},
	)
	tbl.AddRow(folio.TableCell{Text: "No predatory lending; all fees must be within state-specific caps", Style: bodyReg})

	tbl.Render()
}

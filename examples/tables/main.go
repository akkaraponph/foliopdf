// Example: Table Pattern Catalog
//
// Demonstrates every table pattern the Folio library can produce:
//
//   Page 1 — Simple Table & Comparison Table
//   Page 2 — Grouped Rows Table & Multi-Level Header Table
//   Page 3 — Matrix Table & Summary/Totals Table
//   Page 4 — Irregular Table (complex merges)
//   Page 5 — Visual Patterns (zebra, minimal borders, styled headers)
//
// Run from the repo root:
//
//	go run ./examples/tables
package main

import (
	"fmt"
	"os"

	"github.com/akkaraponph/presspdf"
)

const (
	lM = 10.0  // left margin
	tM = 10.0  // top margin
	bW = 190.0 // body width (A4 portrait: 210 - 2*10)
)

// Shared styles
var (
	hdr = &presspdf.CellStyle{
		FontFamily: "helvetica", FontStyle: "B", FontSize: 9,
		TextColor: [3]int{255, 255, 255},
		FillColor: [3]int{44, 62, 80},
		Fill:      true,
	}
	bodyBold = &presspdf.CellStyle{
		FontFamily: "helvetica", FontStyle: "B", FontSize: 8,
	}
	catStyle = &presspdf.CellStyle{
		FontFamily: "helvetica", FontStyle: "B", FontSize: 8,
		FillColor: [3]int{236, 240, 241},
		Fill:      true,
	}
	totalStyle = &presspdf.CellStyle{
		FontFamily: "helvetica", FontStyle: "B", FontSize: 8,
		FillColor: [3]int{44, 62, 80},
		TextColor: [3]int{255, 255, 255},
		Fill:      true,
	}
	subtotalStyle = &presspdf.CellStyle{
		FontFamily: "helvetica", FontStyle: "B", FontSize: 8,
		FillColor: [3]int{189, 195, 199},
		Fill:      true,
	}
	zebraEven = &presspdf.CellStyle{
		FontFamily: "helvetica", FontStyle: "", FontSize: 8,
		FillColor: [3]int{245, 247, 250},
		Fill:      true,
	}
	zebraOdd = &presspdf.CellStyle{
		FontFamily: "helvetica", FontStyle: "", FontSize: 8,
		FillColor: [3]int{255, 255, 255},
		Fill:      true,
	}
	accentHdr = &presspdf.CellStyle{
		FontFamily: "helvetica", FontStyle: "B", FontSize: 9,
		TextColor: [3]int{255, 255, 255},
		FillColor: [3]int{41, 128, 185},
		Fill:      true,
	}
	greenHdr = &presspdf.CellStyle{
		FontFamily: "helvetica", FontStyle: "B", FontSize: 9,
		TextColor: [3]int{255, 255, 255},
		FillColor: [3]int{39, 174, 96},
		Fill:      true,
	}
)

func main() {
	doc := presspdf.New(presspdf.WithUnit(presspdf.UnitMM))
	doc.SetTitle("Folio Table Pattern Catalog")
	doc.SetAuthor("Folio Library")
	doc.SetMargins(lM, tM, lM)

	page1Simple(doc)
	page2Grouped(doc)
	page3MatrixSummary(doc)
	page4Irregular(doc)
	page5Visual(doc)

	out := "/tmp/folio_table_patterns.pdf"
	if err := doc.Save(out); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Table pattern catalog saved to %s\n", out)
}

// ─── Page 1: Simple Table & Comparison Table ─────────────────────────

func page1Simple(doc *presspdf.Document) {
	p := doc.AddPage(presspdf.A4)
	sectionTitle(doc, p, "1. Simple Data Table")
	subtitle(doc, p, "One header row, one value per cell. Best for flat data, exports, CSV-like records.")

	tbl := presspdf.NewTable(doc, p)
	tbl.SetWidths(15, 50, 30, 30, 35, 30)
	tbl.SetRowHeight(7)
	tbl.SetCellPadding(1.5)
	tbl.SetLineHeight(5)
	tbl.SetBorder("1")

	tbl.AddHeader(
		presspdf.TableCell{Text: "ID", Style: hdr, Align: "C"},
		presspdf.TableCell{Text: "Employee", Style: hdr, Align: "L"},
		presspdf.TableCell{Text: "Department", Style: hdr, Align: "L"},
		presspdf.TableCell{Text: "Role", Style: hdr, Align: "L"},
		presspdf.TableCell{Text: "Start Date", Style: hdr, Align: "C"},
		presspdf.TableCell{Text: "Salary", Style: hdr, Align: "R"},
	)
	rows := [][6]string{
		{"001", "Alice Chen", "Engineering", "Senior Dev", "2021-03-15", "$142,000"},
		{"002", "Bob Martinez", "Design", "Lead Designer", "2020-07-01", "$128,000"},
		{"003", "Carol Nguyen", "Engineering", "Staff Eng", "2019-01-10", "$165,000"},
		{"004", "David Park", "Marketing", "Manager", "2022-06-20", "$115,000"},
		{"005", "Emma Wilson", "Engineering", "Junior Dev", "2023-09-01", "$95,000"},
		{"006", "Frank Osei", "Design", "UX Researcher", "2022-11-12", "$108,000"},
		{"007", "Grace Liu", "Marketing", "Analyst", "2023-02-28", "$88,000"},
	}
	for _, r := range rows {
		tbl.AddRow(
			presspdf.TableCell{Text: r[0], Align: "C"},
			presspdf.TableCell{Text: r[1]},
			presspdf.TableCell{Text: r[2]},
			presspdf.TableCell{Text: r[3]},
			presspdf.TableCell{Text: r[4], Align: "C"},
			presspdf.TableCell{Text: r[5], Align: "R"},
		)
	}
	tbl.Render()

	// --- Comparison Table ---
	p.SetY(p.GetY() + 12)
	sectionTitle(doc, p, "2. Comparison Table")
	subtitle(doc, p, "Features in rows, options in columns. Best for product comparisons, plan tiers.")

	tbl2 := presspdf.NewTable(doc, p)
	tbl2.SetWidths(55, 45, 45, 45)
	tbl2.SetRowHeight(7)
	tbl2.SetCellPadding(1.5)
	tbl2.SetLineHeight(5)
	tbl2.SetBorder("1")

	tbl2.AddHeader(
		presspdf.TableCell{Text: "Feature", Style: hdr, Align: "L"},
		presspdf.TableCell{Text: "Starter", Style: accentHdr, Align: "C"},
		presspdf.TableCell{Text: "Professional", Style: accentHdr, Align: "C"},
		presspdf.TableCell{Text: "Enterprise", Style: greenHdr, Align: "C"},
	)
	features := [][4]string{
		{"Monthly Price", "$9", "$29", "$99"},
		{"API Requests / day", "1,000", "50,000", "Unlimited"},
		{"Storage", "5 GB", "100 GB", "1 TB"},
		{"Team Members", "1", "10", "Unlimited"},
		{"Custom Domain", "No", "Yes", "Yes"},
		{"Priority Support", "No", "Email", "24/7 Phone"},
		{"SLA Guarantee", "-", "99.9%", "99.99%"},
		{"SSO / SAML", "No", "No", "Yes"},
		{"Audit Logs", "No", "30 days", "1 year"},
	}
	for _, f := range features {
		tbl2.AddRow(
			presspdf.TableCell{Text: f[0], Style: bodyBold},
			presspdf.TableCell{Text: f[1], Align: "C"},
			presspdf.TableCell{Text: f[2], Align: "C"},
			presspdf.TableCell{Text: f[3], Align: "C"},
		)
	}
	tbl2.Render()
}

// ─── Page 2: Grouped Rows & Multi-Level Headers ─────────────────────

func page2Grouped(doc *presspdf.Document) {
	p := doc.AddPage(presspdf.A4)
	sectionTitle(doc, p, "3. Grouped Rows Table (vertical merge)")
	subtitle(doc, p, "Repeated category values merged vertically. Best for hierarchical data, org charts.")

	tbl := presspdf.NewTable(doc, p)
	tbl.SetWidths(35, 45, 45, 35, 30)
	tbl.SetRowHeight(7)
	tbl.SetCellPadding(1.5)
	tbl.SetLineHeight(5)
	tbl.SetBorder("1")

	tbl.AddHeader(
		presspdf.TableCell{Text: "Department", Style: hdr, Align: "C"},
		presspdf.TableCell{Text: "Team", Style: hdr, Align: "L"},
		presspdf.TableCell{Text: "Project", Style: hdr, Align: "L"},
		presspdf.TableCell{Text: "Status", Style: hdr, Align: "C"},
		presspdf.TableCell{Text: "Budget", Style: hdr, Align: "R"},
	)

	tbl.AddRow(
		presspdf.TableCell{Text: "Engineering", Style: catStyle, RowSpan: 4, Align: "C"},
		presspdf.TableCell{Text: "Backend", RowSpan: 2},
		presspdf.TableCell{Text: "API v3 Migration"},
		presspdf.TableCell{Text: "In Progress", Align: "C"},
		presspdf.TableCell{Text: "$180,000", Align: "R"},
	)
	tbl.AddRow(
		presspdf.TableCell{Text: "Database Sharding"},
		presspdf.TableCell{Text: "Planning", Align: "C"},
		presspdf.TableCell{Text: "$250,000", Align: "R"},
	)
	tbl.AddRow(
		presspdf.TableCell{Text: "Frontend", RowSpan: 2},
		presspdf.TableCell{Text: "Design System v2"},
		presspdf.TableCell{Text: "Complete", Align: "C"},
		presspdf.TableCell{Text: "$90,000", Align: "R"},
	)
	tbl.AddRow(
		presspdf.TableCell{Text: "Performance Audit"},
		presspdf.TableCell{Text: "In Progress", Align: "C"},
		presspdf.TableCell{Text: "$45,000", Align: "R"},
	)
	tbl.AddRow(
		presspdf.TableCell{Text: "Marketing", Style: catStyle, RowSpan: 3, Align: "C"},
		presspdf.TableCell{Text: "Growth", RowSpan: 2},
		presspdf.TableCell{Text: "Q2 Campaign"},
		presspdf.TableCell{Text: "In Progress", Align: "C"},
		presspdf.TableCell{Text: "$120,000", Align: "R"},
	)
	tbl.AddRow(
		presspdf.TableCell{Text: "SEO Overhaul"},
		presspdf.TableCell{Text: "Planning", Align: "C"},
		presspdf.TableCell{Text: "$60,000", Align: "R"},
	)
	tbl.AddRow(
		presspdf.TableCell{Text: "Brand"},
		presspdf.TableCell{Text: "Rebranding Initiative"},
		presspdf.TableCell{Text: "Complete", Align: "C"},
		presspdf.TableCell{Text: "$200,000", Align: "R"},
	)
	tbl.AddRow(
		presspdf.TableCell{Text: "Operations", Style: catStyle, RowSpan: 2, Align: "C"},
		presspdf.TableCell{Text: "IT"},
		presspdf.TableCell{Text: "Cloud Migration"},
		presspdf.TableCell{Text: "In Progress", Align: "C"},
		presspdf.TableCell{Text: "$340,000", Align: "R"},
	)
	tbl.AddRow(
		presspdf.TableCell{Text: "HR"},
		presspdf.TableCell{Text: "Onboarding Revamp"},
		presspdf.TableCell{Text: "Planning", Align: "C"},
		presspdf.TableCell{Text: "$75,000", Align: "R"},
	)
	tbl.Render()

	// --- Multi-Level Header ---
	p.SetY(p.GetY() + 12)
	sectionTitle(doc, p, "4. Multi-Level Header Table")
	subtitle(doc, p, "Two+ header rows with column groups (colspan). Best for financial reports, schedules.")

	tbl2 := presspdf.NewTable(doc, p)
	tbl2.SetWidths(40, 25, 25, 25, 25, 25, 25)
	tbl2.SetRowHeight(7)
	tbl2.SetCellPadding(1.5)
	tbl2.SetLineHeight(5)
	tbl2.SetBorder("1")

	tbl2.AddHeader(
		presspdf.TableCell{Text: "", Style: hdr},
		presspdf.TableCell{Text: "2024", ColSpan: 3, Style: accentHdr, Align: "C"},
		presspdf.TableCell{Text: "2025", ColSpan: 3, Style: greenHdr, Align: "C"},
	)
	tbl2.AddHeader(
		presspdf.TableCell{Text: "Region", Style: hdr, Align: "L"},
		presspdf.TableCell{Text: "Q1", Style: accentHdr, Align: "C"},
		presspdf.TableCell{Text: "Q2", Style: accentHdr, Align: "C"},
		presspdf.TableCell{Text: "Q3", Style: accentHdr, Align: "C"},
		presspdf.TableCell{Text: "Q1", Style: greenHdr, Align: "C"},
		presspdf.TableCell{Text: "Q2", Style: greenHdr, Align: "C"},
		presspdf.TableCell{Text: "Q3", Style: greenHdr, Align: "C"},
	)
	regions := [][7]string{
		{"North America", "2.4M", "2.8M", "3.1M", "3.3M", "3.6M", "3.9M"},
		{"Europe", "1.8M", "2.0M", "2.2M", "2.3M", "2.5M", "2.7M"},
		{"Asia Pacific", "1.2M", "1.5M", "1.8M", "2.0M", "2.3M", "2.6M"},
		{"Latin America", "0.6M", "0.7M", "0.8M", "0.9M", "1.0M", "1.1M"},
		{"Middle East", "0.3M", "0.4M", "0.5M", "0.5M", "0.6M", "0.7M"},
	}
	for _, r := range regions {
		tbl2.AddRow(
			presspdf.TableCell{Text: r[0], Style: bodyBold},
			presspdf.TableCell{Text: r[1], Align: "R"},
			presspdf.TableCell{Text: r[2], Align: "R"},
			presspdf.TableCell{Text: r[3], Align: "R"},
			presspdf.TableCell{Text: r[4], Align: "R"},
			presspdf.TableCell{Text: r[5], Align: "R"},
			presspdf.TableCell{Text: r[6], Align: "R"},
		)
	}
	tbl2.Render()
}

// ─── Page 3: Matrix & Summary Tables ────────────────────────────────

func page3MatrixSummary(doc *presspdf.Document) {
	p := doc.AddPage(presspdf.A4)
	sectionTitle(doc, p, "5. Matrix Table")
	subtitle(doc, p, "Row dimensions on left, column dimensions on top, values in body. Best for schedules, cross-tabs.")

	tbl := presspdf.NewTable(doc, p)
	tbl.SetWidths(30, 22, 22, 22, 22, 22, 22, 22)
	tbl.SetRowHeight(7)
	tbl.SetCellPadding(1.5)
	tbl.SetLineHeight(5)
	tbl.SetBorder("1")

	tbl.AddHeader(
		presspdf.TableCell{Text: "Skill \\ Level", Style: hdr, Align: "C"},
		presspdf.TableCell{Text: "Mon", Style: hdr, Align: "C"},
		presspdf.TableCell{Text: "Tue", Style: hdr, Align: "C"},
		presspdf.TableCell{Text: "Wed", Style: hdr, Align: "C"},
		presspdf.TableCell{Text: "Thu", Style: hdr, Align: "C"},
		presspdf.TableCell{Text: "Fri", Style: hdr, Align: "C"},
		presspdf.TableCell{Text: "Sat", Style: hdr, Align: "C"},
		presspdf.TableCell{Text: "Sun", Style: hdr, Align: "C"},
	)
	matrix := [][8]string{
		{"Go", "4h", "3h", "4h", "2h", "4h", "1h", "-"},
		{"TypeScript", "2h", "4h", "2h", "3h", "2h", "2h", "-"},
		{"Python", "1h", "-", "2h", "3h", "1h", "3h", "2h"},
		{"Rust", "-", "1h", "-", "1h", "2h", "2h", "4h"},
		{"SQL", "1h", "1h", "1h", "1h", "1h", "-", "-"},
		{"DevOps", "-", "1h", "1h", "-", "-", "2h", "1h"},
	}
	for _, r := range matrix {
		tbl.AddRow(
			presspdf.TableCell{Text: r[0], Style: bodyBold, Align: "L"},
			presspdf.TableCell{Text: r[1], Align: "C"},
			presspdf.TableCell{Text: r[2], Align: "C"},
			presspdf.TableCell{Text: r[3], Align: "C"},
			presspdf.TableCell{Text: r[4], Align: "C"},
			presspdf.TableCell{Text: r[5], Align: "C"},
			presspdf.TableCell{Text: r[6], Align: "C"},
			presspdf.TableCell{Text: r[7], Align: "C"},
		)
	}
	tbl.Render()

	// --- Summary / Totals Table ---
	p.SetY(p.GetY() + 12)
	sectionTitle(doc, p, "6. Summary Table with Subtotals & Grand Total")
	subtitle(doc, p, "Includes category subtotals and grand total row. Best for financial statements, invoices.")

	tbl2 := presspdf.NewTable(doc, p)
	tbl2.SetWidths(50, 50, 30, 30, 30)
	tbl2.SetRowHeight(7)
	tbl2.SetCellPadding(1.5)
	tbl2.SetLineHeight(5)
	tbl2.SetBorder("1")

	tbl2.AddHeader(
		presspdf.TableCell{Text: "Category", Style: hdr, Align: "L"},
		presspdf.TableCell{Text: "Item", Style: hdr, Align: "L"},
		presspdf.TableCell{Text: "Qty", Style: hdr, Align: "R"},
		presspdf.TableCell{Text: "Unit Price", Style: hdr, Align: "R"},
		presspdf.TableCell{Text: "Amount", Style: hdr, Align: "R"},
	)

	// Hardware
	tbl2.AddRow(
		presspdf.TableCell{Text: "Hardware", Style: catStyle, RowSpan: 3},
		presspdf.TableCell{Text: "Laptop (M4 Pro)"},
		presspdf.TableCell{Text: "5", Align: "R"},
		presspdf.TableCell{Text: "$2,499", Align: "R"},
		presspdf.TableCell{Text: "$12,495", Align: "R"},
	)
	tbl2.AddRow(
		presspdf.TableCell{Text: "Monitor 4K 32\""},
		presspdf.TableCell{Text: "5", Align: "R"},
		presspdf.TableCell{Text: "$599", Align: "R"},
		presspdf.TableCell{Text: "$2,995", Align: "R"},
	)
	tbl2.AddRow(
		presspdf.TableCell{Text: "Keyboard + Mouse"},
		presspdf.TableCell{Text: "5", Align: "R"},
		presspdf.TableCell{Text: "$199", Align: "R"},
		presspdf.TableCell{Text: "$995", Align: "R"},
	)
	tbl2.AddRow(
		presspdf.TableCell{Text: "", ColSpan: 3, Style: subtotalStyle},
		presspdf.TableCell{Text: "Subtotal", Style: subtotalStyle, Align: "R"},
		presspdf.TableCell{Text: "$16,485", Style: subtotalStyle, Align: "R"},
	)

	// Software
	tbl2.AddRow(
		presspdf.TableCell{Text: "Software", Style: catStyle, RowSpan: 3},
		presspdf.TableCell{Text: "IDE License (annual)"},
		presspdf.TableCell{Text: "5", Align: "R"},
		presspdf.TableCell{Text: "$289", Align: "R"},
		presspdf.TableCell{Text: "$1,445", Align: "R"},
	)
	tbl2.AddRow(
		presspdf.TableCell{Text: "Cloud Platform"},
		presspdf.TableCell{Text: "1", Align: "R"},
		presspdf.TableCell{Text: "$4,800", Align: "R"},
		presspdf.TableCell{Text: "$4,800", Align: "R"},
	)
	tbl2.AddRow(
		presspdf.TableCell{Text: "Monitoring SaaS"},
		presspdf.TableCell{Text: "1", Align: "R"},
		presspdf.TableCell{Text: "$1,200", Align: "R"},
		presspdf.TableCell{Text: "$1,200", Align: "R"},
	)
	tbl2.AddRow(
		presspdf.TableCell{Text: "", ColSpan: 3, Style: subtotalStyle},
		presspdf.TableCell{Text: "Subtotal", Style: subtotalStyle, Align: "R"},
		presspdf.TableCell{Text: "$7,445", Style: subtotalStyle, Align: "R"},
	)

	// Grand total
	tbl2.AddRow(
		presspdf.TableCell{Text: "", ColSpan: 3, Style: totalStyle},
		presspdf.TableCell{Text: "Grand Total", Style: totalStyle, Align: "R"},
		presspdf.TableCell{Text: "$23,930", Style: totalStyle, Align: "R"},
	)
	tbl2.Render()
}

// ─── Page 4: Irregular Table ────────────────────────────────────────

func page4Irregular(doc *presspdf.Document) {
	p := doc.AddPage(presspdf.A4)
	sectionTitle(doc, p, "7. Irregular Table (complex merges)")
	subtitle(doc, p, "Not every cell maps to a flat header. Combines rowspan + colspan + per-cell styling.")

	tbl := presspdf.NewTable(doc, p)
	tbl.SetWidths(30, 30, 25, 25, 25, 25, 30)
	tbl.SetRowHeight(7)
	tbl.SetCellPadding(1.5)
	tbl.SetLineHeight(5)
	tbl.SetBorder("1")

	// Complex header: 3 rows deep
	tbl.AddHeader(
		presspdf.TableCell{Text: "Test Suite Results", ColSpan: 7, Align: "C", Style: hdr},
	)
	tbl.AddHeader(
		presspdf.TableCell{Text: "Module", Style: hdr, Align: "C", RowSpan: 2},
		presspdf.TableCell{Text: "Component", Style: hdr, Align: "C", RowSpan: 2},
		presspdf.TableCell{Text: "Unit Tests", ColSpan: 2, Style: accentHdr, Align: "C"},
		presspdf.TableCell{Text: "Integration", ColSpan: 2, Style: greenHdr, Align: "C"},
		presspdf.TableCell{Text: "Overall", Style: hdr, Align: "C", RowSpan: 2},
	)
	tbl.AddHeader(
		presspdf.TableCell{Text: "Pass", Style: accentHdr, Align: "C"},
		presspdf.TableCell{Text: "Fail", Style: accentHdr, Align: "C"},
		presspdf.TableCell{Text: "Pass", Style: greenHdr, Align: "C"},
		presspdf.TableCell{Text: "Fail", Style: greenHdr, Align: "C"},
	)

	// Body with mixed rowspan
	tbl.AddRow(
		presspdf.TableCell{Text: "Auth", Style: catStyle, RowSpan: 3, Align: "C"},
		presspdf.TableCell{Text: "Login"},
		presspdf.TableCell{Text: "42", Align: "C"},
		presspdf.TableCell{Text: "0", Align: "C"},
		presspdf.TableCell{Text: "8", Align: "C"},
		presspdf.TableCell{Text: "1", Align: "C"},
		presspdf.TableCell{Text: "98%", Align: "C"},
	)
	tbl.AddRow(
		presspdf.TableCell{Text: "OAuth"},
		presspdf.TableCell{Text: "28", Align: "C"},
		presspdf.TableCell{Text: "2", Align: "C"},
		presspdf.TableCell{Text: "5", Align: "C"},
		presspdf.TableCell{Text: "0", Align: "C"},
		presspdf.TableCell{Text: "94%", Align: "C"},
	)
	tbl.AddRow(
		presspdf.TableCell{Text: "2FA"},
		presspdf.TableCell{Text: "15", Align: "C"},
		presspdf.TableCell{Text: "1", Align: "C"},
		presspdf.TableCell{Text: "3", Align: "C"},
		presspdf.TableCell{Text: "0", Align: "C"},
		presspdf.TableCell{Text: "95%", Align: "C"},
	)
	tbl.AddRow(
		presspdf.TableCell{Text: "API", Style: catStyle, RowSpan: 2, Align: "C"},
		presspdf.TableCell{Text: "REST"},
		presspdf.TableCell{Text: "87", Align: "C"},
		presspdf.TableCell{Text: "3", Align: "C"},
		presspdf.TableCell{Text: "12", Align: "C"},
		presspdf.TableCell{Text: "1", Align: "C"},
		presspdf.TableCell{Text: "96%", Align: "C"},
	)
	tbl.AddRow(
		presspdf.TableCell{Text: "GraphQL"},
		presspdf.TableCell{Text: "64", Align: "C"},
		presspdf.TableCell{Text: "5", Align: "C"},
		presspdf.TableCell{Text: "9", Align: "C"},
		presspdf.TableCell{Text: "2", Align: "C"},
		presspdf.TableCell{Text: "91%", Align: "C"},
	)
	tbl.AddRow(
		presspdf.TableCell{Text: "Storage", Style: catStyle, Align: "C"},
		presspdf.TableCell{Text: "S3 + DB"},
		presspdf.TableCell{Text: "53", Align: "C"},
		presspdf.TableCell{Text: "0", Align: "C"},
		presspdf.TableCell{Text: "7", Align: "C"},
		presspdf.TableCell{Text: "0", Align: "C"},
		presspdf.TableCell{Text: "100%", Align: "C"},
	)

	// Full-width summary row
	tbl.AddRow(
		presspdf.TableCell{Text: "Total: 289 passed, 11 failed, 44 integration - 96.3% overall", ColSpan: 7, Style: totalStyle, Align: "C"},
	)
	tbl.Render()

	// --- Second irregular: mixed colspan in body ---
	p.SetY(p.GetY() + 12)
	sectionTitle(doc, p, "8. Irregular Table - Mixed Body Merges")
	subtitle(doc, p, "Full-width note rows, section breaks, and body-level colspan for annotations.")

	tbl2 := presspdf.NewTable(doc, p)
	tbl2.SetWidths(38, 38, 38, 38, 38)
	tbl2.SetRowHeight(7)
	tbl2.SetCellPadding(1.5)
	tbl2.SetLineHeight(5)
	tbl2.SetBorder("1")

	tbl2.AddHeader(
		presspdf.TableCell{Text: "Server", Style: hdr, Align: "C"},
		presspdf.TableCell{Text: "Region", Style: hdr, Align: "C"},
		presspdf.TableCell{Text: "CPU", Style: hdr, Align: "C"},
		presspdf.TableCell{Text: "Memory", Style: hdr, Align: "C"},
		presspdf.TableCell{Text: "Status", Style: hdr, Align: "C"},
	)

	// Section header row
	tbl2.AddRow(presspdf.TableCell{Text: "Production Cluster", ColSpan: 5, Style: catStyle, Align: "L"})
	tbl2.AddRow(
		presspdf.TableCell{Text: "web-prod-01"},
		presspdf.TableCell{Text: "us-east-1", Align: "C"},
		presspdf.TableCell{Text: "72%", Align: "C"},
		presspdf.TableCell{Text: "84%", Align: "C"},
		presspdf.TableCell{Text: "Healthy", Align: "C"},
	)
	tbl2.AddRow(
		presspdf.TableCell{Text: "web-prod-02"},
		presspdf.TableCell{Text: "us-east-1", Align: "C"},
		presspdf.TableCell{Text: "68%", Align: "C"},
		presspdf.TableCell{Text: "79%", Align: "C"},
		presspdf.TableCell{Text: "Healthy", Align: "C"},
	)
	tbl2.AddRow(
		presspdf.TableCell{Text: "web-prod-03"},
		presspdf.TableCell{Text: "eu-west-1", Align: "C"},
		presspdf.TableCell{Text: "45%", Align: "C"},
		presspdf.TableCell{Text: "62%", Align: "C"},
		presspdf.TableCell{Text: "Healthy", Align: "C"},
	)

	// Section header row
	tbl2.AddRow(presspdf.TableCell{Text: "Staging Cluster", ColSpan: 5, Style: catStyle, Align: "L"})
	tbl2.AddRow(
		presspdf.TableCell{Text: "web-stg-01"},
		presspdf.TableCell{Text: "us-west-2", Align: "C"},
		presspdf.TableCell{Text: "12%", Align: "C"},
		presspdf.TableCell{Text: "34%", Align: "C"},
		presspdf.TableCell{Text: "Healthy", Align: "C"},
	)
	tbl2.AddRow(
		presspdf.TableCell{Text: "web-stg-02"},
		presspdf.TableCell{Text: "us-west-2", Align: "C"},
		presspdf.TableCell{Text: "8%", Align: "C"},
		presspdf.TableCell{Text: "28%", Align: "C"},
		presspdf.TableCell{Text: "Degraded", Align: "C"},
	)

	// Full-width annotation
	tbl2.AddRow(presspdf.TableCell{Text: "Note: web-stg-02 disk I/O latency elevated since 2026-04-12 03:22 UTC. Investigation ongoing.", ColSpan: 5, Style: bodyBold, Align: "L"})
	tbl2.Render()
}

// ─── Page 5: Visual Patterns ────────────────────────────────────────

func page5Visual(doc *presspdf.Document) {
	p := doc.AddPage(presspdf.A4)
	sectionTitle(doc, p, "9. Visual Patterns - Zebra Striping")
	subtitle(doc, p, "Alternating row colors improve scanability for long datasets.")

	tbl := presspdf.NewTable(doc, p)
	tbl.SetWidths(15, 50, 40, 40, 45)
	tbl.SetRowHeight(7)
	tbl.SetCellPadding(1.5)
	tbl.SetLineHeight(5)
	tbl.SetBorder("1")

	tbl.AddHeader(
		presspdf.TableCell{Text: "#", Style: hdr, Align: "C"},
		presspdf.TableCell{Text: "Transaction", Style: hdr, Align: "L"},
		presspdf.TableCell{Text: "Date", Style: hdr, Align: "C"},
		presspdf.TableCell{Text: "Category", Style: hdr, Align: "L"},
		presspdf.TableCell{Text: "Amount", Style: hdr, Align: "R"},
	)

	txns := [][5]string{
		{"1", "Office Supplies", "2026-04-01", "Operations", "$234.50"},
		{"2", "AWS Monthly", "2026-04-01", "Infrastructure", "$8,412.00"},
		{"3", "Team Lunch", "2026-04-03", "Team Building", "$187.25"},
		{"4", "Conference Tickets", "2026-04-05", "Education", "$1,500.00"},
		{"5", "Figma License", "2026-04-07", "Software", "$540.00"},
		{"6", "Uber for Clients", "2026-04-08", "Travel", "$94.30"},
		{"7", "Datadog", "2026-04-10", "Infrastructure", "$2,100.00"},
		{"8", "New Hire Equipment", "2026-04-11", "Hardware", "$3,298.00"},
		{"9", "Legal Review", "2026-04-12", "Professional", "$4,500.00"},
		{"10", "Marketing Ads", "2026-04-13", "Marketing", "$6,250.00"},
	}
	for i, t := range txns {
		style := zebraEven
		if i%2 == 1 {
			style = zebraOdd
		}
		tbl.AddRow(
			presspdf.TableCell{Text: t[0], Style: style, Align: "C"},
			presspdf.TableCell{Text: t[1], Style: style},
			presspdf.TableCell{Text: t[2], Style: style, Align: "C"},
			presspdf.TableCell{Text: t[3], Style: style},
			presspdf.TableCell{Text: t[4], Style: style, Align: "R"},
		)
	}
	tbl.Render()

	// --- Minimal borders ---
	p.SetY(p.GetY() + 12)
	sectionTitle(doc, p, "10. Minimal Borders - Horizontal Lines Only")
	subtitle(doc, p, "Reduced visual noise. Best for clean reports and dashboards.")

	tbl2 := presspdf.NewTable(doc, p)
	tbl2.SetWidths(50, 40, 40, 30, 30)
	tbl2.SetRowHeight(7)
	tbl2.SetCellPadding(1.5)
	tbl2.SetLineHeight(5)
	tbl2.SetBorder("TB") // top + bottom borders only

	tbl2.AddHeader(
		presspdf.TableCell{Text: "Metric", Style: hdr, Align: "L"},
		presspdf.TableCell{Text: "Current", Style: hdr, Align: "R"},
		presspdf.TableCell{Text: "Previous", Style: hdr, Align: "R"},
		presspdf.TableCell{Text: "Change", Style: hdr, Align: "R"},
		presspdf.TableCell{Text: "Target", Style: hdr, Align: "R"},
	)

	metrics := [][5]string{
		{"Revenue", "$4.2M", "$3.8M", "+10.5%", "$4.5M"},
		{"Active Users", "148,200", "132,500", "+11.8%", "150,000"},
		{"Churn Rate", "2.1%", "2.8%", "-0.7%", "< 2.5%"},
		{"NPS Score", "72", "68", "+4", "> 70"},
		{"Avg Response Time", "142ms", "189ms", "-24.9%", "< 200ms"},
		{"Uptime", "99.97%", "99.92%", "+0.05%", "> 99.9%"},
	}
	for _, m := range metrics {
		tbl2.AddRow(
			presspdf.TableCell{Text: m[0], Style: bodyBold},
			presspdf.TableCell{Text: m[1], Align: "R"},
			presspdf.TableCell{Text: m[2], Align: "R"},
			presspdf.TableCell{Text: m[3], Align: "R"},
			presspdf.TableCell{Text: m[4], Align: "R"},
		)
	}
	tbl2.Render()
}

// --- Helpers ---

func sectionTitle(doc *presspdf.Document, p *presspdf.Page, title string) {
	doc.SetFont("helvetica", "B", 14)
	doc.SetTextColor(44, 62, 80)
	p.SetX(lM)
	p.Cell(bW, 10, title, "", "L", false, 1)
}

func subtitle(doc *presspdf.Document, p *presspdf.Page, text string) {
	doc.SetFont("helvetica", "", 9)
	doc.SetTextColor(127, 140, 141)
	p.SetX(lM)
	p.Cell(bW, 6, text, "", "L", false, 1)
	p.SetY(p.GetY() + 2)
	// Reset text color for tables
	doc.SetTextColor(0, 0, 0)
}

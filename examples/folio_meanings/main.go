// Example: "Folio" ในบริบทเศรษฐกิจและธุรกิจ
//
// Demonstrates a multi-page bilingual (Thai/English) document covering the
// various business meanings of the word "Folio":
//   - Guest Folio (hospitality)
//   - Folio Number (investment/finance)
//   - Folio Brand (leather goods)
//   - Other uses (FOLIO library system, iFolio portfolio)
//
// Run from the repo root:
//
//	go run ./examples/folio_meanings
package main

import (
	"fmt"
	"os"

	"github.com/akkaraponph/folio"
	"github.com/akkaraponph/folio/fonts/sarabun"
	"github.com/akkaraponph/folio/thai"
)

// --- Page geometry (A4, mm) ---

const (
	pageW   = 210.0
	pageH   = 297.0
	lMargin = 20.0
	tMargin = 15.0
	rMargin = 20.0
)

// --- Palette ---

var (
	navy      = [3]int{17, 38, 77}
	darkTeal  = [3]int{0, 102, 102}
	accent    = [3]int{0, 120, 180}
	white     = [3]int{255, 255, 255}
	bodyText  = [3]int{30, 30, 35}
	mutedText = [3]int{100, 100, 110}
	rule      = [3]int{200, 200, 210}
	zebra     = [3]int{245, 247, 252}
	lightBlue = [3]int{230, 242, 255}
	lightGold = [3]int{255, 248, 230}
	lightGreen = [3]int{230, 248, 240}
)

type writer struct {
	doc       *folio.Document
	page      *folio.Page
	bodyWidth float64
}

func (w *writer) newPage() {
	w.page = w.doc.AddPage(folio.A4)
	w.page.SetXY(lMargin, tMargin)
}


func main() {
	doc := folio.New(folio.WithCompression(true))
	doc.SetTitle("ความหมายของ Folio ในบริบทเศรษฐกิจและธุรกิจ")
	doc.SetAuthor("Folio Library Examples")
	doc.SetCreator("folio examples/folio_meanings")
	doc.SetMargins(lMargin, tMargin, rMargin)

	if err := sarabun.Register(doc); err != nil {
		fail("register sarabun: %v", err)
	}
	thai.Setup(doc)

	w := &writer{
		doc:       doc,
		bodyWidth: pageW - lMargin - rMargin,
	}

	// --- Page 1: Cover / Title ---
	w.drawCover()

	// --- Page 2: Guest Folio ---
	w.drawGuestFolio()

	// --- Page 3: Folio Number (Investment) ---
	w.drawFolioNumber()

	// --- Page 4: Folio Brand + Others ---
	w.drawFolioBrandAndOthers()

	out := "/tmp/folio_meanings.pdf"
	if err := doc.Save(out); err != nil {
		fail("save: %v", err)
	}
	fmt.Printf("PDF saved to %s\n", out)
}

// =====================================================================
// Cover page
// =====================================================================

func (w *writer) drawCover() {
	w.newPage()
	p := w.page

	// Navy header bar
	w.doc.SetFillColor(navy[0], navy[1], navy[2])
	p.Rect(0, 0, pageW, 70, "F")

	// Title in header
	w.doc.SetTextColor(white[0], white[1], white[2])
	w.doc.SetFont("sarabun", "B", 26)
	p.SetXY(lMargin, 18)
	p.MultiCell(w.bodyWidth, 11, "ความหมายของ \"Folio\"\nในบริบทเศรษฐกิจและธุรกิจ", "", "C", false)

	w.doc.SetFont("sarabun", "", 13)
	p.SetXY(lMargin, 48)
	p.Cell(w.bodyWidth, 7, "The Many Meanings of \"Folio\" in Business & Economy", "", "C", false, 1)

	// Decorative line below header
	w.doc.SetDrawColor(accent[0], accent[1], accent[2])
	w.doc.SetLineWidth(1.5)
	p.Line(lMargin, 74, lMargin+w.bodyWidth, 74)

	// Intro paragraph
	p.SetY(85)
	w.doc.SetTextColor(bodyText[0], bodyText[1], bodyText[2])
	w.doc.SetFont("sarabun", "", 12)
	p.SetX(lMargin)
	p.MultiCell(w.bodyWidth, 6.5,
		"คำว่า \"Folio\" มีรากศัพท์มาจากภาษาละติน แปลว่า \"ใบ\" หรือ \"แผ่น\" "+
			"ในโลกธุรกิจและเศรษฐกิจสมัยใหม่ คำนี้ถูกนำมาใช้ในหลายบริบทที่แตกต่างกัน "+
			"ตั้งแต่อุตสาหกรรมโรงแรม การเงินการลงทุน ไปจนถึงแบรนด์สินค้า "+
			"เอกสารนี้จะอธิบายความหมายที่สำคัญพร้อมตัวอย่างที่เป็นรูปธรรม",
		"", "J", false)

	// Overview boxes
	p.SetY(p.GetY() + 8)
	sections := []struct {
		title string
		desc  string
		bg    [3]int
	}{
		{"1. โรงแรมและที่พัก (Guest Folio)",
			"เอกสารแสดงรายการใช้จ่ายทั้งหมดของแขกในโรงแรม ใช้เรียกเก็บเงินตอนเช็คเอาท์",
			lightBlue},
		{"2. การเงิน/การลงทุน (Folio Number)",
			"เลขที่บัญชีที่ระบุตัวตนนักลงทุนในกองทุนรวม เพื่อดูประวัติการทำธุรกรรม",
			lightGold},
		{"3. แบรนด์สินค้า (Folio Brand)",
			"ร้านค้าที่ผลิตสมุดบันทึก ไดอารี่ ออกาไนเซอร์หนัง และกระเป๋าหนัง",
			lightGreen},
		{"4. อื่นๆ",
			"ระบบบริหารจัดการห้องสมุด (FOLIO), แฟ้มสะสมผลงานออนไลน์ (iFolio)",
			zebra},
	}

	for _, s := range sections {
		y := p.GetY()
		w.doc.SetFillColor(s.bg[0], s.bg[1], s.bg[2])
		p.Rect(lMargin, y, w.bodyWidth, 22, "F")
		w.doc.SetDrawColor(rule[0], rule[1], rule[2])
		p.Rect(lMargin, y, w.bodyWidth, 22, "D")

		w.doc.SetFont("sarabun", "B", 11)
		w.doc.SetTextColor(navy[0], navy[1], navy[2])
		p.SetXY(lMargin+4, y+2)
		p.Cell(w.bodyWidth-8, 6, s.title, "", "L", false, 1)

		w.doc.SetFont("sarabun", "", 10)
		w.doc.SetTextColor(bodyText[0], bodyText[1], bodyText[2])
		p.SetX(lMargin + 4)
		p.MultiCell(w.bodyWidth-8, 5, s.desc, "", "L", false)

		p.SetY(y + 24)
	}

	// Footer
	w.doc.SetFont("sarabun", "", 9)
	w.doc.SetTextColor(mutedText[0], mutedText[1], mutedText[2])
	p.SetXY(lMargin, pageH-15)
	p.Cell(w.bodyWidth, 5, "สร้างโดย Folio PDF Library — github.com/akkaraponph/folio", "", "C", false, 0)
}

// =====================================================================
// Guest Folio (Hospitality)
// =====================================================================

func (w *writer) drawGuestFolio() {
	w.newPage()
	p := w.page

	// Section header
	w.sectionHeader("1", "โรงแรมและที่พัก — Guest Folio", "Hotel & Hospitality")

	// Explanation
	w.doc.SetFont("sarabun", "", 11)
	w.doc.SetTextColor(bodyText[0], bodyText[1], bodyText[2])
	p.SetX(lMargin)
	p.MultiCell(w.bodyWidth, 5.8,
		"Guest Folio (เกสต์ โฟลิโอ) คือเอกสารสรุปค่าใช้จ่ายทั้งหมดของผู้เข้าพัก "+
			"ในโรงแรมหรือรีสอร์ท ครอบคลุมตั้งแต่ค่าห้องพัก ค่าอาหารและเครื่องดื่ม "+
			"ค่าบริการซักรีด มินิบาร์ สปา ไปจนถึงค่าโทรศัพท์ "+
			"โดยระบบ PMS (Property Management System) จะบันทึกรายการโดยอัตโนมัติ "+
			"และพิมพ์ออกมาเป็น Folio เมื่อแขกเช็คเอาท์",
		"", "J", false)

	p.SetY(p.GetY() + 6)

	// Example: Hotel Guest Folio
	w.doc.SetFont("sarabun", "B", 13)
	w.doc.SetTextColor(navy[0], navy[1], navy[2])
	p.SetX(lMargin)
	p.Cell(w.bodyWidth, 7, "ตัวอย่าง: Guest Folio — โรงแรมโฟลิโอ แกรนด์", "", "C", false, 1)

	// Hotel info box
	y := p.GetY() + 2
	w.doc.SetFillColor(navy[0], navy[1], navy[2])
	p.Rect(lMargin, y, w.bodyWidth, 18, "F")
	w.doc.SetTextColor(white[0], white[1], white[2])
	w.doc.SetFont("sarabun", "B", 14)
	p.SetXY(lMargin, y+2)
	p.Cell(w.bodyWidth, 7, "FOLIO GRAND HOTEL & RESORT", "", "C", false, 1)
	w.doc.SetFont("sarabun", "", 9)
	p.SetX(lMargin)
	p.Cell(w.bodyWidth, 5, "123 ถนนสุขุมวิท กรุงเทพฯ 10110 | โทร. 02-123-4567 | www.foliogrand.example.com", "", "C", false, 1)

	p.SetY(y + 22)

	// Guest details (two columns)
	w.doc.SetTextColor(bodyText[0], bodyText[1], bodyText[2])
	w.doc.SetFont("sarabun", "", 10)
	leftCol := []struct{ label, value string }{
		{"Folio No.", "F-2026-04521"},
		{"Guest Name", "คุณสมชาย รักไทย / Mr. Somchai Rakthai"},
		{"Room No.", "1205 (Deluxe River View)"},
	}
	rightCol := []struct{ label, value string }{
		{"Check-in", "09 เม.ย. 2026"},
		{"Check-out", "12 เม.ย. 2026 (3 คืน)"},
		{"Payment", "VISA **** 4532"},
	}

	startY := p.GetY()
	halfW := w.bodyWidth / 2
	for _, r := range leftCol {
		p.SetX(lMargin)
		w.doc.SetFont("sarabun", "B", 10)
		p.Cell(28, 5.5, r.label+":", "", "L", false, 0)
		w.doc.SetFont("sarabun", "", 10)
		p.Cell(halfW-28, 5.5, r.value, "", "L", false, 1)
	}
	endLeftY := p.GetY()

	p.SetY(startY)
	for _, r := range rightCol {
		p.SetX(lMargin + halfW)
		w.doc.SetFont("sarabun", "B", 10)
		p.Cell(24, 5.5, r.label+":", "", "L", false, 0)
		w.doc.SetFont("sarabun", "", 10)
		p.Cell(halfW-24, 5.5, r.value, "", "L", false, 1)
	}
	if p.GetY() < endLeftY {
		p.SetY(endLeftY)
	}

	// Thin rule
	p.SetY(p.GetY() + 2)
	w.doc.SetDrawColor(rule[0], rule[1], rule[2])
	w.doc.SetLineWidth(0.3)
	p.Line(lMargin, p.GetY(), lMargin+w.bodyWidth, p.GetY())
	p.SetY(p.GetY() + 3)

	// Charges table
	type charge struct {
		date   string
		desc   string
		ref    string
		amount float64
	}
	charges := []charge{
		{"09/04", "ค่าห้องพัก Deluxe River View", "RM", 4500.00},
		{"09/04", "อาหารเย็น — ห้องอาหาร The Terrace", "FB", 1850.00},
		{"10/04", "ค่าห้องพัก Deluxe River View", "RM", 4500.00},
		{"10/04", "บริการซักรีด", "LN", 350.00},
		{"10/04", "มินิบาร์", "MB", 480.00},
		{"10/04", "สปา — นวดแผนไทย 90 นาที", "SP", 1200.00},
		{"11/04", "ค่าห้องพัก Deluxe River View", "RM", 4500.00},
		{"11/04", "อาหารเช้า Room Service", "FB", 650.00},
		{"11/04", "ค่าโทรศัพท์ (ทางไกล)", "TL", 120.00},
	}

	colW := []float64{20, 85, 15, 30, 20}
	headers := []string{"วันที่", "รายการ", "รหัส", "จำนวนเงิน (฿)", ""}
	aligns := []string{"C", "L", "C", "R", "C"}

	// Header row
	w.doc.SetFont("sarabun", "B", 9)
	w.doc.SetFillColor(darkTeal[0], darkTeal[1], darkTeal[2])
	w.doc.SetTextColor(white[0], white[1], white[2])
	w.doc.SetDrawColor(darkTeal[0], darkTeal[1], darkTeal[2])
	p.SetX(lMargin)
	for i := 0; i < 4; i++ {
		p.Cell(colW[i], 7, headers[i], "1", aligns[i], true, 0)
	}
	p.SetY(p.GetY() + 7)

	// Body rows
	w.doc.SetFont("sarabun", "", 9)
	w.doc.SetTextColor(bodyText[0], bodyText[1], bodyText[2])
	w.doc.SetDrawColor(rule[0], rule[1], rule[2])

	var total float64
	for i, c := range charges {
		if i%2 == 0 {
			w.doc.SetFillColor(zebra[0], zebra[1], zebra[2])
		} else {
			w.doc.SetFillColor(white[0], white[1], white[2])
		}
		p.SetX(lMargin)
		p.Cell(colW[0], 6, c.date, "1", "C", true, 0)
		p.Cell(colW[1], 6, c.desc, "1", "L", true, 0)
		p.Cell(colW[2], 6, c.ref, "1", "C", true, 0)
		p.Cell(colW[3], 6, fmtNum(c.amount), "1", "R", true, 0)
		p.SetY(p.GetY() + 6)
		total += c.amount
	}

	// Totals
	vat := total * 0.07
	svc := total * 0.10
	grand := total + vat + svc

	p.SetY(p.GetY() + 2)
	totalsX := lMargin + colW[0] + colW[1]
	summaryW := colW[2] + colW[3]

	for _, row := range []struct {
		label string
		value float64
		bold  bool
	}{
		{"รวมค่าใช้จ่าย / Subtotal", total, false},
		{"ค่าบริการ 10% / Service Charge", svc, false},
		{"ภาษีมูลค่าเพิ่ม 7% / VAT", vat, false},
		{"ยอดรวมทั้งสิ้น / Grand Total", grand, true},
	} {
		if row.bold {
			w.doc.SetFont("sarabun", "B", 10)
			w.doc.SetFillColor(navy[0], navy[1], navy[2])
			w.doc.SetTextColor(white[0], white[1], white[2])
			w.doc.SetDrawColor(navy[0], navy[1], navy[2])
		} else {
			w.doc.SetFont("sarabun", "", 9)
			w.doc.SetFillColor(zebra[0], zebra[1], zebra[2])
			w.doc.SetTextColor(bodyText[0], bodyText[1], bodyText[2])
			w.doc.SetDrawColor(rule[0], rule[1], rule[2])
		}
		p.SetX(totalsX)
		p.Cell(summaryW-30, 6.5, row.label, "1", "L", true, 0)
		p.Cell(30, 6.5, fmtNum(row.value), "1", "R", true, 0)
		p.SetY(p.GetY() + 6.5)
	}

	// Reset colors
	w.doc.SetTextColor(bodyText[0], bodyText[1], bodyText[2])
	p.SetY(p.GetY() + 4)
	w.doc.SetFont("sarabun", "", 9)
	w.doc.SetTextColor(mutedText[0], mutedText[1], mutedText[2])
	p.SetX(lMargin)
	p.Cell(w.bodyWidth, 5, "** RM=Room, FB=Food&Beverage, LN=Laundry, MB=Minibar, SP=Spa, TL=Telephone", "", "L", false, 1)
	p.SetX(lMargin)
	p.Cell(w.bodyWidth, 5, "ขอบคุณที่เข้าพักกับเรา — Thank you for staying with us!", "", "C", false, 0)

	w.pageFooter()
}

// =====================================================================
// Folio Number (Investment / Finance)
// =====================================================================

func (w *writer) drawFolioNumber() {
	w.newPage()
	p := w.page

	w.sectionHeader("2", "การเงิน/การลงทุน — Folio Number", "Investment & Finance")

	w.doc.SetFont("sarabun", "", 11)
	w.doc.SetTextColor(bodyText[0], bodyText[1], bodyText[2])
	p.SetX(lMargin)
	p.MultiCell(w.bodyWidth, 5.8,
		"Folio Number (โฟลิโอ นัมเบอร์) คือเลขที่บัญชีเฉพาะที่บริษัทหลักทรัพย์จัดการกองทุน (บลจ.) "+
			"กำหนดให้แก่ผู้ถือหน่วยลงทุนแต่ละราย เพื่อใช้ในการระบุตัวตนและติดตามธุรกรรม "+
			"เช่น การซื้อ ขาย สับเปลี่ยน หรือรับเงินปันผล "+
			"นักลงทุนหนึ่งคนอาจมีหลาย Folio Number สำหรับกองทุนต่างๆ ที่ลงทุน",
		"", "J", false)

	p.SetY(p.GetY() + 6)

	// Example: Investment statement
	w.doc.SetFont("sarabun", "B", 13)
	w.doc.SetTextColor(navy[0], navy[1], navy[2])
	p.SetX(lMargin)
	p.Cell(w.bodyWidth, 7, "ตัวอย่าง: รายงานพอร์ตการลงทุน", "", "C", false, 1)

	// Fund house header
	y := p.GetY() + 2
	w.doc.SetFillColor(accent[0], accent[1], accent[2])
	p.Rect(lMargin, y, w.bodyWidth, 14, "F")
	w.doc.SetTextColor(white[0], white[1], white[2])
	w.doc.SetFont("sarabun", "B", 13)
	p.SetXY(lMargin, y+1)
	p.Cell(w.bodyWidth, 6, "FOLIO ASSET MANAGEMENT CO., LTD.", "", "C", false, 1)
	w.doc.SetFont("sarabun", "", 9)
	p.SetX(lMargin)
	p.Cell(w.bodyWidth, 5, "บลจ. โฟลิโอ จำกัด — ใบแจ้งยอดหน่วยลงทุน / Portfolio Statement", "", "C", false, 1)

	p.SetY(y + 18)

	// Investor info
	w.doc.SetTextColor(bodyText[0], bodyText[1], bodyText[2])
	w.doc.SetFont("sarabun", "B", 10)
	p.SetX(lMargin)
	p.Cell(30, 5.5, "ชื่อผู้ลงทุน:", "", "L", false, 0)
	w.doc.SetFont("sarabun", "", 10)
	p.Cell(80, 5.5, "นางสาวพิมพ์ใจ เจริญสุข", "", "L", false, 0)

	w.doc.SetFont("sarabun", "B", 10)
	p.Cell(25, 5.5, "วันที่:", "", "L", false, 0)
	w.doc.SetFont("sarabun", "", 10)
	p.Cell(35, 5.5, "31 มี.ค. 2026", "", "L", false, 1)

	p.SetY(p.GetY() + 3)

	// Portfolio table
	type fund struct {
		folioNo string
		name    string
		units   string
		nav     string
		value   string
		gain    string
	}
	funds := []fund{
		{"FN-001234", "กองทุนเปิด โฟลิโอ หุ้นระยะยาว (FEQLTF)", "12,450.3821", "15.6789", "195,202.45", "+8.52%"},
		{"FN-001235", "กองทุนเปิด โฟลิโอ ตราสารหนี้ (FFIX)", "50,000.0000", "10.2345", "511,725.00", "+2.35%"},
		{"FN-001236", "กองทุนเปิด โฟลิโอ ผสม (FMIX)", "8,200.5000", "22.4567", "184,161.11", "+5.10%"},
		{"FN-001237", "กองทุนเปิด โฟลิโอ ทอง (FGOLD)", "3,000.0000", "8.9012", "26,703.60", "-1.20%"},
	}

	fColW := []float64{22, 68, 24, 18, 22, 16}
	fHeaders := []string{"Folio No.", "ชื่อกองทุน", "หน่วย", "NAV", "มูลค่า (฿)", "กำไร/ขาดทุน"}

	w.doc.SetFont("sarabun", "B", 8)
	w.doc.SetFillColor(navy[0], navy[1], navy[2])
	w.doc.SetTextColor(white[0], white[1], white[2])
	w.doc.SetDrawColor(navy[0], navy[1], navy[2])
	p.SetX(lMargin)
	for _, h := range fHeaders {
		p.Cell(fColW[0], 7, h, "1", "C", true, 0)
		fColW = fColW[1:]
	}
	p.SetY(p.GetY() + 7)

	// Reset column widths
	fColW = []float64{22, 68, 24, 18, 22, 16}

	w.doc.SetFont("sarabun", "", 8)
	w.doc.SetTextColor(bodyText[0], bodyText[1], bodyText[2])
	w.doc.SetDrawColor(rule[0], rule[1], rule[2])

	for i, f := range funds {
		if i%2 == 0 {
			w.doc.SetFillColor(zebra[0], zebra[1], zebra[2])
		} else {
			w.doc.SetFillColor(white[0], white[1], white[2])
		}
		p.SetX(lMargin)
		vals := []string{f.folioNo, f.name, f.units, f.nav, f.value, f.gain}
		als := []string{"C", "L", "R", "R", "R", "R"}
		for j, v := range vals {
			p.Cell(fColW[j], 6.5, v, "1", als[j], true, 0)
		}
		p.SetY(p.GetY() + 6.5)
	}

	// Portfolio total
	p.SetY(p.GetY() + 2)
	w.doc.SetFont("sarabun", "B", 10)
	w.doc.SetFillColor(lightGold[0], lightGold[1], lightGold[2])
	w.doc.SetTextColor(navy[0], navy[1], navy[2])
	w.doc.SetDrawColor(rule[0], rule[1], rule[2])
	p.SetX(lMargin)
	p.Cell(114, 7, "มูลค่ารวมพอร์ต / Total Portfolio Value", "1", "R", true, 0)
	p.Cell(56, 7, "฿917,792.16", "1", "R", true, 0)
	p.SetY(p.GetY() + 12)

	// Transaction history
	w.doc.SetFont("sarabun", "B", 12)
	w.doc.SetTextColor(navy[0], navy[1], navy[2])
	p.SetX(lMargin)
	p.Cell(w.bodyWidth, 7, "ประวัติธุรกรรมล่าสุด (Folio: FN-001234)", "", "L", false, 1)

	p.SetY(p.GetY() + 1)

	type txn struct {
		date   string
		txType string
		units  string
		nav    string
		amount string
	}
	txns := []txn{
		{"15/03/2026", "ซื้อ (Subscribe)", "+500.0000", "15.5000", "7,750.00"},
		{"28/02/2026", "เงินปันผล (Dividend)", "—", "—", "1,234.56"},
		{"15/01/2026", "ซื้อ (Subscribe)", "+1,000.0000", "15.2000", "15,200.00"},
		{"30/12/2025", "ขาย (Redeem)", "-200.0000", "14.8500", "2,970.00"},
	}

	tColW := []float64{24, 38, 28, 20, 28}
	tHeaders := []string{"วันที่", "ประเภท", "หน่วย", "NAV", "จำนวนเงิน (฿)"}

	w.doc.SetFont("sarabun", "B", 9)
	w.doc.SetFillColor(darkTeal[0], darkTeal[1], darkTeal[2])
	w.doc.SetTextColor(white[0], white[1], white[2])
	w.doc.SetDrawColor(darkTeal[0], darkTeal[1], darkTeal[2])
	p.SetX(lMargin)
	for i, h := range tHeaders {
		p.Cell(tColW[i], 7, h, "1", "C", true, 0)
	}
	p.SetY(p.GetY() + 7)

	w.doc.SetFont("sarabun", "", 9)
	w.doc.SetTextColor(bodyText[0], bodyText[1], bodyText[2])
	w.doc.SetDrawColor(rule[0], rule[1], rule[2])

	for i, t := range txns {
		if i%2 == 0 {
			w.doc.SetFillColor(zebra[0], zebra[1], zebra[2])
		} else {
			w.doc.SetFillColor(white[0], white[1], white[2])
		}
		p.SetX(lMargin)
		vals := []string{t.date, t.txType, t.units, t.nav, t.amount}
		als := []string{"C", "L", "R", "R", "R"}
		for j, v := range vals {
			p.Cell(tColW[j], 6, v, "1", als[j], true, 0)
		}
		p.SetY(p.GetY() + 6)
	}

	w.pageFooter()
}

// =====================================================================
// Folio Brand + Others
// =====================================================================

func (w *writer) drawFolioBrandAndOthers() {
	w.newPage()
	p := w.page

	w.sectionHeader("3", "แบรนด์สินค้า — Folio Brand", "Consumer Products")

	w.doc.SetFont("sarabun", "", 11)
	w.doc.SetTextColor(bodyText[0], bodyText[1], bodyText[2])
	p.SetX(lMargin)
	p.MultiCell(w.bodyWidth, 5.8,
		"Folio ยังเป็นชื่อแบรนด์สินค้าที่ผลิตและจำหน่ายผลิตภัณฑ์เครื่องหนัง "+
			"คุณภาพสูง ได้แก่ สมุดบันทึก (Notebook) ไดอารี่ (Diary) ออกาไนเซอร์หนัง "+
			"(Leather Organizer) และกระเป๋าใส่เอกสาร (Portfolio Bag) "+
			"สินค้าเหล่านี้มักใช้เป็นของขวัญในองค์กรหรือสำหรับมืออาชีพ",
		"", "J", false)

	p.SetY(p.GetY() + 5)

	// Product catalog table
	w.doc.SetFont("sarabun", "B", 12)
	w.doc.SetTextColor(navy[0], navy[1], navy[2])
	p.SetX(lMargin)
	p.Cell(w.bodyWidth, 7, "ตัวอย่างสินค้า Folio Brand", "", "C", false, 1)
	p.SetY(p.GetY() + 1)

	type product struct {
		code  string
		name  string
		mat   string
		price string
	}
	products := []product{
		{"FL-NB-001", "สมุดบันทึกหนังแท้ A5", "หนังวัวฟอกฝาด", "฿1,290"},
		{"FL-DY-002", "ไดอารี่รายวัน 2026", "หนัง PU เกรดพรีเมียม", "฿890"},
		{"FL-OG-003", "ออกาไนเซอร์ 6 ห่วง", "หนังแท้อิตาลี", "฿2,450"},
		{"FL-PB-004", "กระเป๋าใส่เอกสาร", "หนังวัวเต็มตัว", "฿3,890"},
		{"FL-PH-005", "ซองใส่พาสปอร์ต", "หนังแท้ขัดมัน", "฿690"},
	}

	pColW := []float64{24, 52, 48, 22}
	pHeaders := []string{"รหัสสินค้า", "ชื่อสินค้า", "วัสดุ", "ราคา"}

	w.doc.SetFont("sarabun", "B", 9)
	w.doc.SetFillColor(darkTeal[0], darkTeal[1], darkTeal[2])
	w.doc.SetTextColor(white[0], white[1], white[2])
	w.doc.SetDrawColor(darkTeal[0], darkTeal[1], darkTeal[2])
	p.SetX(lMargin + 12)
	for i, h := range pHeaders {
		p.Cell(pColW[i], 7, h, "1", "C", true, 0)
	}
	p.SetY(p.GetY() + 7)

	w.doc.SetFont("sarabun", "", 9)
	w.doc.SetTextColor(bodyText[0], bodyText[1], bodyText[2])
	w.doc.SetDrawColor(rule[0], rule[1], rule[2])

	for i, pr := range products {
		if i%2 == 0 {
			w.doc.SetFillColor(zebra[0], zebra[1], zebra[2])
		} else {
			w.doc.SetFillColor(white[0], white[1], white[2])
		}
		p.SetX(lMargin + 12)
		vals := []string{pr.code, pr.name, pr.mat, pr.price}
		als := []string{"C", "L", "L", "R"}
		for j, v := range vals {
			p.Cell(pColW[j], 6.5, v, "1", als[j], true, 0)
		}
		p.SetY(p.GetY() + 6.5)
	}

	// --- Section 4: Others ---
	p.SetY(p.GetY() + 10)

	w.sectionHeader("4", "อื่นๆ — Other Uses of \"Folio\"", "Library, Education & More")

	w.doc.SetFont("sarabun", "", 11)
	w.doc.SetTextColor(bodyText[0], bodyText[1], bodyText[2])

	others := []struct {
		title string
		desc  string
		bg    [3]int
	}{
		{
			"FOLIO — ระบบบริหารจัดการห้องสมุด (Library Services Platform)",
			"FOLIO (The Future of Libraries is Open) เป็นแพลตฟอร์มโอเพนซอร์ส " +
				"สำหรับบริหารจัดการห้องสมุด ครอบคลุมระบบจัดหา (Acquisitions) " +
				"ระบบยืม-คืน (Circulation) การจัดการเมทาดาต้า (Cataloging) " +
				"และการจัดการทรัพยากรอิเล็กทรอนิกส์ (E-Resource Management) " +
				"พัฒนาโดยชุมชนห้องสมุดทั่วโลก",
			lightBlue,
		},
		{
			"iFolio — แฟ้มสะสมผลงานออนไลน์ (Digital Portfolio)",
			"iFolio เป็นเครื่องมือสำหรับสร้างแฟ้มสะสมผลงานดิจิทัล (e-Portfolio) " +
				"ที่นักเรียน นักศึกษา และมืออาชีพใช้รวบรวมผลงาน โปรเจกต์ " +
				"ใบรับรอง และประสบการณ์การทำงาน เพื่อนำเสนอตัวเองแก่สถาบันการศึกษา " +
				"หรือนายจ้างในอนาคต",
			lightGreen,
		},
		{
			"Folio (การพิมพ์ / สิ่งพิมพ์)",
			"ในวงการสิ่งพิมพ์ Folio หมายถึงขนาดกระดาษที่พับครึ่งหนึ่ง " +
				"ทำให้ได้ 2 ใบ (4 หน้า) ซึ่งเป็นขนาดหนังสือที่ใหญ่ที่สุด " +
				"นอกจากนี้ยังหมายถึงเลขหน้าที่พิมพ์ไว้มุมกระดาษอีกด้วย",
			lightGold,
		},
	}

	// Use fixed-height boxes so we can draw background first, then text on top.
	boxHeights := []float64{30, 26, 22}
	for i, o := range others {
		y := p.GetY() + 2
		boxH := boxHeights[i]

		// Draw filled background box first
		w.doc.SetFillColor(o.bg[0], o.bg[1], o.bg[2])
		w.doc.SetDrawColor(rule[0], rule[1], rule[2])
		p.Rect(lMargin, y, w.bodyWidth, boxH, "DF")

		// Title
		w.doc.SetFont("sarabun", "B", 10)
		w.doc.SetTextColor(navy[0], navy[1], navy[2])
		p.SetXY(lMargin+4, y+2)
		p.Cell(w.bodyWidth-8, 5, o.title, "", "L", false, 1)

		// Body
		w.doc.SetFont("sarabun", "", 9.5)
		w.doc.SetTextColor(bodyText[0], bodyText[1], bodyText[2])
		p.SetX(lMargin + 4)
		p.MultiCell(w.bodyWidth-8, 4.8, o.desc, "", "J", false)

		p.SetY(y + boxH + 2)
	}

	w.pageFooter()
}

// =====================================================================
// Shared helpers
// =====================================================================

func (w *writer) sectionHeader(_, thTitle, enSubtitle string) {
	p := w.page
	y := p.GetY()

	// Colored accent bar
	w.doc.SetFillColor(accent[0], accent[1], accent[2])
	p.Rect(lMargin, y, 5, 14, "F")

	// Thai title
	w.doc.SetFont("sarabun", "B", 16)
	w.doc.SetTextColor(navy[0], navy[1], navy[2])
	p.SetXY(lMargin+8, y)
	p.Cell(w.bodyWidth-8, 8, thTitle, "", "L", false, 1)

	// English subtitle
	w.doc.SetFont("sarabun", "", 11)
	w.doc.SetTextColor(mutedText[0], mutedText[1], mutedText[2])
	p.SetX(lMargin + 8)
	p.Cell(w.bodyWidth-8, 6, enSubtitle, "", "L", false, 1)

	// Rule
	p.SetY(p.GetY() + 2)
	w.doc.SetDrawColor(accent[0], accent[1], accent[2])
	w.doc.SetLineWidth(0.5)
	p.Line(lMargin, p.GetY(), lMargin+w.bodyWidth, p.GetY())
	p.SetY(p.GetY() + 5)
}

func (w *writer) pageFooter() {
	w.doc.SetFont("sarabun", "", 8)
	w.doc.SetTextColor(mutedText[0], mutedText[1], mutedText[2])
	w.page.SetXY(lMargin, pageH-12)
	w.page.Cell(w.bodyWidth, 4, "สร้างโดย Folio PDF Library — github.com/akkaraponph/folio", "", "C", false, 0)
}

func fmtNum(v float64) string {
	s := fmt.Sprintf("%.2f", v)
	parts := [2]string{}
	for i, c := range s {
		if c == '.' {
			parts[0] = s[:i]
			parts[1] = s[i+1:]
			break
		}
	}
	// Insert thousands separators
	n := len(parts[0])
	var b []byte
	for i := range n {
		if i > 0 && (n-i)%3 == 0 {
			b = append(b, ',')
		}
		b = append(b, parts[0][i])
	}
	return string(b) + "." + parts[1]
}

func fail(format string, a ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", a...)
	os.Exit(1)
}

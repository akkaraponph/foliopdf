package main

import (
	"fmt"
	"os"

	"github.com/akkaraponph/presspdf"
	"github.com/akkaraponph/presspdf/fonts/sarabun"
)

func main() {
	doc := presspdf.New(presspdf.WithCompression(false))
	doc.SetTitle("Thai Language Demo")
	doc.SetAuthor("Folio")
	doc.SetMargins(15, 15, 15)

	// Register all Sarabun styles
	if err := sarabun.Register(doc); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	page := doc.AddPage(presspdf.A4)

	// Title
	doc.SetFont("sarabun", "B", 24)
	page.SetXY(15, 15)
	page.Cell(180, 12, "ภาษาไทยใน Folio", "", "C", false, 1)

	// Subtitle
	doc.SetFont("sarabun", "", 14)
	doc.SetTextColor(100, 100, 100)
	page.Cell(180, 8, "Thai Language Support Demo", "", "C", false, 1)
	page.SetY(page.GetY() + 5)

	// Regular text
	doc.SetTextColor(0, 0, 0)
	doc.SetFont("sarabun", "", 12)
	page.SetXY(15, page.GetY())
	page.MultiCell(180, 6,
		"สวัสดีครับ นี่คือตัวอย่างการใช้งานภาษาไทยในไลบรารี Folio "+
			"ซึ่งรองรับฟอนต์ TrueType แบบ Unicode ผ่าน CIDFont Type2 "+
			"ทำให้สามารถแสดงผลภาษาไทยได้อย่างสมบูรณ์ รวมถึงสระ วรรณยุกต์ "+
			"และเครื่องหมายต่างๆ",
		"", "L", false)
	page.SetY(page.GetY() + 5)

	// Bold text
	doc.SetFont("sarabun", "B", 12)
	page.SetXY(15, page.GetY())
	page.Cell(180, 6, "ตัวหนา: กขคงจฉชซฌญฎฏฐฑฒณดตถทธนบปผฝพฟภมยรลวศษสหฬอฮ", "", "L", false, 1)

	// Italic text
	doc.SetFont("sarabun", "I", 12)
	page.SetXY(15, page.GetY())
	page.Cell(180, 6, "ตัวเอียง: สวัสดีปีใหม่ มีความสุขมากๆ นะครับ", "", "L", false, 1)

	// Bold Italic
	doc.SetFont("sarabun", "BI", 12)
	page.SetXY(15, page.GetY())
	page.Cell(180, 6, "ตัวหนาเอียง: ภาษาไทยสวยงามมาก", "", "L", false, 1)
	page.SetY(page.GetY() + 5)

	// Tone marks demo
	doc.SetFont("sarabun", "", 14)
	doc.SetTextColor(40, 80, 140)
	page.SetXY(15, page.GetY())
	page.Cell(180, 7, "วรรณยุกต์ทั้ง 4:", "", "L", false, 1)

	doc.SetTextColor(0, 0, 0)
	doc.SetFont("sarabun", "", 12)
	page.SetXY(15, page.GetY())
	page.Cell(45, 6, "ก่า (เอก)", "", "L", false, 0)
	page.Cell(45, 6, "ก้า (โท)", "", "L", false, 0)
	page.Cell(45, 6, "ก๊า (ตรี)", "", "L", false, 0)
	page.Cell(45, 6, "ก๋า (จัตวา)", "", "L", false, 1)
	page.SetY(page.GetY() + 5)

	// Mixed Thai and English
	doc.SetFont("sarabun", "", 12)
	doc.SetTextColor(40, 80, 140)
	page.SetXY(15, page.GetY())
	page.Cell(180, 7, "ภาษาไทยผสมภาษาอังกฤษ:", "", "L", false, 1)

	doc.SetTextColor(0, 0, 0)
	page.SetXY(15, page.GetY())
	page.MultiCell(180, 6,
		"Folio เป็น PDF library ที่เขียนด้วยภาษา Go "+
			"มี architecture แบบ 4 layers ทำให้โค้ดเป็นระเบียบ "+
			"และง่ายต่อการ extend ฟีเจอร์ใหม่ๆ "+
			"เช่น Thai language support ที่คุณเห็นอยู่ตอนนี้",
		"", "L", false)
	page.SetY(page.GetY() + 5)

	// Numbers
	doc.SetFont("sarabun", "", 12)
	doc.SetTextColor(40, 80, 140)
	page.SetXY(15, page.GetY())
	page.Cell(180, 7, "ตัวเลข:", "", "L", false, 1)

	doc.SetTextColor(0, 0, 0)
	page.SetXY(15, page.GetY())
	page.Cell(90, 6, "เลขไทย: ๐ ๑ ๒ ๓ ๔ ๕ ๖ ๗ ๘ ๙", "", "L", false, 0)
	page.Cell(90, 6, "เลขอารบิก: 0 1 2 3 4 5 6 7 8 9", "", "L", false, 1)

	path := "/tmp/folio_thai.pdf"
	if err := doc.Save(path); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Thai PDF saved to %s\n", path)
}

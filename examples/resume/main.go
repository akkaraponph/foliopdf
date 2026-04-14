package main

import (
	"fmt"
	"os"

	"github.com/akkaraponph/presspdf"
)

const (
	pageW    = 210.0 // A4 width in mm
	marginL  = 15.0
	marginR  = 15.0
	marginT  = 15.0
	contentW = pageW - marginL - marginR
)

func main() {
	doc := presspdf.New(presspdf.WithCompression(false))
	doc.SetTitle("Resume - Bob Smith")
	doc.SetAuthor("Bob Smith")
	doc.SetMargins(marginL, marginT, marginR)

	page := doc.AddPage(presspdf.A4)

	// ── Header: Name ──
	doc.SetFont("helvetica", "B", 24)
	page.SetXY(marginL, marginT)
	page.Cell(contentW, 10, "Bob Smith", "", "L", false, 1)

	// ── Header: Title ──
	doc.SetTextColor(100, 100, 100)
	doc.SetFont("helvetica", "", 12)
	page.Cell(contentW, 6, "Senior Software Engineer", "", "L", false, 1)

	// ── Contact line ──
	doc.SetFont("helvetica", "", 9)
	doc.SetTextColor(80, 80, 80)
	page.SetXY(marginL, page.GetY()+1)
	page.Cell(contentW, 5, "akkarapon@example.com  |  +66 123 456 789  |  Bangkok, Thailand  |  github.com/akkaraponph", "", "L", false, 1)

	// ── Divider ──
	page.SetXY(marginL, page.GetY()+2)
	doc.SetDrawColor(60, 60, 60)
	doc.SetLineWidth(0.4)
	page.Line(marginL, page.GetY(), pageW-marginR, page.GetY())
	page.SetY(page.GetY() + 4)

	// ── Summary ──
	sectionHeader(doc, page, "SUMMARY")
	doc.SetFont("helvetica", "", 10)
	doc.SetTextColor(40, 40, 40)
	page.SetXY(marginL, page.GetY())
	page.MultiCell(contentW, 5,
		"Software engineer with 8+ years of experience building scalable backend systems, "+
			"PDF generation libraries, and developer tools in Go. Passionate about clean architecture, "+
			"performance optimization, and open-source contributions.",
		"", "L", false)
	page.SetY(page.GetY() + 3)

	// ── Experience ──
	sectionHeader(doc, page, "EXPERIENCE")

	jobEntry(doc, page,
		"Senior Software Engineer", "PressPDF Co., Ltd.", "2022 - Present",
		[]string{
			"Designed and built PressPDF, a layered PDF generation library in Go with 4-layer architecture",
			"Led backend team of 5 engineers, established code review process and CI/CD pipelines",
			"Reduced PDF generation latency by 60% through resource deduplication and stream compression",
			"Migrated monolithic service to microservices architecture serving 50K+ daily active users",
		})

	jobEntry(doc, page,
		"Software Engineer", "TechCorp International", "2019 - 2022",
		[]string{
			"Built REST APIs in Go serving 10M+ requests/day with p99 latency under 50ms",
			"Implemented real-time notification system using WebSockets and Redis pub/sub",
			"Designed PostgreSQL schema and query optimization reducing report generation time by 75%",
		})

	jobEntry(doc, page,
		"Junior Developer", "StartupHub", "2017 - 2019",
		[]string{
			"Developed internal tools for inventory management and order processing",
			"Created automated testing framework that improved test coverage from 40% to 85%",
		})

	// ── Skills ──
	sectionHeader(doc, page, "SKILLS")

	skillRow(doc, page, "Languages:", "Go, Python, TypeScript, SQL")
	skillRow(doc, page, "Frameworks:", "Gin, Echo, gRPC, React")
	skillRow(doc, page, "Infrastructure:", "Docker, Kubernetes, AWS, Terraform, GitHub Actions")
	skillRow(doc, page, "Databases:", "PostgreSQL, Redis, MongoDB, Elasticsearch")
	skillRow(doc, page, "Practices:", "Clean Architecture, TDD, CI/CD, Code Review, Agile")
	page.SetY(page.GetY() + 3)

	// ── Education ──
	sectionHeader(doc, page, "EDUCATION")

	doc.SetFont("helvetica", "B", 10)
	doc.SetTextColor(30, 30, 30)
	page.SetXY(marginL, page.GetY())
	page.Cell(0, 5, "B.Eng. Computer Engineering", "", "L", false, 1)

	doc.SetFont("helvetica", "", 9)
	doc.SetTextColor(80, 80, 80)
	page.SetXY(marginL, page.GetY())
	page.Cell(contentW/2, 5, "Chulalongkorn University", "", "L", false, 0)
	page.Cell(contentW/2, 5, "2013 - 2017", "", "R", false, 1)
	page.SetY(page.GetY() + 3)

	// ── Projects ──
	sectionHeader(doc, page, "PROJECTS")

	projectEntry(doc, page, "PressPDF", "github.com/akkaraponph/presspdf",
		"Layered PDF generation library for Go. Clean 4-layer architecture with core font support, "+
			"JPEG image embedding, cell/multicell with word wrapping, and resource deduplication.")

	projectEntry(doc, page, "API Gateway", "internal",
		"High-performance API gateway handling authentication, rate limiting, and request routing "+
			"for 15+ microservices with circuit breaker pattern.")

	// ── Save ──
	path := "/tmp/resume.pdf"
	if err := doc.Save(path); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Resume saved to %s\n", path)
}

// sectionHeader draws a section title with a thin line underneath.
func sectionHeader(doc *presspdf.Document, page *presspdf.Page, title string) {
	doc.SetFont("helvetica", "B", 11)
	doc.SetTextColor(40, 80, 140)
	page.SetXY(marginL, page.GetY())
	page.Cell(contentW, 6, title, "", "L", false, 1)

	doc.SetDrawColor(40, 80, 140)
	doc.SetLineWidth(0.3)
	page.Line(marginL, page.GetY(), pageW-marginR, page.GetY())
	page.SetY(page.GetY() + 3)
}

// jobEntry draws a single work experience block.
func jobEntry(doc *presspdf.Document, page *presspdf.Page, title, company, dates string, bullets []string) {
	// Title row
	doc.SetFont("helvetica", "B", 10)
	doc.SetTextColor(30, 30, 30)
	page.SetXY(marginL, page.GetY())
	page.Cell(contentW*0.7, 5, title, "", "L", false, 0)

	doc.SetFont("helvetica", "", 9)
	doc.SetTextColor(80, 80, 80)
	page.Cell(contentW*0.3, 5, dates, "", "R", false, 1)

	// Company
	doc.SetFont("helvetica", "I", 9)
	doc.SetTextColor(60, 60, 60)
	page.SetXY(marginL, page.GetY())
	page.Cell(contentW, 5, company, "", "L", false, 1)

	// Bullets
	doc.SetFont("helvetica", "", 9)
	doc.SetTextColor(40, 40, 40)
	for _, b := range bullets {
		page.SetXY(marginL+4, page.GetY())
		page.Cell(4, 4.5, "-", "", "L", false, 0)
		page.MultiCell(contentW-8, 4.5, b, "", "L", false)
	}
	page.SetY(page.GetY() + 3)
}

// skillRow draws a label + value row for skills.
func skillRow(doc *presspdf.Document, page *presspdf.Page, label, value string) {
	doc.SetFont("helvetica", "B", 9)
	doc.SetTextColor(40, 40, 40)
	page.SetXY(marginL, page.GetY())
	page.Cell(28, 5, label, "", "L", false, 0)

	doc.SetFont("helvetica", "", 9)
	page.Cell(contentW-28, 5, value, "", "L", false, 1)
}

// projectEntry draws a project block with name, url, and description.
func projectEntry(doc *presspdf.Document, page *presspdf.Page, name, url, description string) {
	doc.SetFont("helvetica", "B", 10)
	doc.SetTextColor(30, 30, 30)
	page.SetXY(marginL, page.GetY())
	page.Cell(0, 5, name, "", "L", false, 1)

	doc.SetFont("helvetica", "I", 8)
	doc.SetTextColor(80, 80, 80)
	page.SetXY(marginL, page.GetY())
	page.Cell(0, 4, url, "", "L", false, 1)

	doc.SetFont("helvetica", "", 9)
	doc.SetTextColor(40, 40, 40)
	page.SetXY(marginL, page.GetY())
	page.MultiCell(contentW, 4.5, description, "", "L", false)
	page.SetY(page.GetY() + 3)
}

// Package foliopdf is a layered PDF generation library for Go.
//
// Folio generates valid PDF files with text, drawing primitives, and images
// using an idiomatic Go API built on clean internal layers.
//
// Basic usage:
//
//	doc := foliopdf.New()
//	doc.SetTitle("Invoice")
//	doc.SetFont("helvetica", "", 16)
//
//	page := doc.AddPage(foliopdf.A4)
//	page.TextAt(40, 60, "Hello Folio")
//	page.Line(40, 70, 200, 70)
//
//	err := doc.Save("output.pdf")
package foliopdf

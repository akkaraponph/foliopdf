package folio

import (
	"fmt"
	"time"

	"github.com/akkaraponph/folio/internal/pdfcore"
)

// serialize builds the complete PDF byte stream.
func (d *Document) serialize() (*pdfcore.Writer, error) {
	if len(d.pages) == 0 {
		return nil, fmt.Errorf("folio: no pages")
	}

	w := pdfcore.NewWriter()

	// 1. Header
	w.WriteHeader("1.4")

	// 2. Pages: page dicts + content streams, then Pages root at obj 1
	pageObjNums := d.putPages(w)

	// 3. Fonts
	d.putFonts(w)

	// 4. Images
	d.putImages(w)

	// 5. Resource dictionary at obj 2
	d.putResourceDict(w)

	// 6. Info dictionary
	infoObjNum := d.putInfo(w)

	// 7. Catalog
	catalogObjNum := d.putCatalog(w, pageObjNums)

	// 8. Xref
	xrefOffset := w.WriteXref()

	// 9. Trailer
	w.WriteTrailer(catalogObjNum, infoObjNum)
	w.WriteStartXref(xrefOffset)

	return w, w.Err()
}

// putPages writes page dictionaries and content streams.
// Returns the object numbers of each page dictionary.
func (d *Document) putPages(w *pdfcore.Writer) []int {
	pageObjNums := make([]int, len(d.pages))

	for i, p := range d.pages {
		// Page dictionary
		pageObj := w.NewObj()
		pageObjNums[i] = pageObj
		contentObj := pageObj + 1 // content stream will be the next object

		w.Put("<<")
		w.Put("/Type /Page")
		w.Putf("/Parent 1 0 R")
		w.Putf("/MediaBox [0 0 %.2f %.2f]", p.size.WidthPt, p.size.HeightPt)
		w.Put("/Resources 2 0 R")
		w.Putf("/Contents %d 0 R", contentObj)
		w.Put(">>")
		w.EndObj()

		// Content stream
		data := p.stream.Bytes()
		w.NewObj()
		if d.compress && len(data) > 0 {
			w.PutCompressedStream(data)
		} else {
			w.Putf("<</Length %d>>", len(data))
			w.PutStream(data)
		}
		w.EndObj()
	}

	// Pages root at object 1
	w.SetOffset(1)
	w.Putf("%d 0 obj", 1)
	w.Put("<<")
	w.Put("/Type /Pages")

	// Kids array
	kids := "/Kids ["
	for i, n := range pageObjNums {
		if i > 0 {
			kids += " "
		}
		kids += fmt.Sprintf("%d 0 R", n)
	}
	kids += "]"
	w.Put(kids)

	w.Putf("/Count %d", len(d.pages))
	w.Put(">>")
	w.Put("endobj")

	return pageObjNums
}

// putFonts writes font objects for all registered fonts.
func (d *Document) putFonts(w *pdfcore.Writer) {
	for _, fe := range d.fonts.All() {
		n := w.NewObj()
		fe.ObjNum = n
		w.Put("<<")
		w.Put("/Type /Font")
		w.Put("/Subtype /Type1")
		w.Putf("/BaseFont /%s", fe.Name)
		if fe.Name != "Symbol" && fe.Name != "ZapfDingbats" {
			w.Put("/Encoding /WinAnsiEncoding")
		}
		w.Put(">>")
		w.EndObj()
	}
}

// putImages writes image XObject for all registered images.
func (d *Document) putImages(w *pdfcore.Writer) {
	for _, ie := range d.images.All() {
		n := w.NewObj()
		ie.ObjNum = n
		w.Put("<<")
		w.Put("/Type /XObject")
		w.Put("/Subtype /Image")
		w.Putf("/Width %d", ie.Width)
		w.Putf("/Height %d", ie.Height)
		w.Putf("/ColorSpace /%s", ie.ColorSpace)
		w.Putf("/BitsPerComponent %d", ie.BPC)
		w.Putf("/Filter /%s", ie.Filter)
		w.Putf("/Length %d", len(ie.Data))
		w.Put(">>")
		w.PutStream(ie.Data)
		w.EndObj()
	}
}

// putResourceDict writes the shared resource dictionary at object 2.
func (d *Document) putResourceDict(w *pdfcore.Writer) {
	w.SetOffset(2)
	w.Putf("%d 0 obj", 2)
	w.Put("<<")
	w.Put("/ProcSet [/PDF /Text /ImageB /ImageC /ImageI]")

	// Font references
	fonts := d.fonts.All()
	if len(fonts) > 0 {
		s := "/Font <<"
		for _, fe := range fonts {
			s += fmt.Sprintf(" /F%s %d 0 R", fe.Index, fe.ObjNum)
		}
		s += " >>"
		w.Put(s)
	}

	// Image references
	images := d.images.All()
	if len(images) > 0 {
		s := "/XObject <<"
		for _, ie := range images {
			s += fmt.Sprintf(" /Im%s %d 0 R", ie.Name, ie.ObjNum)
		}
		s += " >>"
		w.Put(s)
	}

	w.Put(">>")
	w.Put("endobj")
}

// putInfo writes the document info dictionary.
func (d *Document) putInfo(w *pdfcore.Writer) int {
	n := w.NewObj()
	w.Put("<<")
	w.Putf("/Producer %s", pdfString(d.producer))
	if d.title != "" {
		w.Putf("/Title %s", pdfString(d.title))
	}
	if d.author != "" {
		w.Putf("/Author %s", pdfString(d.author))
	}
	if d.subject != "" {
		w.Putf("/Subject %s", pdfString(d.subject))
	}
	if d.creator != "" {
		w.Putf("/Creator %s", pdfString(d.creator))
	}
	w.Putf("/CreationDate %s", pdfString(pdfDate(time.Now())))
	w.Put(">>")
	w.EndObj()
	return n
}

// putCatalog writes the document catalog.
func (d *Document) putCatalog(w *pdfcore.Writer, pageObjNums []int) int {
	n := w.NewObj()
	w.Put("<<")
	w.Put("/Type /Catalog")
	w.Put("/Pages 1 0 R")
	w.Put(">>")
	w.EndObj()
	return n
}

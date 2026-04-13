package presspdf

import (
	"crypto"
	"crypto/x509"
	"fmt"

	"github.com/akkaraponph/presspdf/internal/state"
)

// SignOptions configures the digital signature.
type SignOptions struct {
	// Name of the signer (appears in signature panel).
	Name string
	// Reason for signing.
	Reason string
	// Location of signing.
	Location string
	// ContactInfo for the signer.
	ContactInfo string
}

// signatureState holds the signing configuration for the document.
type signatureState struct {
	cert    *x509.Certificate
	key     crypto.Signer
	opts    SignOptions
	page    *Page
	x, y    float64 // position of visible signature rect
	w, h    float64 // size of visible signature rect
}

// Sign configures the document to be digitally signed using the given
// certificate and private key. The signature is placed on the specified
// page at position (x, y) with size (w, h) in user units.
// The actual signing happens during serialization.
func (d *Document) Sign(cert *x509.Certificate, key crypto.Signer, page *Page, x, y, w, h float64, opts SignOptions) {
	d.sigState = &signatureState{
		cert: cert,
		key:  key,
		opts: opts,
		page: page,
		x:    x,
		y:    y,
		w:    w,
		h:    h,
	}
}

// putSignature writes the signature field, widget annotation, and
// placeholder PKCS#7 content. Returns the signature field object number.
// The /Contents value is a hex placeholder that would be filled with
// the actual PKCS#7 signature in a production implementation.
func (d *Document) putSignature(w pdfObjWriter, pageObjNums []int) int {
	if d.sigState == nil {
		return 0
	}

	sig := d.sigState
	k := d.k

	pageIdx := d.pageIndex(sig.page)
	if pageIdx < 0 {
		d.err = fmt.Errorf("Sign: signature page not found")
		return 0
	}

	// Signature value object (contains the PKCS#7 /Contents placeholder).
	sigValObj := w.NewObj()
	w.Put("<<")
	w.Put("/Type /Sig")
	w.Put("/Filter /Adobe.PPKLite")
	w.Put("/SubFilter /adbe.pkcs7.detached")

	if sig.opts.Name != "" {
		w.Putf("/Name %s", pdfString(sig.opts.Name))
	}
	if sig.opts.Reason != "" {
		w.Putf("/Reason %s", pdfString(sig.opts.Reason))
	}
	if sig.opts.Location != "" {
		w.Putf("/Location %s", pdfString(sig.opts.Location))
	}
	if sig.opts.ContactInfo != "" {
		w.Putf("/ContactInfo %s", pdfString(sig.opts.ContactInfo))
	}

	// Placeholder /Contents: 8KB of zeros (hex encoded = 16KB).
	// In a full implementation, this would be replaced with actual PKCS#7.
	placeholder := make([]byte, 8192)
	w.Putf("/Contents <%x>", placeholder)

	// ByteRange placeholder — would be filled after PDF is assembled.
	w.Put("/ByteRange [0 0 0 0]")

	w.Put(">>")
	w.EndObj()

	// Signature field / widget annotation.
	fieldObj := w.NewObj()
	x1 := state.ToPointsX(sig.x, k)
	y1 := state.ToPointsY(sig.y+sig.h, sig.page.h, k)
	x2 := x1 + sig.w*k
	y2 := y1 + sig.h*k

	w.Put("<<")
	w.Put("/Type /Annot")
	w.Put("/Subtype /Widget")
	w.Put("/FT /Sig")
	w.Putf("/Rect [%.2f %.2f %.2f %.2f]", x1, y1, x2, y2)
	w.Putf("/V %d 0 R", sigValObj)
	w.Putf("/T %s", pdfString("Signature1"))
	w.Put("/F 132") // Print + Locked
	w.Putf("/P %d 0 R", pageObjNums[pageIdx])
	w.Put(">>")
	w.EndObj()

	return fieldObj
}

// pdfObjWriter is the interface for writing PDF objects.
type pdfObjWriter interface {
	NewObj() int
	EndObj()
	Put(s string)
	Putf(format string, args ...any)
	PutStream(data []byte)
}

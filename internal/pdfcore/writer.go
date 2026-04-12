package pdfcore

import (
	"bytes"
	"compress/zlib"
	"fmt"
	"io"
)

// Writer handles low-level PDF byte-stream generation.
// It owns the main document buffer and the object-numbering system.
//
// Objects 1 and 2 are reserved:
//   - Object 1: Pages root (/Type /Pages)
//   - Object 2: Resource dictionary
//
// They are written out of order; call SetOffset when writing them.
type Writer struct {
	buf     bytes.Buffer
	n       int   // current (highest allocated) object number
	offsets []int // offsets[objNum] = byte position in buf
	err     error

	// encryption state
	encKey     []byte // file encryption key (nil = no encryption)
	currentObj int    // object number being written (for per-object key)
	encryptFn  func(key []byte, objNum, genNum int, data []byte) []byte
}

// NewWriter creates a Writer with objects 1 and 2 reserved.
func NewWriter() *Writer {
	w := &Writer{
		n:       2,
		offsets: make([]int, 3), // indices 0, 1, 2
	}
	return w
}

// NewObj allocates the next object number, records its byte offset,
// and writes the "N 0 obj\n" header. Returns the new object number.
func (w *Writer) NewObj() int {
	if w.err != nil {
		return 0
	}
	w.n++
	w.currentObj = w.n
	w.grow(w.n)
	w.offsets[w.n] = w.buf.Len()
	fmt.Fprintf(&w.buf, "%d 0 obj\n", w.n)
	return w.n
}

// SetEncryption enables RC4 encryption for all subsequently written streams.
// encryptFn receives (key, objNum, genNum, data) and returns encrypted data.
func (w *Writer) SetEncryption(key []byte, fn func(key []byte, objNum, genNum int, data []byte) []byte) {
	w.encKey = key
	w.encryptFn = fn
}

// EndObj writes "endobj\n".
func (w *Writer) EndObj() {
	if w.err != nil {
		return
	}
	w.buf.WriteString("endobj\n")
}

// Putf writes a formatted string followed by a newline.
func (w *Writer) Putf(format string, args ...any) {
	if w.err != nil {
		return
	}
	fmt.Fprintf(&w.buf, format, args...)
	w.buf.WriteByte('\n')
}

// Put writes a string followed by a newline.
func (w *Writer) Put(s string) {
	if w.err != nil {
		return
	}
	w.buf.WriteString(s)
	w.buf.WriteByte('\n')
}

// PutStream writes a raw (uncompressed) stream.
// The caller must have already written the dictionary with /Length.
// If encryption is enabled, the stream data is encrypted with the
// per-object RC4 key.
func (w *Writer) PutStream(data []byte) {
	if w.err != nil {
		return
	}
	if w.encKey != nil && w.encryptFn != nil {
		data = w.encryptFn(w.encKey, w.currentObj, 0, data)
	}
	w.buf.WriteString("stream\n")
	w.buf.Write(data)
	w.buf.WriteString("\nendstream\n")
}

// PutCompressedStream writes a zlib-compressed stream with its dictionary.
func (w *Writer) PutCompressedStream(data []byte) {
	if w.err != nil {
		return
	}
	var compressed bytes.Buffer
	zw, err := zlib.NewWriterLevel(&compressed, zlib.DefaultCompression)
	if err != nil {
		w.err = fmt.Errorf("pdfcore: zlib init: %w", err)
		return
	}
	if _, err := zw.Write(data); err != nil {
		w.err = fmt.Errorf("pdfcore: zlib write: %w", err)
		return
	}
	if err := zw.Close(); err != nil {
		w.err = fmt.Errorf("pdfcore: zlib close: %w", err)
		return
	}
	cb := compressed.Bytes()
	w.Putf("<</Length %d /Filter /FlateDecode>>", len(cb))
	w.PutStream(cb)
}

// PutRawStream writes a stream with a pre-built dictionary line.
// dictLine should look like "<</Length N>>" (no trailing newline).
func (w *Writer) PutRawStream(dictLine string, data []byte) {
	if w.err != nil {
		return
	}
	w.Put(dictLine)
	w.PutStream(data)
}

// SetOffset records the current buffer position as the byte offset
// for the given object number. Used for reserved objects (1 and 2)
// that are written out of their allocation order.
func (w *Writer) SetOffset(objNum int) {
	w.grow(objNum)
	w.offsets[objNum] = w.buf.Len()
}

// WriteHeader writes the PDF header line.
func (w *Writer) WriteHeader(version string) {
	if w.err != nil {
		return
	}
	fmt.Fprintf(&w.buf, "%%PDF-%s\n", version)
}

// WriteXref writes the cross-reference table and returns the byte
// offset where the xref section begins (needed for startxref).
func (w *Writer) WriteXref() int {
	if w.err != nil {
		return 0
	}
	offset := w.buf.Len()
	w.buf.WriteString("xref\n")
	fmt.Fprintf(&w.buf, "0 %d\n", w.n+1)
	w.buf.WriteString("0000000000 65535 f \n")
	for j := 1; j <= w.n; j++ {
		fmt.Fprintf(&w.buf, "%010d 00000 n \n", w.offsets[j])
	}
	return offset
}

// WriteTrailer writes the trailer dictionary.
func (w *Writer) WriteTrailer(rootObjNum, infoObjNum int) {
	w.WriteTrailerEncrypt(rootObjNum, infoObjNum, 0, nil)
}

// WriteTrailerEncrypt writes the trailer with optional encryption refs.
// encryptObjNum is the /Encrypt dict object number (0 = no encryption).
// fileID is the document ID for /ID array (nil = omit).
func (w *Writer) WriteTrailerEncrypt(rootObjNum, infoObjNum, encryptObjNum int, fileID []byte) {
	if w.err != nil {
		return
	}
	w.buf.WriteString("trailer\n")
	w.buf.WriteString("<<\n")
	fmt.Fprintf(&w.buf, "/Size %d\n", w.n+1)
	fmt.Fprintf(&w.buf, "/Root %d 0 R\n", rootObjNum)
	fmt.Fprintf(&w.buf, "/Info %d 0 R\n", infoObjNum)
	if encryptObjNum > 0 {
		fmt.Fprintf(&w.buf, "/Encrypt %d 0 R\n", encryptObjNum)
	}
	if fileID != nil {
		hex := fmt.Sprintf("%x", fileID)
		fmt.Fprintf(&w.buf, "/ID [<%s> <%s>]\n", hex, hex)
	}
	w.buf.WriteString(">>\n")
}

// WriteStartXref writes the startxref pointer and %%EOF marker.
func (w *Writer) WriteStartXref(xrefOffset int) {
	if w.err != nil {
		return
	}
	w.buf.WriteString("startxref\n")
	fmt.Fprintf(&w.buf, "%d\n", xrefOffset)
	w.buf.WriteString("%%EOF\n")
}

// WriteTo copies the entire buffer to dst.
func (w *Writer) WriteTo(dst io.Writer) (int64, error) {
	if w.err != nil {
		return 0, w.err
	}
	return w.buf.WriteTo(dst)
}

// Len returns the current buffer length in bytes.
func (w *Writer) Len() int { return w.buf.Len() }

// ObjCount returns the number of allocated objects.
func (w *Writer) ObjCount() int { return w.n }

// Err returns the first accumulated error.
func (w *Writer) Err() error { return w.err }

// SetError records an error if none has been set yet.
func (w *Writer) SetError(err error) {
	if w.err == nil {
		w.err = err
	}
}

// grow ensures offsets has capacity for index n.
func (w *Writer) grow(n int) {
	for len(w.offsets) <= n {
		w.offsets = append(w.offsets, 0)
	}
}

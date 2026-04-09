package content

import (
	"bytes"
	"fmt"
)

// Stream builds a PDF content stream — the sequence of operators
// between "stream" and "endstream" in a page content object.
// All coordinates must be in PDF points (bottom-left origin).
type Stream struct {
	buf bytes.Buffer
}

// Bytes returns the content stream bytes.
func (s *Stream) Bytes() []byte { return s.buf.Bytes() }

// Len returns the byte length of the stream.
func (s *Stream) Len() int { return s.buf.Len() }

// Reset clears the stream.
func (s *Stream) Reset() { s.buf.Reset() }

// --- Graphics state ---

// SaveState emits q (save graphics state).
func (s *Stream) SaveState() { s.buf.WriteString("q\n") }

// RestoreState emits Q (restore graphics state).
func (s *Stream) RestoreState() { s.buf.WriteString("Q\n") }

// SetLineWidth emits the w operator.
func (s *Stream) SetLineWidth(w float64) {
	fmt.Fprintf(&s.buf, "%.2f w\n", w)
}

// SetLineCap emits the J operator (0=butt, 1=round, 2=square).
func (s *Stream) SetLineCap(style int) {
	fmt.Fprintf(&s.buf, "%d J\n", style)
}

// SetLineJoin emits the j operator (0=miter, 1=round, 2=bevel).
func (s *Stream) SetLineJoin(style int) {
	fmt.Fprintf(&s.buf, "%d j\n", style)
}

// --- Color ---

// SetStrokeColorRGB emits the RG operator. Values in 0..1.
func (s *Stream) SetStrokeColorRGB(r, g, b float64) {
	fmt.Fprintf(&s.buf, "%.3f %.3f %.3f RG\n", r, g, b)
}

// SetFillColorRGB emits the rg operator. Values in 0..1.
func (s *Stream) SetFillColorRGB(r, g, b float64) {
	fmt.Fprintf(&s.buf, "%.3f %.3f %.3f rg\n", r, g, b)
}

// SetStrokeGray emits the G operator.
func (s *Stream) SetStrokeGray(g float64) {
	fmt.Fprintf(&s.buf, "%.3f G\n", g)
}

// SetFillGray emits the g operator.
func (s *Stream) SetFillGray(g float64) {
	fmt.Fprintf(&s.buf, "%.3f g\n", g)
}

// --- Text ---

// BeginText emits BT.
func (s *Stream) BeginText() { s.buf.WriteString("BT\n") }

// EndText emits ET.
func (s *Stream) EndText() { s.buf.WriteString("ET\n") }

// SetFont emits the Tf operator. name is the resource name (e.g. "F1").
func (s *Stream) SetFont(name string, sizePt float64) {
	fmt.Fprintf(&s.buf, "/%s %.2f Tf\n", name, sizePt)
}

// MoveText emits the Td operator to position text.
func (s *Stream) MoveText(tx, ty float64) {
	fmt.Fprintf(&s.buf, "%.2f %.2f Td\n", tx, ty)
}

// ShowText emits the Tj operator. The string must already be PDF-escaped.
func (s *Stream) ShowText(escaped string) {
	fmt.Fprintf(&s.buf, "(%s) Tj\n", escaped)
}

// SetTextLeading emits the TL operator.
func (s *Stream) SetTextLeading(leading float64) {
	fmt.Fprintf(&s.buf, "%.2f TL\n", leading)
}

// NextLine emits T* (move to start of next line).
func (s *Stream) NextLine() { s.buf.WriteString("T*\n") }

// --- Path construction ---

// MoveTo emits the m operator (begin new subpath).
func (s *Stream) MoveTo(x, y float64) {
	fmt.Fprintf(&s.buf, "%.2f %.2f m\n", x, y)
}

// LineTo emits the l operator (line segment).
func (s *Stream) LineTo(x, y float64) {
	fmt.Fprintf(&s.buf, "%.2f %.2f l\n", x, y)
}

// CurveTo emits the c operator (cubic bezier).
func (s *Stream) CurveTo(x1, y1, x2, y2, x3, y3 float64) {
	fmt.Fprintf(&s.buf, "%.2f %.2f %.2f %.2f %.2f %.2f c\n", x1, y1, x2, y2, x3, y3)
}

// Rect emits the re operator (rectangle).
func (s *Stream) Rect(x, y, w, h float64) {
	fmt.Fprintf(&s.buf, "%.2f %.2f %.2f %.2f re\n", x, y, w, h)
}

// ClosePath emits h (close current subpath).
func (s *Stream) ClosePath() { s.buf.WriteString("h\n") }

// --- Path painting ---

// Stroke emits S (stroke the path).
func (s *Stream) Stroke() { s.buf.WriteString("S\n") }

// Fill emits f (fill the path using non-zero winding rule).
func (s *Stream) Fill() { s.buf.WriteString("f\n") }

// FillStroke emits B (fill then stroke).
func (s *Stream) FillStroke() { s.buf.WriteString("B\n") }

// --- Image ---

// DrawImage emits operators to draw an image XObject.
// The CTM matrix [a b c d e f] positions and scales the image.
// name is the resource name (e.g. "Im1").
func (s *Stream) DrawImage(name string, a, b, c, d, e, f float64) {
	s.buf.WriteString("q\n")
	fmt.Fprintf(&s.buf, "%.5f %.5f %.5f %.5f %.5f %.5f cm\n", a, b, c, d, e, f)
	fmt.Fprintf(&s.buf, "/%s Do\n", name)
	s.buf.WriteString("Q\n")
}

// --- Raw ---

// Raw writes an arbitrary operator string followed by newline.
func (s *Stream) Raw(op string) {
	s.buf.WriteString(op)
	s.buf.WriteByte('\n')
}

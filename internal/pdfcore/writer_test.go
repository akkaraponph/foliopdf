package pdfcore

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewWriter(t *testing.T) {
	w := NewWriter()
	if w.n != 2 {
		t.Fatalf("initial n = %d, want 2", w.n)
	}
	if len(w.offsets) != 3 {
		t.Fatalf("initial offsets len = %d, want 3", len(w.offsets))
	}
}

func TestNewObj(t *testing.T) {
	w := NewWriter()
	w.WriteHeader("1.4")
	n := w.NewObj()
	if n != 3 {
		t.Fatalf("first NewObj = %d, want 3", n)
	}
	s := w.buf.String()
	if !strings.Contains(s, "3 0 obj\n") {
		t.Fatalf("buffer missing '3 0 obj': %q", s)
	}
}

func TestPutStream(t *testing.T) {
	w := NewWriter()
	w.WriteHeader("1.4")
	w.NewObj()
	data := []byte("BT /F1 12 Tf ET")
	w.Putf("<</Length %d>>", len(data))
	w.PutStream(data)
	w.EndObj()

	s := w.buf.String()
	if !strings.Contains(s, "stream\n") {
		t.Fatal("missing stream keyword")
	}
	if !strings.Contains(s, "BT /F1 12 Tf ET") {
		t.Fatal("missing stream data")
	}
	if !strings.Contains(s, "\nendstream\n") {
		t.Fatal("missing endstream")
	}
}

func TestXrefOffsets(t *testing.T) {
	w := NewWriter()
	w.WriteHeader("1.4")

	// Write object 3
	w.NewObj()
	w.Put("<< /Type /Page >>")
	w.EndObj()

	// Write reserved object 1
	w.SetOffset(1)
	w.Putf("%d 0 obj", 1)
	w.Put("<< /Type /Pages >>")
	w.Put("endobj")

	// Verify offsets are correct
	s := w.buf.String()
	for objNum := 1; objNum <= w.n; objNum++ {
		off := w.offsets[objNum]
		if off == 0 && objNum > 2 {
			t.Fatalf("offset for object %d is 0", objNum)
		}
		// Check that the recorded offset points to valid content
		if off > 0 && off < len(s) {
			chunk := s[off:]
			if objNum == 1 {
				if !strings.HasPrefix(chunk, "1 0 obj") {
					t.Errorf("offset %d for obj 1 points to %q", off, chunk[:min(20, len(chunk))])
				}
			}
		}
	}
}

func TestWriteXref(t *testing.T) {
	w := NewWriter()
	w.WriteHeader("1.4")
	w.NewObj()
	w.Put("<< /Test true >>")
	w.EndObj()

	// Write reserved objects (simplified)
	w.SetOffset(1)
	w.Putf("%d 0 obj\n<< /Type /Pages >>\nendobj", 1)
	w.SetOffset(2)
	w.Putf("%d 0 obj\n<< /Type /Resources >>\nendobj", 2)

	xrefOff := w.WriteXref()
	w.WriteTrailer(3, 3) // just for testing
	w.WriteStartXref(xrefOff)

	s := w.buf.String()
	if !strings.Contains(s, "xref\n0 4\n") {
		t.Fatalf("bad xref header in: %s", s)
	}
	if !strings.Contains(s, "0000000000 65535 f \n") {
		t.Fatal("missing free entry")
	}
	if !strings.Contains(s, "%%EOF") {
		t.Fatal("missing EOF marker")
	}
}

func TestWriteTo(t *testing.T) {
	w := NewWriter()
	w.WriteHeader("1.4")
	var buf bytes.Buffer
	n, err := w.WriteTo(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if n == 0 {
		t.Fatal("WriteTo wrote 0 bytes")
	}
	if !strings.HasPrefix(buf.String(), "%PDF-1.4") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestErrorAccumulation(t *testing.T) {
	w := NewWriter()
	w.SetError(bytes.ErrTooLarge)
	n := w.NewObj()
	if n != 0 {
		t.Fatalf("NewObj after error returned %d, want 0", n)
	}
	if w.Err() != bytes.ErrTooLarge {
		t.Fatalf("Err() = %v, want ErrTooLarge", w.Err())
	}
}

package content

import (
	"testing"
)

func TestTextStream(t *testing.T) {
	var s Stream
	s.BeginText()
	s.SetFont("F1", 12)
	s.MoveText(72, 720)
	s.ShowText("Hello")
	s.EndText()

	want := "BT\n/F1 12.00 Tf\n72.00 720.00 Td\n(Hello) Tj\nET\n"
	got := s.buf.String()
	if got != want {
		t.Fatalf("text stream:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestPathStream(t *testing.T) {
	var s Stream
	s.MoveTo(100, 200)
	s.LineTo(300, 200)
	s.Stroke()

	want := "100.00 200.00 m\n300.00 200.00 l\nS\n"
	got := s.buf.String()
	if got != want {
		t.Fatalf("path stream:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestRectStream(t *testing.T) {
	var s Stream
	s.Rect(50, 50, 200, 100)
	s.Stroke()

	want := "50.00 50.00 200.00 100.00 re\nS\n"
	got := s.buf.String()
	if got != want {
		t.Fatalf("rect stream:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestColorStream(t *testing.T) {
	var s Stream
	s.SetStrokeColorRGB(1, 0, 0)
	s.SetFillColorRGB(0, 0, 1)

	want := "1.000 0.000 0.000 RG\n0.000 0.000 1.000 rg\n"
	got := s.buf.String()
	if got != want {
		t.Fatalf("color stream:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestGraphicsState(t *testing.T) {
	var s Stream
	s.SaveState()
	s.SetLineWidth(2.5)
	s.RestoreState()

	want := "q\n2.50 w\nQ\n"
	got := s.buf.String()
	if got != want {
		t.Fatalf("state stream:\ngot:  %q\nwant: %q", got, want)
	}
}

func TestDrawImage(t *testing.T) {
	var s Stream
	s.DrawImage("Im1", 200, 0, 0, 150, 72, 600)

	got := s.buf.String()
	if got == "" {
		t.Fatal("empty image stream")
	}
	// Should contain q, cm, Do, Q
	for _, kw := range []string{"q\n", "cm\n", "/Im1 Do\n", "Q\n"} {
		if !contains(got, kw) {
			t.Errorf("image stream missing %q", kw)
		}
	}
}

func TestBytesAndLen(t *testing.T) {
	var s Stream
	s.Raw("test")
	if s.Len() != 5 { // "test\n"
		t.Fatalf("Len() = %d, want 5", s.Len())
	}
	if string(s.Bytes()) != "test\n" {
		t.Fatalf("Bytes() = %q", s.Bytes())
	}
	s.Reset()
	if s.Len() != 0 {
		t.Fatal("Reset did not clear")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

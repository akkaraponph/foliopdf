package state

import (
	"math"
	"testing"
)

func TestScaleFactor(t *testing.T) {
	tests := []struct {
		unit Unit
		want float64
	}{
		{UnitPt, 1.0},
		{UnitMM, 72.0 / 25.4},
		{UnitCM, 72.0 / 2.54},
		{UnitInch, 72.0},
	}
	for _, tt := range tests {
		got := ScaleFactor(tt.unit)
		if math.Abs(got-tt.want) > 0.0001 {
			t.Errorf("ScaleFactor(%d) = %f, want %f", tt.unit, got, tt.want)
		}
	}
}

func TestToPointsY(t *testing.T) {
	k := ScaleFactor(UnitMM)
	// Top of A4 page (y=0) should map to full page height in points
	got := ToPointsY(0, 297.0, k)
	want := 297.0 * k // ≈ 841.89
	if math.Abs(got-want) > 0.01 {
		t.Errorf("ToPointsY(0, 297, k) = %f, want %f", got, want)
	}

	// Bottom of page (y=297) should map to 0
	got = ToPointsY(297.0, 297.0, k)
	if math.Abs(got) > 0.01 {
		t.Errorf("ToPointsY(297, 297, k) = %f, want 0", got)
	}
}

func TestToPointsX(t *testing.T) {
	k := ScaleFactor(UnitMM)
	got := ToPointsX(210.0, k) // A4 width
	want := 210.0 * k          // ≈ 595.28
	if math.Abs(got-want) > 0.01 {
		t.Errorf("ToPointsX(210, k) = %f, want %f", got, want)
	}
}

func TestColorFromRGB(t *testing.T) {
	c := ColorFromRGB(255, 0, 128)
	if c.R != 1.0 {
		t.Errorf("R = %f, want 1.0", c.R)
	}
	if c.G != 0.0 {
		t.Errorf("G = %f, want 0.0", c.G)
	}
	// 128/255 ≈ 0.502
	if math.Abs(c.B-128.0/255.0) > 0.001 {
		t.Errorf("B = %f, want %f", c.B, 128.0/255.0)
	}
}

func TestColorOps(t *testing.T) {
	c := ColorFromRGB(255, 0, 0)
	if s := c.StrokeOp(); s != "1.000 0.000 0.000 RG" {
		t.Errorf("StrokeOp = %q", s)
	}
	if s := c.FillOp(); s != "1.000 0.000 0.000 rg" {
		t.Errorf("FillOp = %q", s)
	}
}

func TestColorIsBlack(t *testing.T) {
	if !ColorFromRGB(0, 0, 0).IsBlack() {
		t.Error("black should be black")
	}
	if ColorFromRGB(1, 0, 0).IsBlack() {
		t.Error("red should not be black")
	}
}

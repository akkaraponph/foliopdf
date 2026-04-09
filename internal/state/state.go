package state

import "fmt"

// Color represents an RGB color with components in the range 0..1.
type Color struct {
	R, G, B float64
}

// ColorFromRGB converts 0-255 integer RGB values to a Color.
func ColorFromRGB(r, g, b int) Color {
	return Color{
		R: float64(r) / 255.0,
		G: float64(g) / 255.0,
		B: float64(b) / 255.0,
	}
}

// StrokeOp returns the PDF operator string for setting stroke color.
func (c Color) StrokeOp() string {
	return fmt.Sprintf("%.3f %.3f %.3f RG", c.R, c.G, c.B)
}

// FillOp returns the PDF operator string for setting fill color.
func (c Color) FillOp() string {
	return fmt.Sprintf("%.3f %.3f %.3f rg", c.R, c.G, c.B)
}

// IsBlack returns true if the color is black (all components zero).
func (c Color) IsBlack() bool {
	return c.R == 0 && c.G == 0 && c.B == 0
}

// Equal returns true if two colors are identical.
func (c Color) Equal(other Color) bool {
	return c.R == other.R && c.G == other.G && c.B == other.B
}

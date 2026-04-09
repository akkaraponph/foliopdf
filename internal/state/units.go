package state

// Unit represents a unit of measure.
type Unit int

const (
	UnitMM   Unit = iota // millimeters (default)
	UnitPt               // points (1/72 inch)
	UnitCM               // centimeters
	UnitInch             // inches
)

// ScaleFactor returns points-per-unit for the given Unit.
// Multiply a value in user units by this factor to get PDF points.
func ScaleFactor(u Unit) float64 {
	switch u {
	case UnitPt:
		return 1.0
	case UnitMM:
		return 72.0 / 25.4
	case UnitCM:
		return 72.0 / 2.54
	case UnitInch:
		return 72.0
	default:
		return 72.0 / 25.4 // default to mm
	}
}

// ToPointsX converts a user-unit X coordinate to PDF points.
func ToPointsX(x, k float64) float64 {
	return x * k
}

// ToPointsY converts a user-unit Y coordinate (top-down origin)
// to PDF points (bottom-up origin).
// pageH is the page height in user units.
func ToPointsY(y, pageH, k float64) float64 {
	return (pageH - y) * k
}

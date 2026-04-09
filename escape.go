package folio

import "strings"

// pdfEscape escapes special PDF string characters: \, (, )
func pdfEscape(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, "(", `\(`)
	s = strings.ReplaceAll(s, ")", `\)`)
	return s
}

// pdfString wraps s in PDF string delimiters with escaping.
func pdfString(s string) string {
	return "(" + pdfEscape(s) + ")"
}

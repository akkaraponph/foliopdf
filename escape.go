package presspdf

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

// unicodeToWinAnsi converts a UTF-8 Go string into a byte string using
// WinAnsiEncoding (Windows-1252). Core PDF fonts declare this encoding,
// so multi-byte UTF-8 sequences must be mapped to the correct single
// bytes. Characters without a WinAnsi mapping are replaced with '?'.
func unicodeToWinAnsi(s string) string {
	// Fast path: all ASCII.
	allASCII := true
	for i := 0; i < len(s); i++ {
		if s[i] > 127 {
			allASCII = false
			break
		}
	}
	if allASCII {
		return s
	}

	buf := make([]byte, 0, len(s))
	for _, r := range s {
		if r < 128 {
			buf = append(buf, byte(r))
		} else if b, ok := unicodeToWinAnsiMap[r]; ok {
			buf = append(buf, b)
		} else if r >= 0xA0 && r <= 0xFF {
			// Latin-1 supplement: direct mapping.
			buf = append(buf, byte(r))
		} else {
			buf = append(buf, '?')
		}
	}
	return string(buf)
}

// unicodeToWinAnsiMap maps Unicode code points to WinAnsiEncoding bytes
// for the 0x80–0x9F range (the Windows-1252 extensions not in Latin-1).
var unicodeToWinAnsiMap = map[rune]byte{
	0x20AC: 0x80, // €
	0x201A: 0x82, // ‚
	0x0192: 0x83, // ƒ
	0x201E: 0x84, // „
	0x2026: 0x85, // …
	0x2020: 0x86, // †
	0x2021: 0x87, // ‡
	0x02C6: 0x88, // ˆ
	0x2030: 0x89, // ‰
	0x0160: 0x8A, // Š
	0x2039: 0x8B, // ‹
	0x0152: 0x8C, // Œ
	0x017D: 0x8E, // Ž
	0x2018: 0x91, // '
	0x2019: 0x92, // '
	0x201C: 0x93, // "
	0x201D: 0x94, // "
	0x2022: 0x95, // •
	0x2013: 0x96, // –
	0x2014: 0x97, // —
	0x02DC: 0x98, // ˜
	0x2122: 0x99, // ™
	0x0161: 0x9A, // š
	0x203A: 0x9B, // ›
	0x0153: 0x9C, // œ
	0x017E: 0x9E, // ž
	0x0178: 0x9F, // Ÿ
}

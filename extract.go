package presspdf

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/akkaraponph/presspdf/internal/pdfcore"
)

// ExtractText extracts all text from a PDF file, returning one string per page.
func ExtractText(pdfPath string) ([]string, error) {
	data, err := os.ReadFile(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("ExtractText: %w", err)
	}
	return ExtractTextFromBytes(data)
}

// ExtractTextFromBytes extracts all text from PDF data in memory,
// returning one string per page.
func ExtractTextFromBytes(data []byte) ([]string, error) {
	r, err := pdfcore.ReadPDF(data)
	if err != nil {
		return nil, fmt.Errorf("ExtractText: %w", err)
	}

	pageRefs, err := r.PageRefs()
	if err != nil {
		return nil, fmt.Errorf("ExtractText: %w", err)
	}

	var result []string
	for _, ref := range pageRefs {
		text, err := extractPageText(r, ref)
		if err != nil {
			// Skip pages we can't parse rather than failing entirely.
			result = append(result, "")
			continue
		}
		result = append(result, text)
	}

	return result, nil
}

// extractPageText extracts text from a single page.
func extractPageText(r *pdfcore.Reader, pageRef pdfcore.Ref) (string, error) {
	pageObj, err := r.Object(pageRef.Num)
	if err != nil {
		return "", err
	}
	pageDict := pdfcore.ToDict(pageObj)
	if pageDict == nil {
		return "", fmt.Errorf("page %d is not a dict", pageRef.Num)
	}

	// Get content stream(s).
	contentsVal, err := r.Resolve(pageDict["/Contents"])
	if err != nil {
		return "", err
	}

	var streamData []byte
	switch cv := contentsVal.(type) {
	case *pdfcore.Stream:
		decoded, err := pdfcore.DecodeStream(cv)
		if err != nil {
			return "", err
		}
		streamData = decoded
	case pdfcore.Ref:
		obj, err := r.Object(cv.Num)
		if err != nil {
			return "", err
		}
		if stm, ok := obj.(*pdfcore.Stream); ok {
			decoded, err := pdfcore.DecodeStream(stm)
			if err != nil {
				return "", err
			}
			streamData = decoded
		}
	case []interface{}:
		// Multiple content streams — concatenate.
		var buf bytes.Buffer
		for _, item := range cv {
			ref, ok := item.(pdfcore.Ref)
			if !ok {
				continue
			}
			obj, err := r.Object(ref.Num)
			if err != nil {
				continue
			}
			if stm, ok := obj.(*pdfcore.Stream); ok {
				decoded, err := pdfcore.DecodeStream(stm)
				if err != nil {
					continue
				}
				buf.Write(decoded)
				buf.WriteByte('\n')
			}
		}
		streamData = buf.Bytes()
	}

	if len(streamData) == 0 {
		return "", nil
	}

	return parseTextFromContentStream(streamData), nil
}

// parseTextFromContentStream extracts text strings from a PDF content stream.
// It handles Tj, TJ, ', and " text-showing operators.
func parseTextFromContentStream(data []byte) string {
	var result strings.Builder
	s := string(data)

	// Simple state machine to extract text from PDF operators.
	// We look for text between BT...ET blocks and extract Tj/TJ strings.
	i := 0
	inText := false

	for i < len(s) {
		// Skip whitespace.
		if s[i] == ' ' || s[i] == '\n' || s[i] == '\r' || s[i] == '\t' {
			i++
			continue
		}

		// Check for BT/ET.
		if i+2 <= len(s) && s[i:i+2] == "BT" && (i+2 >= len(s) || isDelim(s[i+2])) {
			inText = true
			i += 2
			continue
		}
		if i+2 <= len(s) && s[i:i+2] == "ET" && (i+2 >= len(s) || isDelim(s[i+2])) {
			inText = false
			if result.Len() > 0 {
				result.WriteByte('\n')
			}
			i += 2
			continue
		}

		if !inText {
			i++
			continue
		}

		// Look for string operands: (text) or <hex>
		if s[i] == '(' {
			text, end := readPDFString(s, i)
			if end > i {
				result.WriteString(text)
				i = end
				// Check if followed by Tj or '
				i = skipWhitespace(s, i)
				if i+2 <= len(s) && s[i:i+2] == "Tj" {
					i += 2
				} else if i+1 <= len(s) && s[i] == '\'' {
					result.WriteByte('\n')
					i++
				}
				continue
			}
		}

		// Look for hex strings: <hex>
		if s[i] == '<' && (i+1 >= len(s) || s[i+1] != '<') {
			text, end := readHexString(s, i)
			if end > i {
				result.WriteString(text)
				i = end
				i = skipWhitespace(s, i)
				if i+2 <= len(s) && s[i:i+2] == "Tj" {
					i += 2
				}
				continue
			}
		}

		// Look for TJ array: [ (text) 123 (more) ] TJ
		if s[i] == '[' {
			text, end := readTJArray(s, i)
			if end > i {
				result.WriteString(text)
				i = end
				i = skipWhitespace(s, i)
				if i+2 <= len(s) && s[i:i+2] == "TJ" {
					i += 2
				}
				continue
			}
		}

		i++
	}

	return strings.TrimSpace(result.String())
}

func isDelim(b byte) bool {
	return b == ' ' || b == '\n' || b == '\r' || b == '\t' || b == '/' || b == '[' || b == '('
}

func skipWhitespace(s string, i int) int {
	for i < len(s) && (s[i] == ' ' || s[i] == '\n' || s[i] == '\r' || s[i] == '\t') {
		i++
	}
	return i
}

// readPDFString reads a parenthesized PDF string starting at index i.
// Returns the decoded text and the position after the closing ')'.
func readPDFString(s string, i int) (string, int) {
	if i >= len(s) || s[i] != '(' {
		return "", i
	}
	i++ // skip '('
	var buf strings.Builder
	depth := 1
	for i < len(s) && depth > 0 {
		ch := s[i]
		switch ch {
		case '\\':
			i++
			if i >= len(s) {
				break
			}
			esc := s[i]
			switch esc {
			case 'n':
				buf.WriteByte('\n')
			case 'r':
				buf.WriteByte('\r')
			case 't':
				buf.WriteByte('\t')
			case 'b':
				buf.WriteByte('\b')
			case 'f':
				buf.WriteByte('\f')
			case '(', ')', '\\':
				buf.WriteByte(esc)
			default:
				// Octal escape
				if esc >= '0' && esc <= '7' {
					oct := int(esc - '0')
					for j := 0; j < 2 && i+1 < len(s) && s[i+1] >= '0' && s[i+1] <= '7'; j++ {
						i++
						oct = oct*8 + int(s[i]-'0')
					}
					buf.WriteByte(byte(oct))
				} else {
					buf.WriteByte(esc)
				}
			}
		case '(':
			depth++
			buf.WriteByte(ch)
		case ')':
			depth--
			if depth > 0 {
				buf.WriteByte(ch)
			}
		default:
			buf.WriteByte(ch)
		}
		i++
	}
	return buf.String(), i
}

// readHexString reads a hex-encoded PDF string starting at index i.
func readHexString(s string, i int) (string, int) {
	if i >= len(s) || s[i] != '<' {
		return "", i
	}
	end := strings.IndexByte(s[i:], '>')
	if end < 0 {
		return "", i
	}
	hex := s[i+1 : i+end]
	end = i + end + 1

	// Decode hex pairs to bytes.
	hex = strings.ReplaceAll(hex, " ", "")
	hex = strings.ReplaceAll(hex, "\n", "")
	hex = strings.ReplaceAll(hex, "\r", "")
	if len(hex)%2 != 0 {
		hex += "0"
	}
	var buf strings.Builder
	for j := 0; j+1 < len(hex); j += 2 {
		var b byte
		for k := 0; k < 2; k++ {
			c := hex[j+k]
			b <<= 4
			switch {
			case c >= '0' && c <= '9':
				b |= c - '0'
			case c >= 'a' && c <= 'f':
				b |= c - 'a' + 10
			case c >= 'A' && c <= 'F':
				b |= c - 'A' + 10
			}
		}
		if b >= 32 && b < 127 { // printable ASCII only
			buf.WriteByte(b)
		}
	}
	return buf.String(), end
}

// readTJArray reads a TJ array [string num string num ...] starting at '['.
func readTJArray(s string, i int) (string, int) {
	if i >= len(s) || s[i] != '[' {
		return "", i
	}
	i++ // skip '['
	var buf strings.Builder
	for i < len(s) {
		if s[i] == ']' {
			return buf.String(), i + 1
		}
		if s[i] == '(' {
			text, end := readPDFString(s, i)
			buf.WriteString(text)
			i = end
			continue
		}
		if s[i] == '<' && (i+1 >= len(s) || s[i+1] != '<') {
			text, end := readHexString(s, i)
			buf.WriteString(text)
			i = end
			continue
		}
		// Skip numbers and whitespace inside TJ array.
		// Large negative numbers (< -100) typically indicate word spacing.
		if s[i] == '-' || (s[i] >= '0' && s[i] <= '9') {
			numStart := i
			if s[i] == '-' {
				i++
			}
			for i < len(s) && ((s[i] >= '0' && s[i] <= '9') || s[i] == '.') {
				i++
			}
			// If the number is a large negative, add a space (word boundary).
			numStr := s[numStart:i]
			if len(numStr) > 0 && numStr[0] == '-' {
				// Parse as a rough magnitude check.
				if len(numStr) >= 4 { // e.g. "-200" or larger
					buf.WriteByte(' ')
				}
			}
			continue
		}
		i++
	}
	return buf.String(), i
}

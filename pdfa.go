package foliopdf

import (
	"encoding/binary"
	"fmt"
	"math"
	"time"
)

// pdfaConf holds the PDF/A conformance configuration.
type pdfaConf struct {
	part        int    // 1 or 2
	conformance string // "B"
}

// WithPDFA returns an Option that enables PDF/A compliance.
// Supported levels: "1b" (PDF/A-1b) and "2b" (PDF/A-2b).
func WithPDFA(level string) Option {
	return func(d *Document) {
		switch level {
		case "1b":
			d.pdfaLevel = &pdfaConf{part: 1, conformance: "B"}
		case "2b":
			d.pdfaLevel = &pdfaConf{part: 2, conformance: "B"}
		default:
			d.err = fmt.Errorf("WithPDFA: unsupported level %q (use \"1b\" or \"2b\")", level)
		}
	}
}

// validatePDFA checks the document for PDF/A conformance issues.
// Returns an error if any violations are found.
func (d *Document) validatePDFA() error {
	if d.pdfaLevel == nil {
		return nil
	}

	// PDF/A requires all fonts to be embedded. Core fonts are not embedded.
	for _, fe := range d.fonts.All() {
		if fe.Type != "TTF" {
			return fmt.Errorf("PDF/A: font %q is a core font and not embedded; PDF/A requires all fonts to be embedded", fe.Name)
		}
	}

	// PDF/A-1b forbids transparency.
	if d.pdfaLevel.part == 1 && len(d.alphaStates) > 0 {
		return fmt.Errorf("PDF/A-1b: transparency (alpha) is not allowed")
	}

	return nil
}

// pdfaVersion returns the PDF version string for the conformance level.
func (d *Document) pdfaVersion() string {
	if d.encryptAES {
		return "2.0" // AES-256 requires PDF 2.0
	}
	if d.pdfaLevel == nil {
		return "1.4"
	}
	if d.pdfaLevel.part >= 2 {
		return "1.7"
	}
	return "1.4"
}

// buildXMPMetadata generates the XMP metadata XML for PDF/A conformance.
func (d *Document) buildXMPMetadata() []byte {
	now := time.Now().UTC()
	ts := now.Format("2006-01-02T15:04:05Z")

	title := d.title
	if title == "" {
		title = "Untitled"
	}
	author := d.author
	if author == "" {
		author = "Unknown"
	}
	creator := d.creator
	if creator == "" {
		creator = d.producer
	}

	xmp := fmt.Sprintf(`<?xpacket begin="%s" id="W5M0MpCehiHzreSzNTczkc9d"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
  <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
    <rdf:Description rdf:about=""
      xmlns:dc="http://purl.org/dc/elements/1.1/"
      xmlns:xmp="http://ns.adobe.com/xap/1.0/"
      xmlns:pdf="http://ns.adobe.com/pdf/1.3/"
      xmlns:pdfaid="http://www.aiim.org/pdfa/ns/id/">
      <pdfaid:part>%d</pdfaid:part>
      <pdfaid:conformance>%s</pdfaid:conformance>
      <dc:title>
        <rdf:Alt>
          <rdf:li xml:lang="x-default">%s</rdf:li>
        </rdf:Alt>
      </dc:title>
      <dc:creator>
        <rdf:Seq>
          <rdf:li>%s</rdf:li>
        </rdf:Seq>
      </dc:creator>
      <xmp:CreatorTool>%s</xmp:CreatorTool>
      <xmp:CreateDate>%s</xmp:CreateDate>
      <xmp:ModifyDate>%s</xmp:ModifyDate>
      <pdf:Producer>%s</pdf:Producer>
    </rdf:Description>
  </rdf:RDF>
</x:xmpmeta>
<?xpacket end="w"?>`,
		"\xEF\xBB\xBF", // UTF-8 BOM
		d.pdfaLevel.part,
		d.pdfaLevel.conformance,
		xmlEscape(title),
		xmlEscape(author),
		xmlEscape(creator),
		ts, ts,
		xmlEscape(d.producer),
	)
	return []byte(xmp)
}

// xmlEscape escapes special XML characters.
func xmlEscape(s string) string {
	var out []byte
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '&':
			out = append(out, []byte("&amp;")...)
		case '<':
			out = append(out, []byte("&lt;")...)
		case '>':
			out = append(out, []byte("&gt;")...)
		case '"':
			out = append(out, []byte("&quot;")...)
		default:
			out = append(out, s[i])
		}
	}
	return string(out)
}

// putMetadata writes the XMP metadata stream. Returns the object number.
func (d *Document) putMetadata(w pdfObjWriter) int {
	if d.pdfaLevel == nil {
		return 0
	}
	data := d.buildXMPMetadata()
	n := w.NewObj()
	w.Put("<<")
	w.Put("/Type /Metadata")
	w.Put("/Subtype /XML")
	w.Putf("/Length %d", len(data))
	w.Put(">>")
	w.PutStream(data)
	w.EndObj()
	return n
}

// putOutputIntent writes the OutputIntent dict with an embedded sRGB ICC profile.
// Returns the object number of the OutputIntent.
func (d *Document) putOutputIntent(w pdfObjWriter) int {
	if d.pdfaLevel == nil {
		return 0
	}

	// Write ICC profile stream.
	iccData := buildSRGBICCProfile()
	iccObj := w.NewObj()
	w.Putf("<</Length %d /N 3>>", len(iccData))
	w.PutStream(iccData)
	w.EndObj()

	// Write OutputIntent dict.
	n := w.NewObj()
	w.Put("<<")
	w.Put("/Type /OutputIntent")
	w.Put("/S /GTS_PDFA1")
	w.Put("/OutputConditionIdentifier (sRGB)")
	w.Put("/RegistryName (http://www.color.org)")
	w.Put("/Info (sRGB IEC61966-2.1)")
	w.Putf("/DestOutputProfile %d 0 R", iccObj)
	w.Put(">>")
	w.EndObj()
	return n
}

// s15f16 converts a float64 to an ICC s15Fixed16Number (int32).
func s15f16(v float64) int32 {
	return int32(math.Round(v * 65536))
}

// buildSRGBICCProfile generates a minimal sRGB ICC v2.1 display profile.
// The profile is ~456 bytes and contains the required tags for a valid
// monitor RGB profile: description, copyright, white point, RGB colorants,
// and tone reproduction curves (gamma 2.2).
func buildSRGBICCProfile() []byte {
	const (
		profileSize = 456
		tagCount    = 9
	)

	buf := make([]byte, profileSize)
	be := binary.BigEndian

	// --- Profile Header (128 bytes) ---
	be.PutUint32(buf[0:], profileSize)          // Profile size
	copy(buf[4:8], []byte{0, 0, 0, 0})          // Preferred CMM
	be.PutUint32(buf[8:], 0x02100000)            // Version 2.1.0
	copy(buf[12:16], "mntr")                     // Device class: monitor
	copy(buf[16:20], "RGB ")                     // Color space
	copy(buf[20:24], "XYZ ")                     // PCS
	// Date/time (12 bytes at offset 24): 2024-01-01
	be.PutUint16(buf[24:], 2024)                 // Year
	be.PutUint16(buf[26:], 1)                    // Month
	be.PutUint16(buf[28:], 1)                    // Day
	be.PutUint16(buf[30:], 0)                    // Hour
	be.PutUint16(buf[32:], 0)                    // Minute
	be.PutUint16(buf[34:], 0)                    // Second
	copy(buf[36:40], "acsp")                     // Profile file signature
	// bytes 40-67: platform, flags, manufacturer, model, attributes (all zero)
	// Rendering intent (4 bytes at offset 64): 0 = perceptual
	// PCS illuminant (12 bytes at offset 68): D50
	putS15F16(buf[68:], 0.9642)  // X
	putS15F16(buf[72:], 1.0000)  // Y
	putS15F16(buf[76:], 0.8249)  // Z
	// bytes 80-127: creator, ID, reserved (all zero)

	// --- Tag Table ---
	off := 128
	be.PutUint32(buf[off:], tagCount)
	off += 4

	// Tag entries: [signature(4), offset(4), size(4)]
	tags := []struct {
		sig          string
		offset, size uint32
	}{
		{"desc", 240, 96},
		{"cprt", 336, 24},
		{"wtpt", 360, 20},
		{"rXYZ", 380, 20},
		{"gXYZ", 400, 20},
		{"bXYZ", 420, 20},
		{"rTRC", 440, 14},
		{"gTRC", 440, 14}, // shares data with rTRC
		{"bTRC", 440, 14}, // shares data with rTRC
	}
	for _, t := range tags {
		copy(buf[off:], t.sig)
		be.PutUint32(buf[off+4:], t.offset)
		be.PutUint32(buf[off+8:], t.size)
		off += 12
	}

	// --- Tag Data ---

	// 'desc' at offset 240: textDescriptionType
	off = 240
	copy(buf[off:], "desc")
	be.PutUint32(buf[off+4:], 0) // reserved
	be.PutUint32(buf[off+8:], 5) // ASCII count (including null)
	copy(buf[off+12:], "sRGB\x00")
	// Unicode lang code (off+17..20): 0
	// Unicode count (off+21..24): 0
	// ScriptCode code (off+25..26): 0
	// ScriptCode count (off+27): 0
	// ScriptCode data (off+28..94): zeros
	// Total: 96 bytes

	// 'cprt' at offset 336: textType
	off = 336
	copy(buf[off:], "text")
	be.PutUint32(buf[off+4:], 0)               // reserved
	copy(buf[off+8:], "No copyright\x00\x00\x00") // 15 chars + padding to reach 24 total

	// 'wtpt' at offset 360: XYZType (D50 media white point)
	putXYZTag(buf[360:], 0.9505, 1.0000, 1.0890)

	// 'rXYZ' at offset 380: red colorant (sRGB, Bradford-adapted to D50)
	putXYZTag(buf[380:], 0.4360747, 0.2225045, 0.0139322)

	// 'gXYZ' at offset 400: green colorant
	putXYZTag(buf[400:], 0.3850649, 0.7168786, 0.0971045)

	// 'bXYZ' at offset 420: blue colorant
	putXYZTag(buf[420:], 0.1430804, 0.0606169, 0.7141733)

	// 'rTRC'/'gTRC'/'bTRC' at offset 440: curveType gamma=2.2
	off = 440
	copy(buf[off:], "curv")
	be.PutUint32(buf[off+4:], 0) // reserved
	be.PutUint32(buf[off+8:], 1) // count = 1 (gamma mode)
	// u8Fixed8Number: 2.2 = 2*256 + round(0.2*256) = 512 + 51 = 563
	be.PutUint16(buf[off+12:], 563)
	// bytes off+14..off+15: padding (zeros)

	return buf
}

// putS15F16 writes an s15Fixed16Number at the given position.
func putS15F16(b []byte, v float64) {
	binary.BigEndian.PutUint32(b, uint32(s15f16(v)))
}

// putXYZTag writes an XYZType tag at the given position.
func putXYZTag(b []byte, x, y, z float64) {
	copy(b[0:4], "XYZ ")
	binary.BigEndian.PutUint32(b[4:], 0) // reserved
	putS15F16(b[8:], x)
	putS15F16(b[12:], y)
	putS15F16(b[16:], z)
}

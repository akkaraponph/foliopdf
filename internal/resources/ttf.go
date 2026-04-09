package resources

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"sort"
	"unicode"
)

// Composite glyph flags used in glyf table.
const (
	glyfArgWords    = 1 << 0
	glyfArgScale    = 1 << 3
	glyfMoreComps   = 1 << 5
	glyfXYScale     = 1 << 6
	glyfTwoByTwo    = 1 << 7
)

// ToUnicodeCMap is the standard CMap for mapping CIDs to Unicode code points.
// Used in the PDF ToUnicode stream for CIDFont Type2 fonts.
const ToUnicodeCMap = "/CIDInit /ProcSet findresource begin\n12 dict begin\nbegincmap\n/CIDSystemInfo\n<</Registry (Adobe)\n/Ordering (UCS)\n/Supplement 0\n>> def\n/CMapName /Adobe-Identity-UCS def\n/CMapType 2 def\n1 begincodespacerange\n<0000> <FFFF>\nendcodespacerange\n1 beginbfrange\n<0000> <FFFF> <0000>\nendbfrange\nendcmap\nCMapName currentdict /CMap defineresource pop\nend\nend"

// TTFFont holds parsed TrueType font data needed for PDF embedding.
type TTFFont struct {
	// Font metrics (scaled to 1000 units per em)
	Ascent             int
	Descent            int
	CapHeight          int
	StemV              int
	ItalicAngle        int
	Flags              int
	Bbox               [4]int // [xMin, yMin, xMax, yMax]
	UnderlinePosition  float64
	UnderlineThickness float64
	DefaultWidth       float64

	// Character widths: indexed by Unicode code point (up to 65535).
	// Values in units of 1/1000 text space.
	CharWidths []int

	// Maps: unicode code point -> glyph index
	CharToGlyph map[int]int

	// Internal state
	unitsPerEm int
	raw        []byte // original TTF file bytes
	tables     map[string]*ttfTable
	glyphPos   []int // glyph offsets in glyf table

	// Reader state
	pos int
}

type ttfTable struct {
	offset   int
	size     int
	checksum [2]int
}

// ParseTTF parses a TrueType font from raw bytes and extracts metrics
// and character widths needed for PDF generation.
func ParseTTF(data []byte) (*TTFFont, error) {
	f := &TTFFont{
		raw:         data,
		tables:      make(map[string]*ttfTable),
		CharToGlyph: make(map[int]int),
	}
	if err := f.parse(); err != nil {
		return nil, err
	}
	return f, nil
}

func (f *TTFFont) parse() error {
	f.pos = 0

	// Validate TrueType signature
	sig := f.readUint32()
	if sig == 0x4F54544F {
		return fmt.Errorf("ttf: OpenType/CFF fonts not supported")
	}
	if sig == 0x74746366 {
		return fmt.Errorf("ttf: TrueType collection (.ttc) not supported")
	}
	if sig != 0x00010000 && sig != 0x74727565 {
		return fmt.Errorf("ttf: not a TrueType font (signature=0x%08X)", sig)
	}

	// Read table directory
	numTables := f.readUint16()
	f.skip(6) // searchRange, entrySelector, rangeShift

	for range numTables {
		tag := string(f.readBytes(4))
		cs0 := f.readUint16()
		cs1 := f.readUint16()
		offset := f.readUint32()
		size := f.readUint32()
		f.tables[tag] = &ttfTable{
			offset:   offset,
			size:     size,
			checksum: [2]int{cs0, cs1},
		}
	}

	// Parse tables in order
	f.parseHead()
	metricsCount := f.parseHhea()
	weight := f.parseOS2()
	f.parsePost(weight)
	cmapPos := f.parseCmap()
	if cmapPos == 0 {
		return fmt.Errorf("ttf: no Unicode cmap found")
	}

	// maxp: number of glyphs
	f.seekTable("maxp")
	f.skip(4)
	numGlyphs := f.readUint16()

	// Build unicode <-> glyph mappings from cmap format 4
	glyphToChars := make(map[int][]int)
	f.buildCmapMappings(cmapPos, glyphToChars)

	// Parse hmtx to get character widths
	scale := 1000.0 / float64(f.unitsPerEm)
	f.parseHmtx(metricsCount, numGlyphs, glyphToChars, scale)

	return nil
}

// --- Table parsers ---

func (f *TTFFont) parseHead() {
	f.seekTable("head")
	f.skip(18) // version, fontRevision, checksumAdjust, magicNumber, flags
	f.unitsPerEm = f.readUint16()
	f.skip(16) // created, modified (2x LONGDATETIME)
	scale := 1000.0 / float64(f.unitsPerEm)
	xMin := f.readInt16()
	yMin := f.readInt16()
	xMax := f.readInt16()
	yMax := f.readInt16()
	f.Bbox = [4]int{
		int(float64(xMin) * scale),
		int(float64(yMin) * scale),
		int(float64(xMax) * scale),
		int(float64(yMax) * scale),
	}
}

func (f *TTFFont) parseHhea() int {
	t := f.tables["hhea"]
	if t == nil {
		return 0
	}
	scale := 1000.0 / float64(f.unitsPerEm)
	f.seekTable("hhea")
	f.skip(4) // version
	ascender := f.readInt16()
	descender := f.readInt16()
	f.Ascent = int(float64(ascender) * scale)
	f.Descent = int(float64(descender) * scale)
	f.skip(24) // lineGap + 11 fields
	metricDataFormat := f.readUint16()
	if metricDataFormat != 0 {
		return 0
	}
	return f.readUint16() // numberOfHMetrics
}

func (f *TTFFont) parseOS2() int {
	scale := 1000.0 / float64(f.unitsPerEm)
	t := f.tables["OS/2"]
	if t == nil {
		weight := 500
		if f.Ascent == 0 {
			f.Ascent = int(float64(f.Bbox[3]) * scale)
		}
		if f.Descent == 0 {
			f.Descent = int(float64(f.Bbox[1]) * scale)
		}
		f.CapHeight = f.Ascent
		f.StemV = 50 + int(math.Pow(float64(weight)/65.0, 2))
		return weight
	}

	f.seekTable("OS/2")
	version := f.readUint16()
	f.skip(2) // xAvgCharWidth
	weight := f.readUint16()
	f.skip(2) // widthClass
	fsType := f.readUint16()
	if fsType == 0x0002 || (fsType&0x0300) != 0 {
		// Font has embedding restrictions; we still parse metrics
	}
	f.skip(20) // ySubscript*, ySuperscript*, yStrikeout*, sFamilyClass
	f.readInt16() // skip panose + vendor + selection etc.
	f.skip(36)    // ulUnicodeRange, achVendID, fsSelection, usFirstChar, usLastChar
	sTypoAscender := f.readInt16()
	sTypoDescender := f.readInt16()
	if f.Ascent == 0 {
		f.Ascent = int(float64(sTypoAscender) * scale)
	}
	if f.Descent == 0 {
		f.Descent = int(float64(sTypoDescender) * scale)
	}
	if version > 1 {
		f.skip(16) // sTypoLineGap, usWinAscent, usWinDescent, ulCodePageRange1, ulCodePageRange2, sxHeight
		sCapHeight := f.readInt16()
		f.CapHeight = int(float64(sCapHeight) * scale)
	} else {
		f.CapHeight = f.Ascent
	}
	f.StemV = 50 + int(math.Pow(float64(weight)/65.0, 2))
	return weight
}

func (f *TTFFont) parsePost(weight int) {
	f.seekTable("post")
	f.skip(4) // version
	f.ItalicAngle = int(f.readInt16()) // integer part (ignoring fraction)
	f.readUint16()                     // fraction — consumed but not used
	scale := 1000.0 / float64(f.unitsPerEm)
	f.UnderlinePosition = float64(f.readInt16()) * scale
	f.UnderlineThickness = float64(f.readInt16()) * scale
	fixed := f.readUint32()

	f.Flags = 4 // Nonsymbolic
	if f.ItalicAngle != 0 {
		f.Flags |= 64 // Italic
	}
	if weight >= 600 {
		f.Flags |= 262144 // ForceBold
	}
	if fixed != 0 {
		f.Flags |= 1 // FixedPitch
	}
}

func (f *TTFFont) parseCmap() int {
	t := f.tables["cmap"]
	if t == nil {
		return 0
	}
	cmapOffset := t.offset
	f.seekTable("cmap")
	f.skip(2) // version
	numSubtables := f.readUint16()

	for range numSubtables {
		platformID := f.readUint16()
		encodingID := f.readUint16()
		offset := f.readUint32()
		saved := f.pos

		// Look for Microsoft Unicode (3,1) or Unicode (0,*)
		if (platformID == 3 && encodingID == 1) || platformID == 0 {
			format := f.getUint16(cmapOffset + offset)
			if format == 4 {
				return cmapOffset + offset
			}
		}
		f.pos = saved
	}
	return 0
}

func (f *TTFFont) buildCmapMappings(cmapPos int, glyphToChars map[int][]int) {
	f.pos = cmapPos + 2
	tableSize := f.readUint16()
	tableEnd := cmapPos + tableSize
	f.skip(2) // language

	segCount := f.readUint16() / 2
	f.skip(6) // searchRange, entrySelector, rangeShift

	endCodes := make([]int, segCount)
	for i := range segCount {
		endCodes[i] = f.readUint16()
	}
	f.skip(2) // reservedPad

	startCodes := make([]int, segCount)
	for i := range segCount {
		startCodes[i] = f.readUint16()
	}

	idDeltas := make([]int, segCount)
	for i := range segCount {
		idDeltas[i] = int(int16(f.readUint16Signed()))
	}

	idRangeOffsetBase := f.pos
	idRangeOffsets := make([]int, segCount)
	for i := range segCount {
		idRangeOffsets[i] = f.readUint16()
	}

	for seg := range segCount {
		for ch := startCodes[seg]; ch <= endCodes[seg]; ch++ {
			var glyph int
			if idRangeOffsets[seg] == 0 {
				glyph = (ch + idDeltas[seg]) & 0xFFFF
			} else {
				pos := (ch-startCodes[seg])*2 + idRangeOffsets[seg]
				pos = idRangeOffsetBase + 2*seg + pos
				if pos >= tableEnd {
					glyph = 0
				} else {
					glyph = f.getUint16(pos)
					if glyph != 0 {
						glyph = (glyph + idDeltas[seg]) & 0xFFFF
					}
				}
			}
			f.CharToGlyph[ch] = glyph
			if ch < 196608 {
				glyphToChars[glyph] = append(glyphToChars[glyph], ch)
			}
		}
	}
}

func (f *TTFFont) parseHmtx(numHMetrics, numGlyphs int, glyphToChars map[int][]int, scale float64) {
	start := f.tables["hmtx"].offset
	f.CharWidths = make([]int, 256*256)

	// Read all hmtx entries as uint16 pairs (1-indexed array from unpackU16Array)
	data := f.getRange(start, numHMetrics*4)
	arr := unpackU16Array(data)

	var lastAdvanceWidth int
	for glyph := range numHMetrics {
		advWidth := arr[(glyph*2)+1]
		lastAdvanceWidth = advWidth
		if advWidth >= (1 << 15) {
			advWidth = 0
		}

		if glyph == 0 {
			f.DefaultWidth = scale * float64(advWidth)
			continue
		}

		if chars, ok := glyphToChars[glyph]; ok {
			for _, ch := range chars {
				if ch != 0 && ch != 65535 && ch < 196608 {
					w := int(math.Round(scale * float64(advWidth)))
					f.CharWidths[ch] = w
				}
			}
		}
	}

	// Glyphs beyond numHMetrics share the last advanceWidth
	diff := numGlyphs - numHMetrics
	for i := range diff {
		glyph := i + numHMetrics
		if chars, ok := glyphToChars[glyph]; ok {
			for _, ch := range chars {
				if ch != 0 && ch != 65535 && ch < 196608 {
					w := int(math.Round(scale * float64(lastAdvanceWidth)))
					f.CharWidths[ch] = w
				}
			}
		}
	}

	// Fill missing widths with default
	defaultWidth := int(math.Round(f.DefaultWidth))
	if defaultWidth == 0 {
		defaultWidth = 500
	}
	for i := 1; i < len(f.CharWidths); i++ {
		if f.CharWidths[i] == 0 {
			f.CharWidths[i] = defaultWidth
		}
	}

	// Zero out widths for Unicode non-spacing combining marks (Mn category).
	// These characters (tone marks, diacritics, etc.) are positioned over/under
	// base characters and should not advance the cursor during text layout.
	for i := 1; i < len(f.CharWidths); i++ {
		if unicode.Is(unicode.Mn, rune(i)) {
			f.CharWidths[i] = 0
		}
	}
}

// Subset creates a subset TrueType font containing only the glyphs
// referenced by the given runes. It preserves original glyph IDs so
// that GPOS/GSUB table references (critical for Thai mark positioning)
// remain valid.
//
// usedRunes maps arbitrary keys to Unicode code points.
// Returns the subset font bytes and a mapping of unicode -> glyph ID.
func (f *TTFFont) Subset(usedRunes map[int]int) ([]byte, map[int]int) {
	// Re-parse from raw bytes for a clean state
	sf := &TTFFont{
		raw:         f.raw,
		tables:      make(map[string]*ttfTable),
		CharToGlyph: make(map[int]int),
	}
	sf.pos = 0
	sf.skip(4) // signature
	sf.readTableDirectory()

	// Read head for loca format
	sf.seekTable("head")
	sf.skip(18)
	sf.unitsPerEm = sf.readUint16()
	sf.seekTable("head")
	sf.skip(50)
	locaFormat := sf.readUint16()

	// Read hhea for metrics count
	sf.seekTable("hhea")
	sf.skip(34)
	metricsCount := sf.readUint16()
	origMetrics := metricsCount

	// Read maxp for numGlyphs
	sf.seekTable("maxp")
	sf.skip(4)
	numGlyphs := sf.readUint16()

	// Build cmap
	glyphToChars := make(map[int][]int)
	cmapPos := sf.findCmapFormat4()
	if cmapPos == 0 {
		return nil, nil
	}
	sf.buildCmapMappings(cmapPos, glyphToChars)

	// Parse hmtx (at scale 1.0 for raw widths in subsetting)
	sf.CharWidths = make([]int, 256*256)
	sf.parseHmtx(metricsCount, numGlyphs, glyphToChars, 1.0)

	// Parse loca
	sf.parseLoca(locaFormat, numGlyphs)

	// Determine which glyphs are needed
	glyphSet := map[int]bool{0: true}
	runeToGlyph := make(map[int]int)
	lastRune := 0
	for _, ch := range usedRunes {
		if gid, ok := sf.CharToGlyph[ch]; ok {
			glyphSet[gid] = true
			runeToGlyph[ch] = gid
		}
		if ch > lastRune {
			lastRune = ch
		}
	}

	// Scan composite glyphs to include their components
	glyfData := sf.getTableData("glyf")
	if glyfData == nil {
		return nil, nil
	}
	sf.addCompositeGlyphs(glyphSet, glyfData)

	// Build output tables map
	outTables := make(map[string][]byte)

	setTable := func(tag string, data []byte) {
		if data == nil {
			return
		}
		if tag == "head" {
			// Zero out checksumAdjustment
			d := make([]byte, len(data))
			copy(d, data)
			copy(d[8:12], []byte{0, 0, 0, 0})
			data = d
		}
		outTables[tag] = data
	}

	// Copy tables as-is
	setTable("name", sf.getTableData("name"))
	setTable("cvt ", sf.getTableData("cvt "))
	setTable("fpgm", sf.getTableData("fpgm"))
	setTable("prep", sf.getTableData("prep"))
	setTable("gasp", sf.getTableData("gasp"))

	// Preserve GPOS/GSUB for Thai combining marks
	setTable("GPOS", sf.getTableData("GPOS"))
	setTable("GSUB", sf.getTableData("GSUB"))
	setTable("GDEF", sf.getTableData("GDEF"))

	// Post table: force format 3 (no glyph names)
	postData := sf.getTableData("post")
	if len(postData) >= 16 {
		newPost := make([]byte, 32)
		newPost[0] = 0x00
		newPost[1] = 0x03
		newPost[2] = 0x00
		newPost[3] = 0x00
		copy(newPost[4:16], postData[4:16])
		setTable("post", newPost)
	}

	// Build cmap for subset
	cidToGlyph := make(map[int]int)
	for ch, gid := range runeToGlyph {
		if ch != 0 {
			cidToGlyph[ch] = gid
		}
	}
	setTable("cmap", buildCmapFormat4(cidToGlyph, numGlyphs))

	// Build glyf, loca, and hmtx for all glyph slots (preserving IDs)
	metricsCount = numGlyphs
	offsets := make([]int, 0, numGlyphs+1)
	var newGlyf []byte
	var newHmtx []byte
	pos := 0

	for gid := range numGlyphs {
		hm := sf.getHmtxEntry(origMetrics, gid)
		newHmtx = append(newHmtx, hm...)
		offsets = append(offsets, pos)

		if !glyphSet[gid] {
			// Unused glyph: empty entry
			continue
		}

		glyphOff := sf.glyphPos[gid]
		glyphLen := sf.glyphPos[gid+1] - glyphOff
		if glyphLen == 0 {
			continue
		}

		data := make([]byte, glyphLen)
		copy(data, glyfData[glyphOff:glyphOff+glyphLen])

		// Update composite glyph references (identity mapping preserves IDs)
		// No remapping needed since we preserve original glyph IDs

		newGlyf = append(newGlyf, data...)
		pos += glyphLen
		if pos%4 != 0 {
			pad := 4 - (pos % 4)
			newGlyf = append(newGlyf, make([]byte, pad)...)
			pos += pad
		}
	}
	offsets = append(offsets, pos)
	setTable("glyf", newGlyf)
	setTable("hmtx", newHmtx)

	// Build loca
	var locaData []byte
	if ((pos + 1) >> 1) > 0xFFFF {
		locaFormat = 1
		for _, off := range offsets {
			locaData = appendUint32(locaData, off)
		}
	} else {
		locaFormat = 0
		for _, off := range offsets {
			locaData = appendUint16(locaData, off/2)
		}
	}
	setTable("loca", locaData)

	// Update head with loca format
	headData := sf.getTableData("head")
	if headData != nil {
		headData = sliceInsertUint16(headData, 50, locaFormat)
		setTable("head", headData)
	}

	// Update hhea with metrics count
	hheaData := sf.getTableData("hhea")
	if hheaData != nil {
		hheaData = sliceInsertUint16(hheaData, 34, metricsCount)
		setTable("hhea", hheaData)
	}

	// Update maxp with numGlyphs
	maxpData := sf.getTableData("maxp")
	if maxpData != nil {
		maxpData = sliceInsertUint16(maxpData, 4, numGlyphs)
		setTable("maxp", maxpData)
	}

	// OS/2
	setTable("OS/2", sf.getTableData("OS/2"))

	return assembleTTF(outTables), runeToGlyph
}

// --- Subset helpers ---

func (f *TTFFont) readTableDirectory() {
	numTables := f.readUint16()
	f.skip(6)
	for range numTables {
		tag := string(f.readBytes(4))
		cs0 := f.readUint16()
		cs1 := f.readUint16()
		offset := f.readUint32()
		size := f.readUint32()
		f.tables[tag] = &ttfTable{
			offset:   offset,
			size:     size,
			checksum: [2]int{cs0, cs1},
		}
	}
}

func (f *TTFFont) findCmapFormat4() int {
	t := f.tables["cmap"]
	if t == nil {
		return 0
	}
	cmapOffset := t.offset
	f.seekTable("cmap")
	f.skip(2)
	numSub := f.readUint16()
	for range numSub {
		pid := f.readUint16()
		eid := f.readUint16()
		off := f.readUint32()
		saved := f.pos
		if (pid == 3 && eid == 1) || pid == 0 {
			fmt := f.getUint16(cmapOffset + off)
			if fmt == 4 {
				return cmapOffset + off
			}
		}
		f.pos = saved
	}
	return 0
}

func (f *TTFFont) parseLoca(format, numGlyphs int) {
	start := f.tables["loca"].offset
	f.glyphPos = make([]int, 0, numGlyphs+2)

	if format == 0 {
		data := f.getRange(start, (numGlyphs*2)+2)
		arr := unpackU16Array(data)
		for i := 0; i <= numGlyphs; i++ {
			f.glyphPos = append(f.glyphPos, arr[i+1]*2)
		}
	} else {
		data := f.getRange(start, (numGlyphs*4)+4)
		arr := unpackU32Array(data)
		for i := 0; i <= numGlyphs; i++ {
			f.glyphPos = append(f.glyphPos, arr[i+1])
		}
	}
}

func (f *TTFFont) addCompositeGlyphs(glyphSet map[int]bool, glyfData []byte) {
	// Iterate until no new glyphs are added
	for {
		added := false
		for gid := range glyphSet {
			if gid >= len(f.glyphPos)-1 {
				continue
			}
			off := f.glyphPos[gid]
			size := f.glyphPos[gid+1] - off
			if size < 2 {
				continue
			}
			numContours := int(int16(binary.BigEndian.Uint16(glyfData[off : off+2])))
			if numContours >= 0 {
				continue // simple glyph
			}
			// Composite glyph: scan components
			p := off + 10 // skip header
			flags := int(glyfMoreComps)
			for (flags & glyfMoreComps) != 0 {
				if p+4 > off+size {
					break
				}
				flags = int(binary.BigEndian.Uint16(glyfData[p : p+2]))
				compGid := int(binary.BigEndian.Uint16(glyfData[p+2 : p+4]))
				if !glyphSet[compGid] {
					glyphSet[compGid] = true
					added = true
				}
				p += 4
				if (flags & glyfArgWords) != 0 {
					p += 4
				} else {
					p += 2
				}
				if (flags & glyfArgScale) != 0 {
					p += 2
				} else if (flags & glyfXYScale) != 0 {
					p += 4
				} else if (flags & glyfTwoByTwo) != 0 {
					p += 8
				}
			}
		}
		if !added {
			break
		}
	}
}

func (f *TTFFont) getHmtxEntry(metricCount, gid int) []byte {
	start := f.tables["hmtx"].offset
	if gid < metricCount {
		return f.getRange(start+(gid*4), 4)
	}
	// Beyond metricCount: use last advance width + glyph's own LSB
	aw := f.getRange(start+((metricCount-1)*4), 2)
	result := make([]byte, 4)
	copy(result[:2], aw)
	lsb := f.getRange(start+(metricCount*4)+((gid-metricCount)*2), 2)
	copy(result[2:], lsb)
	return result
}

// --- Binary reading helpers ---

func (f *TTFFont) readBytes(n int) []byte {
	b := f.raw[f.pos : f.pos+n]
	f.pos += n
	return b
}

func (f *TTFFont) readUint16() int {
	b := f.raw[f.pos : f.pos+2]
	f.pos += 2
	return int(binary.BigEndian.Uint16(b))
}

func (f *TTFFont) readUint16Signed() uint16 {
	b := f.raw[f.pos : f.pos+2]
	f.pos += 2
	return binary.BigEndian.Uint16(b)
}

func (f *TTFFont) readInt16() int16 {
	b := f.raw[f.pos : f.pos+2]
	f.pos += 2
	return int16(binary.BigEndian.Uint16(b))
}

func (f *TTFFont) readUint32() int {
	b := f.raw[f.pos : f.pos+4]
	f.pos += 4
	return int(binary.BigEndian.Uint32(b))
}

func (f *TTFFont) skip(n int) {
	f.pos += n
}

func (f *TTFFont) seekTable(name string) {
	f.pos = f.tables[name].offset
}

func (f *TTFFont) getUint16(pos int) int {
	return int(binary.BigEndian.Uint16(f.raw[pos : pos+2]))
}

func (f *TTFFont) getRange(pos, length int) []byte {
	if length < 1 {
		return nil
	}
	result := make([]byte, length)
	copy(result, f.raw[pos:pos+length])
	return result
}

func (f *TTFFont) getTableData(name string) []byte {
	t := f.tables[name]
	if t == nil || t.size == 0 {
		return nil
	}
	return f.getRange(t.offset, t.size)
}

// --- Cmap generation ---

func buildCmapFormat4(cidToGlyph map[int]int, numGlyphs int) []byte {
	// Build sorted runs
	sortedCids := make([]int, 0, len(cidToGlyph))
	for cid := range cidToGlyph {
		sortedCids = append(sortedCids, cid)
	}
	sort.Ints(sortedCids)

	// Group into segments
	type segment struct {
		start int
		glyphs []int
	}
	var segs []segment
	prevCid := -2
	prevGlyph := -1
	segIdx := -1
	for _, cid := range sortedCids {
		glyph := cidToGlyph[cid]
		if cid == prevCid+1 && glyph == prevGlyph+1 {
			segs[segIdx].glyphs = append(segs[segIdx].glyphs, glyph)
		} else {
			segs = append(segs, segment{start: cid, glyphs: []int{glyph}})
			segIdx++
		}
		prevCid = cid
		prevGlyph = glyph
	}

	segCount := len(segs) + 1 // +1 for 0xFFFF sentinel

	searchRange := 1
	entrySelector := 0
	for searchRange*2 <= segCount {
		searchRange *= 2
		entrySelector++
	}
	searchRange *= 2
	rangeShift := segCount*2 - searchRange

	// Platform 3, encoding 1, offset 12
	length := 16 + (8 * segCount) + (numGlyphs + 1)
	header := []int{
		0, 1,        // version, numTables
		3, 1, 0, 12, // platformID=3, encodingID=1, offset=12
		4, length, 0, // format=4, length, language
		segCount * 2, searchRange, entrySelector, rangeShift,
	}

	var cmap []int
	cmap = append(cmap, header...)

	// endCode
	for _, s := range segs {
		cmap = append(cmap, s.start+len(s.glyphs)-1)
	}
	cmap = append(cmap, 0xFFFF)
	cmap = append(cmap, 0) // reservedPad

	// startCode
	for _, s := range segs {
		cmap = append(cmap, s.start)
	}
	cmap = append(cmap, 0xFFFF)

	// idDelta
	for _, s := range segs {
		delta := -(s.start - s.glyphs[0])
		cmap = append(cmap, delta&0xFFFF)
	}
	cmap = append(cmap, 1)

	// idRangeOffset (all zeros, using delta)
	for range segs {
		cmap = append(cmap, 0)
	}
	cmap = append(cmap, 0)

	// glyphIdArray
	for _, s := range segs {
		cmap = append(cmap, s.glyphs...)
	}
	cmap = append(cmap, 0)

	// Pack to bytes
	var buf []byte
	for _, v := range cmap {
		buf = appendUint16(buf, v)
	}
	return buf
}

// --- TTF assembly ---

func assembleTTF(tables map[string][]byte) []byte {
	numTables := len(tables)

	searchRange := 1
	entrySelector := 0
	for searchRange*2 <= numTables {
		searchRange *= 2
		entrySelector++
	}
	searchRange *= 16
	rangeShift := numTables*16 - searchRange

	// File header
	var out []byte
	out = appendUint32(out, 0x00010000) // sfVersion
	out = appendUint16(out, numTables)
	out = appendUint16(out, searchRange)
	out = appendUint16(out, entrySelector)
	out = appendUint16(out, rangeShift)

	// Sort table tags
	tags := make([]string, 0, numTables)
	for tag := range tables {
		tags = append(tags, tag)
	}
	sort.Strings(tags)

	// Calculate data offset (after header + directory)
	dataOffset := 12 + numTables*16
	headOffset := 0

	// Write directory entries
	for _, tag := range tags {
		data := tables[tag]
		if tag == "head" {
			headOffset = dataOffset
		}
		cs := checksumTTF(data)
		out = append(out, []byte(tag)...)
		out = appendUint16(out, cs[0])
		out = appendUint16(out, cs[1])
		out = appendUint32(out, dataOffset)
		out = appendUint32(out, len(data))
		paddedLen := (len(data) + 3) &^ 3
		dataOffset += paddedLen
	}

	// Write table data
	for _, tag := range tags {
		data := tables[tag]
		out = append(out, data...)
		// Pad to 4-byte boundary
		if rem := len(data) % 4; rem != 0 {
			out = append(out, make([]byte, 4-rem)...)
		}
	}

	// Fix checksum adjustment in head table
	if headOffset > 0 {
		wholeChecksum := checksumTTF(out)
		adj := calcInt32Sub([2]int{0xB1B0, 0xAFBA}, wholeChecksum)
		binary.BigEndian.PutUint16(out[headOffset+8:], uint16(adj[0]))
		binary.BigEndian.PutUint16(out[headOffset+10:], uint16(adj[1]))
	}

	return out
}

// --- Pack/unpack utilities ---

func appendUint16(buf []byte, v int) []byte {
	return append(buf, byte(v>>8), byte(v))
}

func appendUint32(buf []byte, v int) []byte {
	return append(buf, byte(v>>24), byte(v>>16), byte(v>>8), byte(v))
}

func sliceInsertUint16(data []byte, offset, value int) []byte {
	result := make([]byte, len(data))
	copy(result, data)
	binary.BigEndian.PutUint16(result[offset:], uint16(value))
	return result
}

func unpackU16Array(data []byte) []int {
	result := make([]int, 1) // 1-indexed to match gofpdf convention
	r := bytes.NewReader(data)
	var buf [2]byte
	for {
		n, err := r.Read(buf[:])
		if n < 2 || err != nil {
			break
		}
		result = append(result, int(binary.BigEndian.Uint16(buf[:])))
	}
	return result
}

func unpackU32Array(data []byte) []int {
	result := make([]int, 1) // 1-indexed
	r := bytes.NewReader(data)
	var buf [4]byte
	for {
		n, err := r.Read(buf[:])
		if n < 4 || err != nil {
			break
		}
		result = append(result, int(binary.BigEndian.Uint32(buf[:])))
	}
	return result
}

func checksumTTF(data []byte) [2]int {
	// Pad to 4-byte boundary
	padded := data
	if len(data)%4 != 0 {
		padded = make([]byte, len(data)+4-(len(data)%4))
		copy(padded, data)
	}
	var a [2]int
	for i := 0; i < len(padded); i += 4 {
		a[0] += (int(padded[i]) << 8) + int(padded[i+1])
		a[1] += (int(padded[i+2]) << 8) + int(padded[i+3])
		a[0] += a[1] >> 16
		a[1] &= 0xFFFF
		a[0] &= 0xFFFF
	}
	return a
}

func calcInt32Sub(x, y [2]int) [2]int {
	a := x
	if y[1] > a[1] {
		a[1] += 1 << 16
		a[0]++
	}
	a[1] -= y[1]
	if y[0] > a[0] {
		a[0] += 1 << 16
	}
	a[0] -= y[0]
	a[0] &= 0xFFFF
	return a
}

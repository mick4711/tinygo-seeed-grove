package ssd1306

// Bitmap font support. Fonts are generated offline (see the fonts
// subpackage); the runtime side below only blits precomputed page-major
// glyph bitmaps, following the split u8g2 makes between font generation
// and drawing: https://github.com/olikraus/u8g2/wiki/u8g2fontformat

const glyphRecordSize = 4 // offset(2, big endian) | width(1) | advance(1)

// Font is a fixed-height bitmap font generated offline. All glyph data is
// held in strings, which TinyGo keeps in flash: a Font costs no RAM beyond
// the struct header. Construct with NewFont.
type Font struct {
	// Height of every glyph in pixels.
	Height uint8
	// bitmaps holds page-major glyph pixels: for each glyph, all columns of
	// page 0 left-to-right, then page 1, ... Bit k of a byte is glyph row
	// page*8+k, the same layout as the SSD1306 display buffer. Rows past
	// Height are zero padding.
	bitmaps string
	// glyphs holds one glyphRecordSize record per glyph.
	glyphs string
	// ranges maps runes to glyph records; see GlyphRange.
	ranges []GlyphRange
}

// GlyphRange is a run of consecutive runes mapped to consecutive glyph
// records. An ASCII font has exactly one: {' ', 0, 95}.
type GlyphRange struct {
	FirstRune  rune
	GlyphStart uint16
	Count      uint16
}

// NewFont builds a Font from generated data, panicking on malformed input
// since that can only come from a broken generator, not from runtime state.
func NewFont(height uint8, bitmaps, glyphs string, ranges []GlyphRange) Font {
	f := Font{Height: height, bitmaps: bitmaps, glyphs: glyphs, ranges: ranges}
	nglyphs := len(glyphs) / glyphRecordSize
	if height == 0 || len(glyphs) == 0 || len(glyphs)%glyphRecordSize != 0 {
		panic("ssd1306: bad font glyph table")
	}
	for gi := 0; gi < nglyphs; gi++ {
		off, width, _ := f.glyphAt(gi)
		if off+width*f.pages() > len(bitmaps) {
			panic("ssd1306: font glyph bitmap out of range")
		}
	}
	for _, rg := range ranges {
		if rg.Count == 0 || int(rg.GlyphStart)+int(rg.Count) > nglyphs {
			panic("ssd1306: bad font rune range")
		}
	}
	return f
}

func (f *Font) pages() int {
	return (int(f.Height) + 7) / 8
}

func (f *Font) glyphAt(gi int) (offset, width, advance int) {
	rec := f.glyphs[gi*glyphRecordSize : gi*glyphRecordSize+glyphRecordSize]
	return int(rec[0])<<8 | int(rec[1]), int(rec[2]), int(rec[3])
}

// glyph looks r up in the rune ranges. Unknown runes fall back to the first
// glyph, conventionally a blank, so they advance without drawing garbage.
func (f *Font) glyph(r rune) (offset, width, advance int) {
	gi := 0
	for _, rg := range f.ranges {
		if r >= rg.FirstRune && r < rg.FirstRune+rune(rg.Count) {
			gi = int(rg.GlyphStart) + int(r-rg.FirstRune)
			break
		}
	}
	return f.glyphAt(gi)
}

// DrawText blits s with its top-left corner at (x, y). The background is
// transparent: on=true only sets lit glyph pixels, on=false only clears
// them. Returns the x coordinate after the last glyph's advance, for
// chaining calls. On error the glyphs before the offending one stay drawn.
func DrawText(displayBuf []byte, x, y int, s string, f *Font, displayWidth int, on bool) (int, error) {
	if y < 0 {
		return x, errOutOfRange
	}
	pages := f.pages()
	shift := uint(y % 8)
	for _, r := range s {
		off, gw, adv := f.glyph(r)
		if x < 0 || x+gw > displayWidth {
			return x, errOutOfRange
		}
		// All glyph pixels fall inside the width x Height box, whose largest
		// byte index is at the bottom-right corner; one check per glyph.
		if BufferPixelByteOff(x+gw-1, y+int(f.Height)-1, displayWidth) >= len(displayBuf) {
			return x, errShortBuffer
		}
		for p := 0; p < pages; p++ {
			src := f.bitmaps[off+p*gw : off+(p+1)*gw]
			idx := BufferPixelByteOff(x, y+p*8, displayWidth)
			if shift == 0 {
				dst := displayBuf[idx : idx+gw]
				if on {
					for i := 0; i < gw; i++ {
						dst[i] |= src[i]
					}
				} else {
					for i := 0; i < gw; i++ {
						dst[i] &^= src[i]
					}
				}
				continue
			}
			// Unaligned: each glyph byte splits across two page rows. The
			// high spill can only be nonzero when the glyph has lit pixels
			// in rows covered by the bounds check above, so guarding on it
			// keeps the second write in bounds.
			for i := 0; i < gw; i++ {
				lo := src[i] << shift
				hi := src[i] >> (8 - shift)
				if on {
					displayBuf[idx+i] |= lo
					if hi != 0 {
						displayBuf[idx+i+displayWidth] |= hi
					}
				} else {
					displayBuf[idx+i] &^= lo
					if hi != 0 {
						displayBuf[idx+i+displayWidth] &^= hi
					}
				}
			}
		}
		x += adv
	}
	return x, nil
}

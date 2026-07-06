// Package fonts provides fixed-height bitmap fonts for monochrome displays.
//
// Glyph bitmaps are stored page-major: for each glyph, all columns of page 0
// left-to-right, then page 1, and so on, where bit k of a byte is glyph row
// page*8+k. This is the native framebuffer layout of the SSD1306 display
// family (SSD1306, SSD1309, SH1106, ST7567, ...), so those drivers can blit
// glyphs with per-column OR operations. Other displays can still consume the
// fonts by decoding the bitmaps pixel by pixel.
//
// Fonts are generated offline, following the split u8g2 makes between font
// generation and drawing: https://github.com/olikraus/u8g2/wiki/u8g2fontformat
package fonts

const glyphRecordSize = 4 // offset(2, big endian) | width(1) | advance(1)

// Font is a fixed-height bitmap font generated offline. All glyph data is
// held in strings, which TinyGo keeps in flash: a Font costs no RAM beyond
// the struct header. Construct with NewFont.
type Font struct {
	// height of every glyph in pixels.
	height uint8
	// bitmaps holds the page-major glyph pixels described in the package
	// comment. Rows past Height are zero padding.
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
	f := Font{height: height, bitmaps: bitmaps, glyphs: glyphs, ranges: ranges}
	nglyphs := len(glyphs) / glyphRecordSize
	if height == 0 || len(glyphs) == 0 || len(glyphs)%glyphRecordSize != 0 {
		panic("fonts: bad font glyph table")
	}
	for gi := 0; gi < nglyphs; gi++ {
		off, width, _ := f.glyphAt(gi)
		if off+width*f.Pages() > len(bitmaps) {
			panic("fonts: font glyph bitmap out of range")
		}
	}
	for _, rg := range ranges {
		if rg.Count == 0 || int(rg.GlyphStart)+int(rg.Count) > nglyphs {
			panic("fonts: bad font rune range")
		}
	}
	return f
}

// Height returns the pixel height of every glyph.
func (f *Font) Height() int {
	return int(f.height)
}

// Pages returns the number of 8 pixel pages spanned by the font height.
func (f *Font) Pages() int {
	return (int(f.height) + 7) / 8
}

func (f *Font) glyphAt(gi int) (offset, width, advance int) {
	rec := f.glyphs[gi*glyphRecordSize : gi*glyphRecordSize+glyphRecordSize]
	return int(rec[0])<<8 | int(rec[1]), int(rec[2]), int(rec[3])
}

// Glyph returns r's bitmap (Pages()*width bytes, page-major as described in
// the package comment), pixel width, and horizontal advance. Unknown runes
// fall back to the first glyph, conventionally a blank, so they advance
// without drawing garbage.
func (f *Font) Glyph(r rune) (bitmap string, width, advance int) {
	gi := 0
	for _, rg := range f.ranges {
		if r >= rg.FirstRune && r < rg.FirstRune+rune(rg.Count) {
			gi = int(rg.GlyphStart) + int(r-rg.FirstRune)
			break
		}
	}
	off, width, advance := f.glyphAt(gi)
	return f.bitmaps[off : off+width*f.Pages()], width, advance
}

// StringWidth returns the width in pixels a driver advances when drawing s,
// including the advance after the last glyph.
func (f *Font) StringWidth(s string) (w int) {
	for _, r := range s {
		_, _, adv := f.Glyph(r)
		w += adv
	}
	return w
}

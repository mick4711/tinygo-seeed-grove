package ssd1306

// Text drawing. Fonts come from the drivers/fonts package, whose page-major
// bitmap layout matches the SSD1306 display buffer, so glyphs blit with
// per-column OR operations.

import "github.com/soypat/seeed-grove/drivers/fonts"

// DrawText blits s with its top-left corner at (x, y). The background is
// transparent: on=true only sets lit glyph pixels, on=false only clears
// them. Returns the x coordinate after the last glyph's advance, for
// chaining calls. On error the glyphs before the offending one stay drawn.
func DrawText(displayBuf []byte, x, y int, s string, f *fonts.Font, displayWidth int, on bool) (int, error) {
	if y < 0 {
		return x, errOutOfRange
	}
	pages := f.Pages()
	shift := uint(y % 8)
	for _, r := range s {
		bm, gw, adv := f.Glyph(r)
		if x < 0 || x+gw > displayWidth {
			return x, errOutOfRange
		}
		// All glyph pixels fall inside the width x Height box, whose largest
		// byte index is at the bottom-right corner; one check per glyph.
		if BufferPixelByteOff(x+gw-1, y+f.Height()-1, displayWidth) >= len(displayBuf) {
			return x, errShortBuffer
		}
		for p := 0; p < pages; p++ {
			src := bm[p*gw : (p+1)*gw]
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

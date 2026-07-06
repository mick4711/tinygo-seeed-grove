package ssd1306

import (
	"testing"

	"github.com/soypat/seeed-grove/drivers/fonts"
)

// testFont has one 3x16 glyph (two pages) mapped to '0' so multi-page and
// unaligned blits can be checked against a per-pixel reference.
var testFont = fonts.NewFont(16,
	"\xff\x0f\xaa"+ // page 0
		"\x01\xf0\x55", // page 1
	"\x00\x00\x03\x04", // offset 0, width 3, advance 4
	[]fonts.GlyphRange{{FirstRune: '0', GlyphStart: 0, Count: 1}},
)

// testFontPixel reports the reference value of glyph pixel (i, j) straight
// from the bitmap string.
func testFontPixel(i, j int) bool {
	bm, _, _ := testFont.Glyph('0')
	return bm[(j/8)*3+i]&(1<<(j%8)) != 0
}

func TestDrawTextAgainstReference(t *testing.T) {
	const w, h = 32, 32
	for _, y := range []int{0, 5, 8, 13} { // aligned and unaligned starts
		buf := make([]byte, BufferSize(w, h))
		want := make([]byte, BufferSize(w, h))
		const x = 2
		xEnd, err := DrawText(buf, x, y, "0", &testFont, w, true)
		if err != nil {
			t.Fatalf("y=%d: %v", y, err)
		}
		if xEnd != x+4 {
			t.Errorf("y=%d: advance got %d want %d", y, xEnd, x+4)
		}
		for i := 0; i < 3; i++ {
			for j := 0; j < 16; j++ {
				if testFontPixel(i, j) {
					setPixel(want, x+i, y+j, w, true)
				}
			}
		}
		for i := range buf {
			if buf[i] != want[i] {
				t.Errorf("y=%d: byte %d got %#02x want %#02x", y, i, buf[i], want[i])
			}
		}
		// Erasing the same text over a full pattern clears exactly the glyph bits.
		for i := 1; i < len(buf); i++ {
			buf[i] = 0xff
		}
		DrawText(buf, x, y, "0", &testFont, w, false)
		for i := 1; i < len(buf); i++ {
			if buf[i] != 0xff&^want[i] {
				t.Errorf("y=%d erase: byte %d got %#02x want %#02x", y, i, buf[i], 0xff&^want[i])
			}
		}
	}
}

func TestDrawTextFallbackAndErrors(t *testing.T) {
	const w, h = 32, 32
	buf := make([]byte, BufferSize(w, h))
	// Unknown rune falls back to glyph 0 and still advances.
	xEnd, err := DrawText(buf, 0, 0, "Z", &testFont, w, true)
	if err != nil || xEnd != 4 {
		t.Errorf("fallback: xEnd=%d err=%v", xEnd, err)
	}
	if buf[1] != 0xff { // glyph 0 page 0 column 0
		t.Errorf("fallback did not draw glyph 0: %#02x", buf[1])
	}
	for _, tc := range []struct{ x, y int }{{-1, 0}, {0, -1}, {30, 0}} {
		if _, err := DrawText(buf, tc.x, tc.y, "0", &testFont, w, true); err == nil {
			t.Errorf("(%d,%d): want error", tc.x, tc.y)
		}
	}
	// Second glyph out of range: first stays drawn, x reports progress.
	clear(buf)
	xEnd, err = DrawText(buf, 26, 0, "00", &testFont, w, true)
	if err == nil || xEnd != 30 {
		t.Errorf("partial: xEnd=%d err=%v", xEnd, err)
	}
	if buf[1+26] == 0 {
		t.Error("partial: first glyph missing")
	}
	if _, err := DrawText(buf[:5], 0, 0, "0", &testFont, w, true); err == nil {
		t.Error("short buffer: want error")
	}
}

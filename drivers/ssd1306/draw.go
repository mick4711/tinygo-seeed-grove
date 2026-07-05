package ssd1306

import "unicode/utf8"

type Drawer struct {
	DisplayWidth int
	Buffer       []byte
}

func (drw *Drawer) PatternSet(b byte) {
	for i := 1; i < len(drw.Buffer); i++ {
		drw.Buffer[i] = b
	}
}

func (drw *Drawer) Clear() {
	drw.PatternSet(0)
}

func (drw *Drawer) DrawPixel(x, y int, on bool) error {
	return DrawPixel(drw.Buffer, x, y, drw.DisplayWidth, on)
}

func (drw *Drawer) DrawRectangle(x, y, width, height int, on bool) error {
	return DrawRectangle(drw.Buffer, x, y, width, height, drw.DisplayWidth, on)
}

func BufferPixelByteOff(x int, y int, displayWidth int) int {
	return 1 + x + (y/8)*displayWidth // +1 for the data mode byte.
}

func PixelBitOff(y int) int {
	return y % 8
}

// DrawPixel turns the pixel at (x, y) on or off in the display buffer.
func DrawPixel(displayBuf []byte, x, y, displayWidth int, on bool) error {
	if x < 0 || y < 0 || x >= displayWidth {
		return errOutOfRange
	}
	idx := BufferPixelByteOff(x, y, displayWidth)
	if idx >= len(displayBuf) {
		return errShortBuffer
	}
	bit := uint8(1) << PixelBitOff(y)
	if on {
		displayBuf[idx] |= bit
	} else {
		displayBuf[idx] &^= bit
	}
	return nil
}

// DrawRectangle fills a rectangle at given coordinates, turning pixels on or off.
func DrawRectangle(displayBuf []byte, x, y, rectWidth, rectHeight, displayWidth int, pixelOn bool) error {
	if x < 0 || y < 0 || rectWidth <= 0 || rectHeight <= 0 ||
		x+rectWidth > displayWidth {
		return errOutOfRange
	}
	yEnd := y + rectHeight
	if BufferPixelByteOff(x+rectWidth-1, yEnd-1, displayWidth) >= len(displayBuf) {
		return errShortBuffer
	}
	var fillByte uint8
	if pixelOn {
		fillByte = 0xff
	}
	// Each buffer byte holds 8 vertically stacked pixels (one page row).
	// Full-mask pages are overwritten; partial top/bottom pages are masked.
	for page := y / 8; page*8 < yEnd; page++ {
		lo, hi := y-page*8, yEnd-page*8
		if lo < 0 {
			lo = 0
		}
		if hi > 8 {
			hi = 8
		}
		mask := uint8(1<<hi-1) &^ uint8(1<<lo-1)
		idx := BufferPixelByteOff(x, page*8, displayWidth)
		row := displayBuf[idx : idx+rectWidth]
		switch {
		case mask == 0xff:
			for i := range row {
				row[i] = fillByte
			}
		case pixelOn:
			for i := range row {
				row[i] |= mask
			}
		default:
			for i := range row {
				row[i] &^= mask
			}
		}
	}
	return nil
}

// AppendFormatDisplay appends a text rendering of the display buffer to dst
// and returns the extended buffer, one line per pixel row, using drawOn for
// lit pixels and drawOff for unlit pixels. Allocates only if dst lacks capacity.
func AppendFormatDisplay(dst, displaybuffer []byte, displayWidth int, drawOn, drawOff rune) []byte {
	if displayWidth <= 0 || len(displaybuffer) <= 1 {
		return dst
	}
	rows := (len(displaybuffer) - 1) / displayWidth * 8
	for y := 0; y < rows; y++ {
		bit := uint8(1) << PixelBitOff(y)
		rowOff := BufferPixelByteOff(0, y, displayWidth)
		for x := 0; x < displayWidth; x++ {
			r := drawOff
			if displaybuffer[rowOff+x]&bit != 0 {
				r = drawOn
			}
			dst = utf8.AppendRune(dst, r)
		}
		dst = append(dst, '\n')
	}
	return dst
}

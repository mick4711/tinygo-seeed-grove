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

func (drw *Drawer) DrawLine(x0, y0, x1, y1 int, on bool) error {
	return DrawLine(drw.Buffer, x0, y0, x1, y1, drw.DisplayWidth, on)
}

func (drw *Drawer) DrawCircle(x, y, radius int, on bool) error {
	return DrawCircle(drw.Buffer, x, y, radius, drw.DisplayWidth, on)
}

func (drw *Drawer) DrawFilledCircle(x, y, radius int, on bool) error {
	return DrawFilledCircle(drw.Buffer, x, y, radius, drw.DisplayWidth, on)
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
	if BufferPixelByteOff(x, y, displayWidth) >= len(displayBuf) {
		return errShortBuffer
	}
	setPixel(displayBuf, x, y, displayWidth, on)
	return nil
}

// setPixel writes a pixel without bounds checks; the caller must have
// validated that (x, y) falls inside the buffer.
func setPixel(displayBuf []byte, x, y, displayWidth int, on bool) {
	bit := uint8(1) << PixelBitOff(y)
	idx := BufferPixelByteOff(x, y, displayWidth)
	if on {
		displayBuf[idx] |= bit
	} else {
		displayBuf[idx] &^= bit
	}
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

// DrawLine draws a line between (x0, y0) and (x1, y1), both endpoints
// included, using Bresenham's algorithm as in u8g2's u8g2_DrawLine:
// https://github.com/olikraus/u8g2/blob/master/csrc/u8g2_line.c
func DrawLine(displayBuf []byte, x0, y0, x1, y1, displayWidth int, on bool) error {
	// Horizontal and vertical lines use the faster masked byte fills.
	if x0 == x1 || y0 == y1 {
		if x1 < x0 {
			x0, x1 = x1, x0
		}
		if y1 < y0 {
			y0, y1 = y1, y0
		}
		return DrawRectangle(displayBuf, x0, y0, x1-x0+1, y1-y0+1, displayWidth, on)
	}
	if x0 < 0 || y0 < 0 || x1 < 0 || y1 < 0 || x0 >= displayWidth || x1 >= displayWidth {
		return errOutOfRange
	}
	dx, dy := x1-x0, y1-y0
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	// All line pixels fall inside the endpoints' bounding box, whose largest
	// byte index is at (max x, max y), so one check covers the whole loop.
	if BufferPixelByteOff(max(x0, x1), max(y0, y1), displayWidth) >= len(displayBuf) {
		return errShortBuffer
	}
	// Transpose steep lines so the loop always steps along x.
	steep := dy > dx
	if steep {
		x0, y0 = y0, x0
		x1, y1 = y1, x1
		dx, dy = dy, dx
	}
	if x0 > x1 {
		x0, x1 = x1, x0
		y0, y1 = y1, y0
	}
	ystep := 1
	if y1 < y0 {
		ystep = -1
	}
	err := dx / 2
	for x, y := x0, y0; x <= x1; x++ {
		px, py := x, y
		if steep {
			px, py = y, x
		}
		setPixel(displayBuf, px, py, displayWidth, on)
		err -= dy
		if err < 0 {
			y += ystep
			err += dx
		}
	}
	return nil
}

// ShapeStepper advances the octant point (dx, dy) of an 8-fold symmetric
// shape from (radius, 0) towards the dx == dy diagonal, carrying the error
// accumulator t1 between steps. See DrawShape for the walking loop.
type ShapeStepper func(dx, dy, t1 int) (dxnext, dynext, t1next int)

// nextCircle steps along a circle using Jesko's variant of the midpoint
// circle algorithm:
// https://en.wikipedia.org/wiki/Midpoint_circle_algorithm#Jesko's_Method
func nextCircle(dx, dy, t1 int) (dxnext, dynext, t1next int) {
	dy++
	t1 += dy
	t2 := t1 - dx
	if t2 >= 0 {
		t1 = t2
		dx--
	}
	return dx, dy, t1
}

// DrawCircle draws the outline of a circle centered at (x, y).
func DrawCircle(displayBuf []byte, x, y, radius, displayWidth int, on bool) error {
	return DrawShape(displayBuf, x, y, radius, displayWidth, on, false, nextCircle)
}

// DrawFilledCircle draws a filled circle centered at (x, y).
func DrawFilledCircle(displayBuf []byte, x, y, radius, displayWidth int, on bool) error {
	return DrawShape(displayBuf, x, y, radius, displayWidth, on, true, nextCircle)
}

// DrawShape draws an 8-fold symmetric shape centered at (x, y), as outline or
// filled, by walking one octant from (radius, 0) with next and mirroring each
// point into the remaining octants. All shape points must satisfy
// 0 <= dy <= dx <= radius; a step outside that range stops the walk with an
// error before anything is drawn out of bounds.
//
// See https://gist.github.com/soypat/1253c460cac3a5e0c92862b879a02577 for more shape examples.
func DrawShape(displayBuf []byte, x, y, radius, displayWidth int, on, fill bool, next ShapeStepper) error {
	if radius < 0 || x-radius < 0 || y-radius < 0 || x+radius >= displayWidth {
		return errOutOfRange
	}
	// All shape pixels fall inside the bounding box, whose largest byte
	// index is at (x+radius, y+radius), so one check covers the whole loop.
	if BufferPixelByteOff(x+radius, y+radius, displayWidth) >= len(displayBuf) {
		return errShortBuffer
	}
	dx, dy := radius, 0
	t1 := radius / 16 // Initialize for nicer looking shapes (circle case).
	for dx >= dy {
		if fill {
			// Horizontal spans between mirrored octant points; overlapping
			// spans rewrite the same bit value, so overdraw is harmless.
			hspan(displayBuf, x-dx, x+dx, y+dy, displayWidth, on)
			hspan(displayBuf, x-dx, x+dx, y-dy, displayWidth, on)
			hspan(displayBuf, x-dy, x+dy, y+dx, displayWidth, on)
			hspan(displayBuf, x-dy, x+dy, y-dx, displayWidth, on)
		} else {
			// One point per octant, mirrored around the center.
			setPixel(displayBuf, x+dx, y+dy, displayWidth, on)
			setPixel(displayBuf, x-dx, y+dy, displayWidth, on)
			setPixel(displayBuf, x+dx, y-dy, displayWidth, on)
			setPixel(displayBuf, x-dx, y-dy, displayWidth, on)
			setPixel(displayBuf, x+dy, y+dx, displayWidth, on)
			setPixel(displayBuf, x-dy, y+dx, displayWidth, on)
			setPixel(displayBuf, x+dy, y-dx, displayWidth, on)
			setPixel(displayBuf, x-dy, y-dx, displayWidth, on)
		}
		dxprev, dyprev := dx, dy
		dx, dy, t1 = next(dx, dy, t1)
		if dx > radius || dy < 0 || (dx == dxprev && dy == dyprev) {
			return errBadStepper
		}
	}
	return nil
}

// hspan writes the horizontal pixel run [x0, x1] on row y without bounds
// checks; the caller must have validated the range against the buffer.
func hspan(displayBuf []byte, x0, x1, y, displayWidth int, on bool) {
	bit := uint8(1) << PixelBitOff(y)
	row := displayBuf[BufferPixelByteOff(x0, y, displayWidth) : BufferPixelByteOff(x1, y, displayWidth)+1]
	if on {
		for i := range row {
			row[i] |= bit
		}
	} else {
		for i := range row {
			row[i] &^= bit
		}
	}
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

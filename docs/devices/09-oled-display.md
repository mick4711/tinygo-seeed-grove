# 9. OLED display — the I2C bus

**bus**: a shared set of wires over which multiple devices communicate, each
reachable at its own address.

Every device so far had a wire to itself. That doesn't scale — a robot with
twenty sensors can't spend twenty pins. The grown-up answer is a bus, and the
one you'll meet most often is **I2C** (Inter-Integrated Circuit). Our first
I2C citizen is the fanciest module in the kit: a 64×48 pixel OLED display.

<!-- TODO: photo of Grove OLED display module -->

## The concept

I2C uses exactly two wires, no matter how many devices share them: **SDA**
(data) and **SCL** (clock). The microcontroller is the *controller*; every
attached chip is a *target* with a 7-bit address baked in. A conversation
starts with the controller broadcasting an address; only the matching device
listens, and it *acknowledges* — pulling the data line low to say "here". Two
wires buy us addressing, acknowledgement, and multiple devices. The cost is
speed: our bus runs at 400 kHz, leagues slower than the display's appetite.

The display itself is an SSD1306 controller wired to 3072 organic LEDs — one
per pixel, no backlight, which is why the black is so black. We drive it with
the frame buffer pattern from last chapter, one bit per pixel this time:
draw into a 384-byte buffer in RAM, then ship the whole thing over I2C with
`Display()`.

## Wiring

Plug the display into connector **4** (`shield.Conn(4)`). Connectors 3 and 4
are the shield's I2C sockets — and because I2C is a bus, both lead to the
*same* two pins. The white wire is SDA, yellow is SCL.

## The code

This example is bigger than the rest — it's a little demo reel. The hardware
interaction is all in `main`; every function below it is drawing logic.

<!-- code: examples/examples-grove-single/grove-oled-display/oled-display.go -->
```go
package main

import (
	"time"

	grove "github.com/soypat/seeed-grove"
	"github.com/soypat/seeed-grove/drivers/fonts"
	"github.com/soypat/seeed-grove/drivers/ssd1306"
)

const (
	ConnPosition  = 3
	width, height = 64, 48
	frameTime     = 30 * time.Millisecond
)

// Define the hardware we are using.
var (
	shield  grove.ShieldXiao
	display ssd1306.Device
	font    = &fonts.Font5x7
)

func main() {
	time.Sleep(time.Second) // wait for USB connection.
	err := display.ConfigureI2C(shield.Conn(ConnPosition).I2C(400_000), ssd1306.Config{
		Width:  width,
		Height: height,
		Buffer: make([]byte, ssd1306.BufferSize(width, height)),
	})
	if err != nil {
		panic(err)
	}
	drawer := ssd1306.Drawer{
		DisplayWidth: width,
		Buffer:       display.CurrentDisplayBuffer(),
	}
	for {
		splash(&drawer)
		bounce(&drawer)
		ping(&drawer)
		loading(&drawer)
	}
}

// splash shows a framed title card.
func splash(drw *ssd1306.Drawer) {
	drw.Clear()
	border(drw)
	centerText(drw, 6, "Grove")
	centerText(drw, 16, "OLED")
	// Ruler with dot endpoints under the title.
	drw.DrawLine(12, 28, width-13, 28, true)
	drw.DrawFilledCircle(9, 28, 1, true)
	drw.DrawFilledCircle(width-10, 28, 1, true)
	centerText(drw, 34, "64x48")
	display.Display()
	time.Sleep(2500 * time.Millisecond)
}

// bounce animates a ball reflecting off the frame above a caption.
func bounce(drw *ssd1306.Drawer) {
	const radius = 3
	const floor = 35 // separator line above the caption.
	x, y := 20, 12
	dx, dy := 1, 1
	for i := 0; i < 200; i++ {
		drw.Clear()
		border(drw)
		drw.DrawLine(1, floor, width-2, floor, true)
		centerText(drw, floor+3, "bounce!")
		drw.DrawFilledCircle(x, y, radius, true)
		display.Display()

		x, dx = step(x, dx, 1+radius+1, width-2-radius-1)
		y, dy = step(y, dy, 1+radius+1, floor-radius-2)
		time.Sleep(frameTime)
	}
}

// ping expands concentric circles from the center, sonar style.
func ping(drw *ssd1306.Drawer) {
	const cx, cy, maxRadius, spacing = width / 2, 19, 14, 5
	for i := 0; i < 5; i++ {
		for r := 1; r <= maxRadius; r++ {
			drw.Clear()
			border(drw)
			drw.DrawLine(1, 35, width-2, 35, true)
			centerText(drw, 38, "ping...")
			drw.DrawFilledCircle(cx, cy, 1, true)
			for wave := r; wave > 0; wave -= spacing {
				drw.DrawCircle(cx, cy, wave, true)
			}
			display.Display()
			time.Sleep(2 * frameTime)
		}
	}
}

// loading fills a progress bar and prints the percentage under it.
func loading(drw *ssd1306.Drawer) {
	const barX, barY, barW, barH = 6, 20, width - 12, 9
	for p := 0; p <= 100; p++ {
		drw.Clear()
		border(drw)
		centerText(drw, 7, "loading")
		drw.DrawRectangle(barX, barY, barW, 1, true)
		drw.DrawRectangle(barX, barY+barH-1, barW, 1, true)
		drw.DrawRectangle(barX, barY, 1, barH, true)
		drw.DrawRectangle(barX+barW-1, barY, 1, barH, true)
		if fill := p * (barW - 4) / 100; fill > 0 {
			drw.DrawRectangle(barX+2, barY+2, fill, barH-4, true)
		}
		// Center "NN%" without building a string on the heap.
		x := (width - (numberWidth(p) + font.StringWidth("%"))) / 2
		x = drawNumber(drw, x, 34, p)
		drw.DrawText(x, 34, "%", font, true)
		display.Display()
		time.Sleep(frameTime)
	}
	time.Sleep(time.Second)
}

// border draws the 1 pixel screen frame every scene shares.
func border(drw *ssd1306.Drawer) {
	drw.DrawLine(0, 0, width-1, 0, true)
	drw.DrawLine(0, height-1, width-1, height-1, true)
	drw.DrawLine(0, 0, 0, height-1, true)
	drw.DrawLine(width-1, 0, width-1, height-1, true)
}

// centerText draws s horizontally centered at row y.
func centerText(drw *ssd1306.Drawer, y int, s string) {
	drw.DrawText((width-font.StringWidth(s))/2, y, s, font, true)
}

// drawNumber draws n in decimal at (x, y) and returns the next x. Digits are
// sliced from a constant string so nothing is heap allocated.
func drawNumber(drw *ssd1306.Drawer, x, y, n int) int {
	const digits = "0123456789"
	if n >= 10 {
		x = drawNumber(drw, x, y, n/10)
	}
	d := n % 10
	x, _ = drw.DrawText(x, y, digits[d:d+1], font, true)
	return x
}

// numberWidth returns the pixel width drawNumber uses for n.
func numberWidth(n int) int {
	const digits = "0123456789"
	w := font.StringWidth(digits[n%10 : n%10+1])
	for n >= 10 {
		n /= 10
		w += font.StringWidth(digits[n%10 : n%10+1])
	}
	return w
}

// step advances position p by velocity v, reflecting at lo and hi inclusive.
func step(p, v, lo, hi int) (int, int) {
	p += v
	if p <= lo || p >= hi {
		v = -v
	}
	return p, v
}
```

[view source](../../examples/examples-grove-single/grove-oled-display/oled-display.go)

## Walkthrough

```go
err := display.ConfigureI2C(shield.Conn(ConnPosition).I2C(400_000), ssd1306.Config{...})
```

`.I2C(400_000)` configures the two bus pins at 400 kHz. `ConfigureI2C` then
sends the SSD1306 its wake-up ritual — a couple dozen configuration commands
straight from the datasheet (charge pump on, scan direction, contrast...);
see [drivers/ssd1306](../../drivers/ssd1306/ssd1306.go) if you're curious what
a real driver's init sequence looks like.

Notice this is the first `Configure` in the guide that returns an **error**.
Of course it is: for the first time there's someone on the other end who
answers. If the display isn't plugged in, nobody acknowledges the address and
the bus reports it — unplugged is now a *detectable* condition, not a mystery.

The rest is the frame buffer at work. `Drawer` scribbles lines, circles, text
and rectangles into the byte buffer; nothing appears until `display.Display()`
sends all 384 bytes over the bus. One buffer byte holds a vertical strip of 8
pixels — that packing (and the tiny 5×7 pixel font) are the kind of squeeze
embedded graphics always plays.

## Run it

```sh
tinygo flash -target=xiao-esp32c3 ./examples/examples-grove-single/grove-oled-display/
```

You should see a looping demo: a title card, a ball bouncing inside a frame,
sonar pings expanding from center, and a loading bar counting to 100%. All on
a screen smaller than a postage stamp.

## Experiments

1. Make the loading bar real: plug the rotary knob into connector 0 and set
   the bar's fill (and percentage) from the knob fraction.
2. Write a `heartbeat` scene: a filled circle that grows and shrinks. You need
   `DrawFilledCircle`, a radius going up and down, and `display.Display()`.
3. Thermometer: temperature sensor on connector 0, draw celsius on screen with
   `drawNumber`. Your first standalone *instrument* — no computer attached.
4. In `bounce`, remove `drw.Clear()` and see what the frame buffer remembers.

## Takeaways

A bus trades pins for protocol: two wires, addresses, acknowledgements — and
real error values when a device is missing. From here on, "is it plugged in?"
is a question your code can ask. One more device on this very bus, and it's
been measuring gravity the whole time:
**[Chapter 10 — the accelerometer](10-accelerometer.md)**.

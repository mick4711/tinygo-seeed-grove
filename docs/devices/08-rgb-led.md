# 8. RGB LED stick — addressable LEDs and frame buffers

**protocol**: an agreed pattern of signals through which two devices exchange
data.

Ten LEDs, each with its own red, green and blue brightness — thirty values to
control — and the stick's Grove connector gives us *one* data wire. Time to
graduate from setting voltages to sending **data**.

<!-- TODO: photo of Grove RGB LED stick -->

## The concept

Each of the 10 LEDs is a WS2813: not a bare LED but a tiny chip *with* an LED
attached. The chips sit in a chain, data wire of one feeding the next. When
you clock a stream of color values down the wire, the first chip keeps the
first 3 bytes and forwards the rest; the second keeps the next 3; and so on
down the line. One pin, arbitrarily many LEDs — this is why these parts run
the world's LED strips and video walls.

The protocol on the wire encodes bits as pulse widths: a `1` is a long high
pulse, a `0` a short one, each lasting about a microsecond with tolerances
tight enough that the driver counts CPU cycles to hit them. You met "toggle a
pin with timing" as *sound* in the buzzer chapter; sharpened a thousandfold,
it becomes *data*.

The natural way to program such a device is a **frame buffer**: an in-memory
array holding the color of every LED. You edit the array however you like,
then push the whole thing out in one `Display()` call — exactly how your
computer's screen works, in miniature.

## Wiring

Plug the LED stick into connector **2** (`shield.Conn(2)`).

## The code

<!-- code: examples/examples-grove-single/grove-rgbled/rgbled.go -->
```go
package main

import (
	"image/color"
	"time"

	grove "github.com/soypat/seeed-grove"
)

// Define the hardware we are using.
var (
	shield grove.ShieldXiao
	rgb    grove.RGBStrip
)

const (
	numberOfLEDs = 10 // This is the Grove RGB LED Stick with 10-ws2813 minis.
	maxBright    = 64
	framePeriod  = 40 * time.Millisecond
)

func main() {
	rgb.Configure(shield.Conn(2).RGBOutput(), numberOfLEDs, maxBright)

	// Rainbow comet: a bright head bounces end to end leaving a fading
	// tail, while its color walks around the hue wheel.
	var frame [numberOfLEDs]color.RGBA
	head, dir := 0, 1
	hue := uint8(0)
	for {
		// Fade every LED towards black so previous heads become the tail.
		for i := range frame {
			frame[i].R -= frame[i].R / 4
			frame[i].G -= frame[i].G / 4
			frame[i].B -= frame[i].B / 4
		}
		frame[head] = hueToRGB(hue)

		for i := range frame {
			rgb.SetLED(i, frame[i])
		}
		rgb.Display()

		// Bounce at the strip ends and advance the hue each frame.
		head += dir
		if head == 0 || head == numberOfLEDs-1 {
			dir = -dir
		}
		hue += 3
		time.Sleep(framePeriod)
	}
}

// hueToRGB converts a hue (0-255 walks once around the color wheel) to a
// fully saturated RGB color, ramping between red, green and blue thirds.
func hueToRGB(hue uint8) color.RGBA {
	h := hue
	switch {
	case h < 85:
		return color.RGBA{R: 255 - h*3, G: h * 3, A: 255}
	case h < 170:
		h -= 85
		return color.RGBA{G: 255 - h*3, B: h * 3, A: 255}
	default:
		h -= 170
		return color.RGBA{B: 255 - h*3, R: h * 3, A: 255}
	}
}
```

[view source](../../examples/examples-grove-single/grove-rgbled/rgbled.go)

## Walkthrough

The hardware interface is two calls: `SetLED(i, color)` edits the buffer,
`Display()` streams it down the wire (via TinyGo's
[`ws2812` driver](https://pkg.go.dev/tinygo.org/x/drivers/ws2812), which
performs the cycle-counted bit-banging). Everything else in the file is
animation logic, and it shows the standard shape of every LED effect ever
written: mutate buffer, display, sleep, repeat — at 40ms per frame, 25 frames
per second.

Two tricks worth stealing:

```go
frame[i].R -= frame[i].R / 4
```

Exponential fade. Subtracting a quarter of each channel every frame makes
recently-lit LEDs dim smoothly toward black — that's the comet's tail, with no
memory of trajectory needed.

`hueToRGB` walks the color wheel with pure integer math: one byte of hue,
0–255, mapped across three ramps (red→green, green→blue, blue→red). Cheap,
smooth, and no floating point in sight.

Also note `maxBright = 64` of 255: full white on all ten LEDs is genuinely
eye-watering and thirsty (tens of milliamps per LED adds up).

## Run it

```sh
tinygo flash -target=xiao-esp32c3 ./examples/examples-grove-single/grove-rgbled/
```

No monitor needed for this one — the output is the show: a bright comet
bouncing end to end, trailing a fading tail, its color strolling around the
rainbow as it goes.

## Experiments

1. Change `framePeriod` to 10ms, then 100ms. Watch what frame rate does to
   the *feel* of motion.
2. Change the fade from `/4` to `/16` (longer tail) and `/2` (shorter). The
   tail length is a time constant — the same exponential decay you saw cooling
   the thermistor.
3. New effect from scratch: all ten LEDs the same color, hue advancing each
   frame — a slow rainbow breather. You need `SetAllLEDs`, `hueToRGB` and ten
   lines.
4. Combine with chapter 4: rotary knob on connector 0 steers the hue.

## Takeaways

Precisely-timed pulses on one pin turn into data for a chain of chips, and the
frame buffer pattern — edit memory, then push the whole frame — is how
everything from this LED stick to your monitor gets drawn. Next device also
takes frames, but through a bus with rules, addresses and acknowledgements:
**[Chapter 9 — the OLED display](09-oled-display.md)**.

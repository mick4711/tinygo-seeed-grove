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

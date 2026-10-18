package main

import (
	"image/color"
	"time"

	grove "github.com/soypat/seeed-grove"
)
const (
	ConnButton = 7
	ConnRGB = 2
	on, off = true, false
	numberOfLEDs = 10 // This is the Grove RGB LED Stick with 10-ws2813 minis.
	maxBright    = 64
	framePeriod  = 40 * time.Millisecond
)

var (
	shield grove.ShieldXiao
	button grove.PinInput
	rgb    grove.RGBStrip
	count int
	state, currState bool
)

func main() {
	button = shield.Conn(ConnButton).PinInputPulldown()
	rgb.Configure(shield.Conn(ConnRGB).RGBOutput(), numberOfLEDs, maxBright)

	for {
		// detect transition from 'off' to 'on' and increment count
		state = button.GetLevel()

		switch state {
		case currState:
			// continue
		case on:
			// transtioned to on
			currState = state
			count++
			println("count =", count)
		case off:
			// transitioned to off
			currState = state
		}
	}
}

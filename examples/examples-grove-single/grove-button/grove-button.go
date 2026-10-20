package main

import (
	"image/color"
	"time"

	grove "github.com/soypat/seeed-grove"
)
const (
	ConnButton = 0
	ConnRGB = 2
	on, off = true, false
	numberOfLEDs = 10 // This is the Grove RGB LED Stick with 10-ws2813 minis.
	maxBright    = 32
	framePeriod  = 40 * time.Millisecond
)

var (
	shield grove.ShieldXiao
	button grove.PinInput
	rgb    grove.RGBStrip
	count int
	state, currState bool
)

var (
	white = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	red = color.RGBA{R: 255, G: 0, B: 0, A: 255}
	green = color.RGBA{R: 0, G: 255, B: 0, A: 255}
	blue = color.RGBA{R: 0, G: 0, B: 255, A: 255}
	black = color.RGBA{}
)

type buttonState struct {
	currState bool
	currColour color.RGBA
}
// Return next colour in seq r, g, b, w.
func (bs *buttonState)nextColour() color.RGBA {
	switch bs.currColour {
	case white:
		bs.currColour = red
	case red:
		bs.currColour = green
	case green:
		bs.currColour = blue
	case blue:
		bs.currColour = white
	case black:
		bs.currColour = white
	default:
		bs.currColour = white
	}
	return bs.currColour
}
// Return true if transitioned to 'on', otherwise false. Update current state.
func (bs *buttonState)switchedOn(state bool) bool {
	switch state {
	case bs.currState:
		return false
	case on:
		// transtioned to on
		bs.currState = on
		return true
	case off:
		// transitioned to off
		bs.currState = off
		return false
	}
	return false
}

func main() {
	button = shield.Conn(ConnButton).PinInputPulldown()
	btnState := buttonState{
		currState: false,
		currColour: black,
	}

	rgb.Configure(shield.Conn(ConnRGB).RGBOutput(), numberOfLEDs, maxBright)
	clearPowerLight()
	var frame [numberOfLEDs]color.RGBA
	
	for {
		// detect transition from 'off' to 'on' and increment count
		state = button.GetLevel()
		if btnState.switchedOn(state) {
			count++
			println("count =", count)

			light := count % 10
			col := btnState.nextColour()
			frame[light] = col

			// switch off previous light
			if light > 0 {
				frame[light-1] = black
			} else {
				frame[9] = black
			}

			for i := range frame {
				rgb.SetLED(i, frame[i])
			}
			rgb.Display()
		}
		time.Sleep(framePeriod)
	}
}

// need to reset all leds twice to clear the green power on light
func clearPowerLight()  {
	for range 2 {
		rgb.SetAllLEDs(black)
		rgb.Display()
		time.Sleep(framePeriod)
	}
}
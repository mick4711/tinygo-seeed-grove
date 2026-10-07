package main

import (
	"time"

	grove "github.com/soypat/seeed-grove"
)

// Define the hardware we are using.
var (
	shield grove.ShieldXiao
	rotary grove.SensorRotaryAngle
)

func main() {
	rotary.Configure(shield.Conn(0).ADC(), 64)
	for {
		// Full clockwise travel is ~300 degrees across the ADC range.
		angle := rotary.AngleDegrees()
		fraction := rotary.Fraction()
		println("fraction:", fraction, "angle:", angle)
		time.Sleep(200 * time.Millisecond)
	}
}

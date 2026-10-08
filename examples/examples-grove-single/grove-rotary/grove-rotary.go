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
	buzzer := shield.Conn(1).PinOutput()
	rotary.Configure(shield.Conn(0).ADC(), 64)
	lastBucket := "none"
	for {
		// Full clockwise travel is ~300 degrees across the ADC range.
		angle := rotary.AngleDegrees()
		fraction := rotary.Fraction()
		hp := halfPeriod(fraction)
		println("fraction:", fraction, "angle:", angle, "halfPeriod:", hp)

		beep(buzzer, hp, 1000*time.Millisecond)

		currBucket := bucket(fraction)
		if currBucket != lastBucket {
			println("bucket changed to:", currBucket)
			lastBucket = currBucket
		}

		time.Sleep(200 * time.Millisecond)
	}
}

func bucket(frac float32) string {
	switch {
	case frac < 0.125:
		return "one"
	case frac < 0.25:
		return "two"
	case frac < 0.375:
		return "three"
	case frac < 0.5:
		return "four"
	case frac < 0.625:
		return "five"
	case frac < 0.75:
		return "six"
	case frac < 0.875:
		return "seven"
	default:
		return "eight"
	}
}

func halfPeriod(frac float32) time.Duration {
	hi := float32(100 * time.Microsecond)
	lo := float32(2100 * time.Microsecond)
	hp := hi + frac*(lo-hi)
	return time.Duration(hp)
}

func beep(pin grove.PinOutput, halfPeriod, dur time.Duration) {
	for elapsed := time.Duration(0); elapsed < dur; elapsed += 2 * halfPeriod {
		pin.High()
		time.Sleep(halfPeriod)
		pin.Low()
		time.Sleep(halfPeriod)
	}
}

package main

import (
	"time"

	grove "github.com/soypat/seeed-grove"
)

// Define the hardware we are using.
var shield grove.ShieldXiao

func main() {
	buzzer := shield.Conn(0).PinOutput()
	// Toggle the pin to make a square wave; half-period sets the pitch.
	const halfPeriod = 500 * time.Microsecond // ~1 kHz
	for {
		println("buzz")
		beep(buzzer, halfPeriod, 200*time.Millisecond)
		time.Sleep(time.Second)
	}
}

func beep(pin grove.PinOutput, halfPeriod, dur time.Duration) {
	for elapsed := time.Duration(0); elapsed < dur; elapsed += 2 * halfPeriod {
		pin.High()
		time.Sleep(halfPeriod)
		pin.Low()
		time.Sleep(halfPeriod)
	}
}

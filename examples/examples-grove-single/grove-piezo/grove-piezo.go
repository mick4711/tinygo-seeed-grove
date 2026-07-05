package main

import (
	"time"

	grove "github.com/soypat/seeed-grove"
)

// Define the hardware we are using.
var (
	shield grove.ShieldXiao
)

func main() {
	piezo := shield.Conn(0).ADC()
	// A vibration/knock spikes the analog output above the resting level.
	const threshold = 8000
	for {
		v := piezo.ReadAnalogValue()
		if v > threshold {
			println("vibration:", v)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

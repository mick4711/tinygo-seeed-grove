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
	n := 0
	const threshold = 1000
	for {
		v := piezo.ReadAnalogValue()
		if v > threshold {
			n++
			println("vibration:", v, n)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

package main

import (
	"time"

	grove "github.com/soypat/seeed-grove"
)

// Define the hardware we are using.
var shield grove.ShieldXiao

func main() {
	light := shield.Conn(0).ADC()
	for {
		// Brighter light lowers the photoresistor, raising the analog reading.
		v := light.ReadAnalogValue()
		pct := int(uint32(v) * 100 / 65535)
		println("raw:", v, "brightness%:", pct)
		time.Sleep(200 * time.Millisecond)
	}
}

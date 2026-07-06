package main

import (
	"time"

	grove "github.com/soypat/seeed-grove"
)

const ConnPosition = 2

// Define the hardware we are using.
var (
	shield grove.ShieldXiao
	light  grove.SensorLight
)

func main() {
	light.Configure(shield.Conn(ConnPosition).ADC(), 1)
	for {
		// Brighter light lowers the photoresistor, raising the analog reading.
		println("raw:", light.Raw(), "brightness%:", 100*light.Fraction(), "lux:", light.Lux())
		time.Sleep(200 * time.Millisecond)
	}
}

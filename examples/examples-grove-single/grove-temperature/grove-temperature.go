package main

import (
	"time"

	grove "github.com/soypat/seeed-grove"
)

// Define the hardware we are using.
var (
	shield grove.ShieldXiao
	temp   grove.SensorTemperature
)

func main() {
	const NumSamples = 16
	temp.ConfigureThermistor(shield.Conn(0).ADC(), NumSamples)
	for {
		celsius := temp.ReadTemperature()
		println("celsius:", int(celsius))
		time.Sleep(500 * time.Millisecond)
	}
}

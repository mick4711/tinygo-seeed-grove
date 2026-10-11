package main

import (
	"fmt"
	"time"

	grove "github.com/soypat/seeed-grove"
)

// Define the hardware we are using.
var (
	shield grove.ShieldXiao
	temp   grove.SensorTemperature
)

func main() {
	const NumSamples = 2048
	temp.ConfigureThermistor(shield.Conn(0).ADC(), NumSamples)
	for {
		celsius := temp.ReadTemperature()
		fmt.Printf("time: %s celsius: %.3f\n", time.Now().Format("03 hr - 04 min"), celsius)
		time.Sleep(5 * time.Minute)
	}
}

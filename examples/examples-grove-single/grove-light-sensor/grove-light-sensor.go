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
	buzzer := shield.Conn(1).PinOutput()
	light.Configure(shield.Conn(ConnPosition).ADC(), 1)
	for {
		// Brighter light lowers the photoresistor, raising the analog reading.
		raw := light.Raw()
		println("raw:", raw, "brightness%:", 100*light.Fraction(), "lux:", light.Lux())
		hp := halfPeriod(raw)
		println(hp/1000)
		beep(buzzer, hp, 200*time.Millisecond)
		time.Sleep(200 * time.Millisecond)
	}
}

func halfPeriod(adc uint16) time.Duration {
	hp := max(100.0, 2500-0.0533*float64(adc))
	return time.Duration(hp) * time.Microsecond
}

func beep(pin grove.PinOutput, halfPeriod, dur time.Duration) {
	for elapsed := time.Duration(0); elapsed < dur; elapsed += 2 * halfPeriod {
		pin.High()
		time.Sleep(halfPeriod)
		pin.Low()
		time.Sleep(halfPeriod)
	}
}

package main

import (
	"time"

	grove "github.com/soypat/seeed-grove"
	"github.com/soypat/seeed-grove/drivers/ssd1306"
)

const (
	ConnSound   = 0 // ADC pin.
	ConnRotary  = 1 // ADC pin.
	ConnTemp    = 2 // ADC pin.
	ConnDisplay = 3 // I2C
	ConnAccel   = 4 // I2C
	ConnRGB     = 5
	ConnTouch   = 6
)

// All hardware used in this project.
var (
	shield grove.ShieldXiao

	// Conn devices:
	temperature grove.SensorTemperature
	rgb         grove.RGBStrip
	display     ssd1306.Device
	piezo       grove.ADC
	touch       grove.PinInput
	rot         grove.SensorRotaryAngle
	sound       grove.SensorSound
)

func main() {
	time.Sleep(time.Second)
	err := hardwareInit()
	if err != nil {
		for {
			println("hardware initialization failed:", err.Error())
			time.Sleep(time.Second)
		}
	}
	for {

	}
}

func hardwareInit() error {
	const (
		ADCSampling                 = 4
		SoundSamplingRate           = 4000
		DisplayI2CFreq              = 400_000
		DisplayWidth, DisplayHeight = 64, 48
	)
	temperature.ConfigureThermistor(shield.Conn(ConnTemp).ADC(), ADCSampling)
	err := sound.Configure(shield.Conn(ConnSound).ADC(), SoundSamplingRate)
	if err != nil {
		return err
	}

	err = display.ConfigureI2C(shield.Conn(ConnDisplay).I2C(DisplayI2CFreq), ssd1306.Config{
		Buffer: make([]byte, ssd1306.BufferSize(DisplayWidth, DisplayHeight)),
		Width:  DisplayWidth,
		Height: DisplayHeight,
	})
	if err != nil {
		return err
	}
	touch = shield.Conn(ConnTouch).PinInputPulldown()
	rot.Configure(shield.Conn(ConnRotary).ADC(), ADCSampling)
	return nil
}

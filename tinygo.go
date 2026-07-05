//go:build tinygo

package grove

import (
	"machine"
	"sync"
)

var onceADC sync.Once

func analog(pn uint8) ADC {
	onceADC.Do(machine.InitADC)
	var sensor machine.ADC
	sensor.Pin = dPin[pn] // yellow cable is usual ADC pin.
	err := sensor.Configure(machine.ADCConfig{
		Reference:  0,
		Samples:    0,
		SampleTime: 0,
	})
	if err != nil {
		panic(err)
	}
	return sensor.Get
}

func pinout(pn uint8) PinOutput {
	p := dPin[pn]
	p.Configure(machine.PinConfig{Mode: machine.PinOutput})
	return p.Set
}

func pinInPulldown(pn uint8) PinInput {
	p := dPin[pn]
	p.Configure(machine.PinConfig{Mode: machine.PinInputPulldown})
	return p.Get
}

func pinInPullup(pn uint8) PinInput {
	p := dPin[pn]
	p.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
	return p.Get
}

func i2c(sda, scl uint8, freq uint32) I2C {
	bus := machine.I2C0
	err := bus.Configure(machine.I2CConfig{
		Frequency: freq,
		SCL:       dPin[scl],
		SDA:       dPin[sda],
	})
	if err != nil {
		panic(err)
	}
	return bus
}

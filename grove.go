package grove

import (
	"image/color"
	"math"
)

// Hardware Abstraction Layer (HAL) for devices.
type (
	ADC       func() uint16
	PinInput  func() (Level bool)
	PinOutput func(setLevel bool)
	I2C       interface {
		Tx(addr uint16, w, r []byte) error
	}
	RGBOutput interface {
		WriteColors([]color.RGBA) error
		SetBrightness(brightness uint8)
	}
)

func (adc ADC) ReadAnalogValue() uint16 { return adc() }

func (adc ADC) ReadAnalogFraction() float32 { return float32(adc()) / math.MaxUint16 }

func (po PinOutput) SetLevel(level bool) { po(level) }

func (pi PinInput) GetLevel() (level bool) { return pi() }

func (po PinOutput) High() { po(true) }
func (po PinOutput) Low()  { po(false) }

type ShieldXiao struct{}

type Conn struct {
	// white and yellow are pin numbers w.r.t xiao board.
	yellow uint8 // most extreme data/gpio pin.
	white  uint8 // next to yellow towards center. followed by red (3v3), then black (Ground).
}

func (c Conn) ConfigureWithXiaoPinNumber(pnYellow, pnWhite uint8) {
	c.yellow, c.white = pnYellow, pnWhite
}

// ConnBreakoff is a Conn wrapper for after having broken the shield off.
func (sx *ShieldXiao) ConnBreakoff(idx uint8) (c Conn) {
	return sx.Conn(idx + 2*(idx/2))
}

// Conn is indexed on the xiao shield from left to right; top to bottom.
func (sx *ShieldXiao) Conn(idx uint8) (c Conn) {
	switch idx {
	case 0, 1, 2:
		c.yellow, c.white = idx, idx+1
	case 3, 4:
		c.yellow, c.white = 5, 4 // I2C: 5=scl, 4=sda
	case 5:
		c.yellow, c.white = 7, 6 // UART: 7=rx, 6=tx
	case 6, 7:
		c.yellow, c.white = idx+2, idx+3
	default:
		c.yellow, c.white = 0xff, 0xff // no pins.
	}
	return c
}

// ADC returns the analog interface.
func (c Conn) ADC() ADC {
	return analog(c.yellow)
}

func (c Conn) PinInputPulldown() PinInput {
	return pinInPulldown(c.yellow)
}
func (c Conn) PinInputPullup() PinInput {
	return pinInPulldown(c.yellow)
}

func (c Conn) PinOutput() PinOutput {
	return pinout(c.yellow)
}

func (c Conn) I2C(freq uint32) I2C {
	sda, scl := c.white, c.yellow
	return i2c(sda, scl, freq)
}

func (c Conn) RGBOutput() RGBOutput {
	return ws2812Dev(c.yellow)
}

func sampleADC(adc ADC, nsamples uint16) (v uint16) {
	var sum uint32
	n := nsamples
	for nsamples > 0 {
		sum += uint32(adc.ReadAnalogValue())
		nsamples--
	}
	return uint16(sum / uint32(n))
}

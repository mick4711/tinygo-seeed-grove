//go:build !tinygo

package grove

import (
	"image/color"
)

var pinvals [256]bool

func pinout(pn uint8) PinOutput {
	return func(level bool) {
		pinvals[pn] = level
	}
}

func pinInPulldown(pn uint8) PinInput {
	return func() bool {
		return pinvals[pn]
	}
}

func pinInPullup(pn uint8) PinInput {
	return func() bool {
		return pinvals[pn]
	}
}

func analog(pn uint8) ADC {
	return func() uint16 {
		if pinvals[pn] {
			return 65000
		}
		return 0
	}
}
func i2c(sda, scl uint8, freq uint32) I2C {
	return mockI2C{}
}

type mockI2C struct{}

func (mockI2C) Tx(addr uint16, w, r []byte) error {
	println("i2c addr", addr, "wlen", len(w), "rlen", len(r))
	return nil
}

func ws2812Dev(pn uint8) RGBOutput {
	return mockRGBStrip{}
}

type mockRGBStrip struct {
	brighnessPcnt float32
}

func (ms mockRGBStrip) WriteColors(colors []color.RGBA) error {
	println("ws2812 write", len(colors), "colors at brightness", ms.brighnessPcnt, "%")
	return nil
}

func (ms mockRGBStrip) SetBrightness(brightness uint8) {
	ms.brighnessPcnt = 100 * float32(brightness) / 255.
}

package grove

import "image/color"

type RGBStrip struct {
	out RGBOutput
	buf []color.RGBA
}

func (rgb *RGBStrip) Configure(IOoutput RGBOutput, nled int, brightness uint8) {
	if len(rgb.buf) < nled {
		rgb.buf = make([]color.RGBA, nled)
	}
	rgb.out = IOoutput
	rgb.out.SetBrightness(brightness)
}

func (rgb *RGBStrip) SetBrightness(brightness uint8) {
	rgb.out.SetBrightness(brightness)
}

func (rgb *RGBStrip) SetAllLEDs(c color.RGBA) {
	for i := range rgb.buf {
		rgb.buf[i] = c
	}
}

func (rgb *RGBStrip) SetLED(i int, c color.RGBA) {
	rgb.buf[i] = c
}

func (rgb *RGBStrip) Display() error {
	return rgb.out.WriteColors(rgb.buf)
}

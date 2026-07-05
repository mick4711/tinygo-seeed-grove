package grove

import "math"

type SensorRotaryAngle struct {
	adc     ADC
	samples uint16
}

func (rot *SensorRotaryAngle) Configure(adc ADC, nsamples uint16) {
	if nsamples == 0 {
		panic("zero sampling")
	}
	rot.adc = adc
	rot.samples = nsamples
}

func (rot *SensorRotaryAngle) sample() uint16 {
	return sampleADC(rot.adc, rot.samples)
}

func (rot *SensorRotaryAngle) Fraction() float32 {
	return float32(rot.sample()) / math.MaxUint16
}

func (rot *SensorRotaryAngle) Raw() uint16 {
	return rot.adc.ReadAnalogValue()
}

func (rot *SensorRotaryAngle) AngleDegrees() float32 {
	return rot.Fraction() * 300
}

func (rot *SensorRotaryAngle) AngleRadians() float32 {
	return rot.Fraction() * (2 * math.Pi * 300. / 360.)
}

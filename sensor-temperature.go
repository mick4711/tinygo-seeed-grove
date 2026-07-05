package grove

import (
	"github.com/chewxy/math32"
)

type SensorTemperature struct {
	adc      ADC
	b, r0    float32
	nsamples uint16
}

func (temp *SensorTemperature) ConfigureThermistor(adc ADC, nSamples uint16) {
	// Grove Temperature v1.2: NTC thermistor, B=4275, 100k at 25C.
	const B, R0 = 4275.0, 100000.0
	temp.b = B
	temp.r0 = R0
	temp.nsamples = nSamples
	temp.adc = adc
}

func (temp *SensorTemperature) sample() uint16 {
	return sampleADC(temp.adc, temp.nsamples)
}

func (temp *SensorTemperature) ReadTemperature() float32 {
	v := float32(temp.sample())
	r := temp.r0 * (65535.0/v - 1.0) // thermistor resistance from divider
	celsius := 1.0/(math32.Log(r/temp.r0)/temp.b+1.0/298.15) - 273.15
	return celsius
}

func (temp *SensorTemperature) ReadRaw() uint16 {
	return temp.adc()
}

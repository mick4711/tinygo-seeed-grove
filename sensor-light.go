package grove

type SensorLight struct {
	adc      ADC
	rawMax   uint16
	nsamples uint16
}

func (light *SensorLight) Configure(adc ADC, nSamples uint16) {
	if nSamples == 0 {
		panic("zero sampling")
	}
	// Grove Light Sensor v1.2: LS06-S phototransistor into divider+buffer.
	// The output saturates around 70% of the 3.3V rail, well short of full
	// ADC scale, so full-scale is the measured ceiling rather than MaxUint16.
	const rawSaturation = 46000
	light.adc = adc
	light.rawMax = rawSaturation
	light.nsamples = nSamples
}

func (light *SensorLight) sample() uint16 {
	return sampleADC(light.adc, light.nsamples)
}

// Fraction returns brightness in range 0..1 relative to sensor saturation.
func (light *SensorLight) Fraction() float32 {
	v := light.sample()
	if v > light.rawMax {
		v = light.rawMax
	}
	return float32(v) / float32(light.rawMax)
}

// Lux estimates illuminance from the sensor's nominal 0..350 lux range.
// The phototransistor is uncalibrated and nonlinear; treat as approximate.
func (light *SensorLight) Lux() float32 {
	return light.Fraction() * 350
}

func (light *SensorLight) Raw() uint16 {
	return light.adc.ReadAnalogValue()
}

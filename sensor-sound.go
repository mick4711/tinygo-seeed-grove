package grove

import "time"

type SensorSound struct {
	adc ADC

	// realtime sampling API:
	period time.Duration
	n      time.Duration
	start  time.Time
}

func (ss *SensorSound) Configure(adc ADC, samplingRate int) error {
	*ss = SensorSound{
		adc:    adc,
		period: time.Second / time.Duration(samplingRate),
	}
	ss.Sync()
	return nil
}

// Sample integrates ADC reads back-to-back until the sample deadline and
// returns their mean. Filling the whole period (instead of a fixed burst then
// idling) does two jobs: the boxcar average is an anti-alias filter with a
// null at the sample rate, and averaging N uncorrelated reads cuts ADC/analog
// noise by √N. The idle busy-wait bought nothing; reads are the better spin.
func (ss *SensorSound) Sample() uint16 {
	target := ss.commitNextDeadline()
	var sum, n uint32
	for {
		sum += uint32(ss.adc.ReadAnalogValue())
		n++
		if time.Since(target) >= 0 {
			break
		}
	}
	return uint16(sum / n)
}

func (ss *SensorSound) SampleNoWait(nSamples uint16) uint16 {
	return sampleADC(ss.adc, uint16(nSamples))
}

func (s *SensorSound) Sync() {
	s.start = time.Now()
	s.n = 0
}

func (s *SensorSound) commitNextDeadline() time.Time {
	deadline := s.start.Add(s.period * s.n)
	s.n++
	return deadline
}

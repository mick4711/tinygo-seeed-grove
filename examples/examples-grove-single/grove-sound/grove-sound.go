package main

import (
	"math"
	"time"

	grove "github.com/soypat/seeed-grove"
)

// Define the hardware we are using.
var (
	shield grove.ShieldXiao
	sound  grove.SensorSound
)

func main() {
	const SampleRate = 8000 // Not actually used in this example, we sample without waiting.
	sound.Configure(shield.Conn(0).ADC(), SampleRate)
	const n = 1000
	for {
		// Sample fast across a window to capture the waveform swing.
		var sum, sumSq uint64
		for i := 0; i < n; i++ {
			v := uint64(sound.SampleNoWait(1))
			sum += v
			sumSq += v * v
		}
		// Loudness = RMS deviation from the DC bias (variance = E[x^2] - E[x]^2).
		mean := sum / n
		exSq, mSq := sumSq/n, mean*mean
		var variance uint64
		if exSq > mSq {
			variance = exSq - mSq
		}
		loudness := uint64(math.Sqrt(float64(variance)))
		println("dc:", mean, "loudness:", loudness)
		time.Sleep(200 * time.Millisecond)
	}
}

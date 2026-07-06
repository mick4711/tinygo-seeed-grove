# 5. Sound sensor — sampling fast

**sampling**: measuring a signal at regular instants to reconstruct how it
changes over time.

So far our analog readings have been leisurely — brightness and knob positions
change on human timescales, so five readings per second was plenty. Sound is
different. A voice wiggles the air hundreds of times per second. To *see* a
signal like that, we don't read the ADC occasionally; we hammer it.

<!-- TODO: photo of Grove sound sensor module -->

## The concept

The module is a small electret microphone followed by an amplifier. The
microphone's diaphragm vibrates with the air; the amplifier scales that
vibration into a voltage swing our ADC can see.

Two ideas make microphone data different from everything we've read so far:

**DC bias.** Sound waves swing negative and positive, but our ADC only
measures 0 to 3.3V. So the module centers the signal on a middle voltage —
the *bias* — and the audio rides on top as a wiggle. Silence isn't a reading
of 0; it's a steady reading at the bias with no wiggle.

**Loudness is deviation.** Since the interesting part is the wiggle, loudness
is how far readings *spread* around their mean — not the mean itself. The
standard measure is the RMS (root-mean-square) deviation, which statisticians
call the standard deviation: `sqrt(E[x²] − E[x]²)`.

## Wiring

Plug the sound sensor into connector **0** (`shield.Conn(0)`).

## The code

<!-- code: examples/examples-grove-single/grove-sound/grove-sound.go -->
```go
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
```

[view source](../../examples/examples-grove-single/grove-sound/grove-sound.go)

## Walkthrough

The inner loop takes 1000 samples back-to-back, as fast as the ADC will
convert — no sleeping, because sleeping between samples would miss the
waveform's motion. For each window it accumulates the sum and the sum of
squares, then applies the variance identity:

```go
mean := sum / n              // the DC bias
variance := sumSq/n - mean*mean
loudness := sqrt(variance)   // RMS deviation from the bias
```

Watch the two printed numbers do independent jobs: `dc` sits nearly constant
(that's the bias — the amplifier's resting point), while `loudness` leaps when
you speak. All integer math until the final square root, because on a small
chip, floating point in a hot loop is a luxury to spend deliberately.

One honest limitation: a software loop reading the ADC tops out around a few
thousand samples per second. Enough for loudness and voice, far short of music
quality (CDs sample at 44100 Hz). The [sound recorder project](../projects/sound-recorder.md)
pushes this exact setup to its limit — and actually records audio you can play.

## Run it

```sh
tinygo flash -target=xiao-esp32c3 ./examples/examples-grove-single/grove-sound/
tinygo monitor
```

In a quiet room, `loudness` should idle low while `dc` holds steady. Talk,
clap, or hum at the microphone: `loudness` jumps with you.

## Experiments

1. Build a clap detector: when `loudness` crosses a threshold, print `CLAP!`.
   Then try to detect a *double* clap — two crossings within half a second.
2. Print only `loudness` as a bar instead of a number: write a loop that
   prints one `#` per 100 counts. A poor man's VU meter in the monitor.
3. Shrink the window from 1000 samples to 100. The updates come faster but
   jumpier — you've traded averaging for responsiveness. Where's the sweet
   spot for your voice?

## Takeaways

Fast-changing signals demand back-to-back sampling; a biased signal means the
information lives in the deviation, not the average; and mean/variance are
computable in one integer pass. Next chapter the voltage divider returns, but
this time the thing twisting the knob is *temperature itself*:
**[Chapter 6 — the temperature sensor](06-temperature.md)**.

# 6. Temperature sensor — thermistors

**thermistor**: a resistor whose resistance changes strongly with temperature.

No chip on this module reports "23°C" — nothing on it speaks digital at all.
There is a component whose *resistance* depends on temperature, and a bit of
math. This chapter is about how a raw physical property becomes a calibrated
measurement, which is secretly how most sensors in the world work.

<!-- TODO: photo of Grove temperature sensor module -->

## The concept

Our module carries an NTC (negative temperature coefficient) thermistor:
heat it up and its resistance drops, exponentially. Two datasheet numbers
characterize it completely — at 25°C it measures 100kΩ (called `R0`), and the
steepness of the exponential is `B = 4275` kelvin.

But the ADC measures voltage, not resistance. The trick is the **voltage
divider** from the rotary chapter: the module puts the thermistor in series
with a fixed resistor across 3.3V. As temperature moves the thermistor's
resistance, the voltage at the midpoint moves too — and *that* we can sample.

So the measurement chain is:

temperature → resistance → voltage → ADC counts → (math) → °C

## Wiring

Plug the temperature sensor into connector **0** (`shield.Conn(0)`).

## The code

<!-- code: examples/examples-grove-single/grove-temperature/grove-temperature.go -->
```go
package main

import (
	"time"

	grove "github.com/soypat/seeed-grove"
)

// Define the hardware we are using.
var (
	shield grove.ShieldXiao
	temp   grove.SensorTemperature
)

func main() {
	const NumSamples = 16
	temp.ConfigureThermistor(shield.Conn(0).ADC(), NumSamples)
	for {
		celsius := temp.ReadTemperature()
		println("celsius:", int(celsius))
		time.Sleep(500 * time.Millisecond)
	}
}
```

[view source](../../examples/examples-grove-single/grove-temperature/grove-temperature.go)

## Walkthrough

The program is four lines; the knowledge hides in
`SensorTemperature.ReadTemperature`, and it's worth peeling open
([sensor-temperature.go](../../sensor-temperature.go)):

```go
r := temp.r0 * (65535.0/v - 1.0) // undo the voltage divider
celsius := 1.0/(math32.Log(r/temp.r0)/temp.b + 1.0/298.15) - 273.15
```

Line one runs the divider equation backwards: from the ADC fraction, recover
the thermistor's resistance. Line two is the **B-parameter equation** — the
inverted exponential law of the thermistor, anchored at the calibration point
(298.15 K is 25°C in kelvin, where we know the resistance is exactly `R0`).
Straight off the datasheet, and now you know why those two constants, 4275 and
100000, appear in the source.

Also note `ConfigureThermistor(..., NumSamples)` with 16 samples averaged per
reading. Temperature changes slowly, so we can afford heavy averaging for a
rock-steady output — the opposite trade from the sound sensor, which couldn't
afford to average away its own signal.

## Run it

```sh
tinygo flash -target=xiao-esp32c3 ./examples/examples-grove-single/grove-temperature/
tinygo monitor
```

You should see the ambient temperature in celsius, twice a second. Pinch the
black blob (the thermistor) between two fingers: watch it climb toward body
temperature. Let go: an exponential glide back down.

## Experiments

1. Print tenths of a degree instead of whole degrees (print
   `int(celsius*10)`), and *now* pinch the sensor. How many seconds to rise
   one full degree? You're watching thermal mass.
2. Build a fridge alarm: buzzer on connector 1, beep when temperature leaves
   a chosen band.
3. Log a cooldown curve: heat the sensor with your fingers, release, and print
   temperature every 500ms. Paste the numbers into a plot. Exponential decay,
   in the flesh.

## Takeaways

Most sensing is indirect — the thermistor gave us resistance, the divider gave
us voltage, and datasheet constants carried the number the rest of the way to
celsius. And averaging is a dial you set per-sensor: crank it for slow signals,
back off for fast ones. Next, a sensor that only speaks in spikes:
**[Chapter 7 — the piezo vibration sensor](07-piezo.md)**.

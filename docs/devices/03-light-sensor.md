# 3. Light sensor — analog signals and the ADC

**analog signal**: a voltage that varies continuously, carrying information in
*how much* rather than *whether*.

Digital pins answer yes/no questions. But the world mostly doesn't speak in
yes/no — it speaks in *amounts*: how bright, how warm, how loud. This chapter
introduces the peripheral that translates amounts into numbers, and it might be
the most important peripheral in the whole kit: the **Analog to Digital
Converter**, or ADC.

<!-- TODO: photo of Grove light sensor module -->

## The concept

The module's light-facing component is an LS06-S phototransistor: the brighter
the light hitting it, the more current it lets through. The module turns that
current into a voltage on the yellow wire — more light, higher voltage. A
smooth, continuous voltage anywhere between 0 and 3.3V.

A CPU cannot store "2.13 volts" in a variable. The ADC is the measuring
instrument that fixes this: given a voltage, it returns an integer
proportional to it. The ESP32-C3's ADC measures with 12-bit precision — it
divides the voltage range into 4096 steps. The `grove` package scales every
reading to a common 16-bit range (0–65535) so that all boards, whatever their
ADC precision, speak the same units.

## Wiring

Plug the light sensor into connector **2** (`shield.Conn(2)`). Remember only
connectors 0, 1 and 2 reach the chip's ADC — the other connectors can't do
analog.

## The code

<!-- code: examples/examples-grove-single/grove-light-sensor/grove-light-sensor.go -->
```go
package main

import (
	"time"

	grove "github.com/soypat/seeed-grove"
)

const ConnPosition = 2

// Define the hardware we are using.
var (
	shield grove.ShieldXiao
	light  grove.SensorLight
)

func main() {
	light.Configure(shield.Conn(ConnPosition).ADC(), 1)
	for {
		// Brighter light lowers the photoresistor, raising the analog reading.
		println("raw:", light.Raw(), "brightness%:", 100*light.Fraction(), "lux:", light.Lux())
		time.Sleep(200 * time.Millisecond)
	}
}
```

[view source](../../examples/examples-grove-single/grove-light-sensor/grove-light-sensor.go)

## Walkthrough

```go
light.Configure(shield.Conn(ConnPosition).ADC(), 1)
```

`.ADC()` configures the connector's yellow pin as an analog input and returns
a `grove.ADC` — a function that performs one conversion and returns the
0–65535 result. The `1` is how many samples to average per reading (more on
why averaging matters in the sound chapter).

`grove.SensorLight` is a thin layer of *sensor knowledge* on top of that raw
number. Two facts about this particular module live inside it: the output
saturates around 70% of the 3.3V rail (a raw reading of about 46000), so
`Fraction()` reports brightness relative to that real ceiling rather than the
theoretical one; and the datasheet's nominal range tops out around 350 lux,
which `Lux()` uses for a rough real-world estimate. The phototransistor is
uncalibrated and nonlinear — treat lux as a vibe, fraction as the truth.

This is a pattern you'll see all through the kit: the ADC gives honest raw
counts, and a small sensor type turns counts into meaning.

## Run it

```sh
tinygo flash -target=xiao-esp32c3 ./examples/examples-grove-single/grove-light-sensor/
tinygo monitor
```

You should see raw counts, a brightness percentage and a lux estimate stream
by five times a second. Cover the sensor with your hand — the numbers dive.
Point your phone flashlight at it — they slam into the ceiling.

## Experiments

1. Find the darkness threshold of your room: watch `Raw()` with lights on and
   off, pick a value in between, and print `"dark!"` when the reading drops
   below it. You've built a twilight switch.
2. Plug the buzzer into connector 1 and make pitch follow brightness — a light
   theremin. Map `Fraction()` to a `halfPeriod` between 200µs and 2000µs.
3. How fast does the sensor respond? Wave your hand rapidly over it and lower
   the sleep to 20ms. Can you see the shadow of each pass?

## Takeaways

The ADC is the bridge from continuous physics to discrete numbers — 65536
shades of gray where digital pins gave us two. And raw counts only become
measurements when combined with knowledge about the specific sensor. Next, a
sensor where the voltage means *position*:
**[Chapter 4 — the rotary angle sensor](04-rotary.md)**.

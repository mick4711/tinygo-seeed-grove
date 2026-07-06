# 4. Rotary angle sensor — potentiometers

**potentiometer**: a resistor with a movable contact, dividing a voltage in
proportion to the contact's position.

The knob. Every volume dial, joystick axis and analog game controller you have
ever used was, at heart, this chapter. The rotary angle sensor gives our
programs their first *control input* that isn't just on/off.

<!-- TODO: photo of Grove rotary angle sensor module -->

## The concept

Inside the module is a circular resistive track with a wiper — a sliding
contact attached to the shaft. One end of the track sits at 3.3V, the other at
0V. The wiper picks off the voltage at its position: turned fully one way it
touches the 0V end, fully the other way the 3.3V end, and anywhere in between
it reads a proportional fraction. This is a **voltage divider**, one of the two
or three most useful ideas in all of electronics — we will meet it again in the
temperature chapter.

The mechanical rotation range of this pot is about **300 degrees**, so
converting a reading to an angle is one multiplication.

## Wiring

Plug the rotary sensor into connector **0** (`shield.Conn(0)`) — analog again,
so connectors 0–2 only.

## The code

<!-- code: examples/examples-grove-single/grove-rotary/grove-rotary.go -->
```go
package main

import (
	"time"

	grove "github.com/soypat/seeed-grove"
)

// Define the hardware we are using.
var (
	shield grove.ShieldXiao
	rotary grove.SensorRotaryAngle
)

func main() {
	rotary.Configure(shield.Conn(0).ADC(), 1)
	for {
		// Full clockwise travel is ~300 degrees across the ADC range.
		angle := rotary.AngleDegrees()
		fraction := rotary.Fraction()
		println("fraction:", fraction, "angle:", angle)
		time.Sleep(200 * time.Millisecond)
	}
}
```

[view source](../../examples/examples-grove-single/grove-rotary/grove-rotary.go)

## Walkthrough

By now the shape is familiar: `.ADC()` for the raw pipe,
`grove.SensorRotaryAngle` for the sensor knowledge. The knowledge here is
minimal and honest:

```go
func (rot *SensorRotaryAngle) AngleDegrees() float32 {
	return rot.Fraction() * 300
}
```

`Fraction()` is the wiper position as 0..1 of the ADC's full 16-bit range, and
300 is the pot's mechanical travel. That's the entire calibration.

One thing worth noticing in the monitor output: park the knob and the readings
still flutter by a few counts. That's noise — the ADC's least significant bits
are always a little lively. It's why `Configure` takes a sample-averaging
count, and why real products add "dead zones" around joystick centers.

## Run it

```sh
tinygo flash -target=xiao-esp32c3 ./examples/examples-grove-single/grove-rotary/
tinygo monitor
```

Turn the knob slowly end to end. You should see `fraction` sweep 0 to 1 and
`angle` sweep 0 to ~300 degrees, tracking your hand in real time.

## Experiments

1. Raise the averaging: change `Configure(..., 1)` to `Configure(..., 64)`.
   Park the knob — how much calmer is the reading?
2. Plug the buzzer into connector 1 and make the knob a pitch dial: map the
   fraction to a half-period. Congratulations, you built a (terrible) synth.
3. Turn the knob into an 8-position selector switch: divide the fraction into
   8 buckets and print the bucket number only when it changes. Does it ever
   flicker between two buckets at a boundary? How would you fix that?
   (This flickering problem is called hysteresis — worth a search.)

## Takeaways

A potentiometer is a voltage divider you can grab, and mapping ADC counts to
physical units is often a single multiply — *if* you know the sensor's real
range. Next we push the ADC to its limits, sampling not five times a second
but thousands: **[Chapter 5 — the sound sensor](05-sound.md)**.

# 1. Buzzer — digital output

**digital output**: a pin whose voltage software can set to one of exactly two
levels: high (3.3V) or low (0V).

The buzzer is the simplest module in the kit and the perfect first step: we are
going to make sound out of nothing but a pin turning on and off. If you
understand this chapter, you understand the "O" in GPIO — the mechanism behind
every LED, relay and motor driver you will ever control.

<!-- TODO: photo of Grove buzzer module -->

## The concept

Inside the module sits a piezo disc: a ceramic wafer that physically flexes
when voltage is applied across it. Put 3.3V on it — it bends. Remove the
voltage — it springs back. Do that once and nothing audible happens. Do it a
thousand times per second and the disc becomes a tiny drum head pumping the
air at 1000 Hz, which your ear reports as a slightly annoyed beep.

So the recipe for sound is: toggle a digital output at an audible frequency.
Pitch is just how fast we toggle. There is no sound chip here — *the software
is the sound chip*.

## Wiring

Plug the buzzer into connector **1** on the shield (the example uses
`shield.Conn(1)`). That's it. Yellow wire carries our signal, red and black
carry power.

## The code

<!-- code: examples/examples-grove-single/grove-buzzer/grove-buzzer.go -->
```go
package main

import (
	"time"

	grove "github.com/soypat/seeed-grove"
)

const ConnPosition = 1

// Define the hardware we are using.
var shield grove.ShieldXiao

func main() {
	buzzer := shield.Conn(ConnPosition).PinOutput()
	// Toggle the pin to make a square wave; half-period sets the pitch.
	const halfPeriod = 500 * time.Microsecond // ~1 kHz
	for {
		println("buzz")
		beep(buzzer, halfPeriod, 200*time.Millisecond)
		time.Sleep(time.Second)
	}
}

func beep(pin grove.PinOutput, halfPeriod, dur time.Duration) {
	for elapsed := time.Duration(0); elapsed < dur; elapsed += 2 * halfPeriod {
		pin.High()
		time.Sleep(halfPeriod)
		pin.Low()
		time.Sleep(halfPeriod)
	}
}
```

[view source](../../examples/examples-grove-single/grove-buzzer/grove-buzzer.go)

## Walkthrough

The interesting line is this one:

```go
buzzer := shield.Conn(ConnPosition).PinOutput()
```

`shield.Conn(1)` looks up which microcontroller pin is wired to the yellow wire
of connector 1, and `PinOutput()` configures that pin as an output. What comes
back is a `grove.PinOutput` — in truth just a function that sets a voltage.
When you call `pin.High()`, several layers down a value is written into a
memory-mapped register inside the ESP32-C3, and the transistor driving that
physical pin connects it to 3.3V. `pin.Low()` connects it to ground.
That register write is the exact boundary where software ends and physics
begins.

The `beep` function is our square-wave oscillator: high for `halfPeriod`, low
for `halfPeriod`, repeated for the duration. One full cycle takes
`2 × 500µs = 1ms`, so the frequency is 1 kHz. Note what sets the pitch —
not voltage, not some amplitude knob, purely *time*.

## Run it

```sh
tinygo flash -target=xiao-esp32c3 ./examples/examples-grove-single/grove-buzzer/
tinygo monitor
```

You should hear a short beep every second, and see `buzz` printed in the
monitor each time. Music to our ears — that pin is toggling a thousand times
a second on your command.

## Experiments

1. Change `halfPeriod` to `250 * time.Microsecond`. Predict the pitch before
   you flash. Were you right? (Hint: half the period, twice the frequency.)
2. Play a little melody: call `beep` several times with different half-periods.
   A4 is 440 Hz — what `halfPeriod` does that need?
3. Make a siren: sweep `halfPeriod` gradually from 1000µs down to 300µs and
   back in a loop.

## Takeaways

A digital output is a software-controlled switch between 3.3V and 0V, and
timing is everything: the same pin is a beep at 1 kHz and silence at 1 Hz.
Next we flip the direction of that pin:
**[Chapter 2 — the touch sensor](02-touch.md)**.

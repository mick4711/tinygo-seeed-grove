# 2. Touch sensor — digital input

**digital input**: a pin whose voltage the software can read as one of two
levels: high (true) or low (false).

Last chapter we pushed voltage *out* of a pin. Now we let the outside world
push voltage *in*. The touch sensor is our first way of asking the physical
world a question — even if the question is only "is a finger here, yes or no?".

<!-- TODO: photo of Grove touch sensor module -->

## The concept

The module carries a TTP223 capacitive touch chip, the same technology as your
phone screen. Your body is (electrically speaking) a decently sized capacitor;
when a finger approaches the pad, the pad's capacitance changes and the chip
notices. Its answer comes out on the yellow wire: 3.3V while touched, 0V
otherwise.

One subtlety hides in the code: what does an input pin read when *nothing* is
driving it? Not 0 — nothing. The pin floats, picking up whatever
electromagnetic gossip is nearby, reading true or false at random. The fix is
a **pull resistor**: a weak internal resistor that ties the pin to a known
level when nobody is driving it. We configure a pull-*down*, so the pin rests
at a firm 0 and only reads true when the sensor actively drives it high.

## Wiring

Plug the touch sensor into connector **0** (`shield.Conn(0)`).

## The code

<!-- code: examples/examples-grove-single/grove-touch/grove-touch.go -->
```go
package main

import (
	"time"

	grove "github.com/soypat/seeed-grove"
)

// Define the hardware we are using.
var shield grove.ShieldXiao

func main() {
	// Capacitive touch drives the data line high while a finger is present.
	touch := shield.Conn(0).PinInputPulldown()
	lastPrint := time.Now()
	last := false
	for {
		beingTouched := touch.GetLevel()
		if beingTouched && time.Since(lastPrint) > time.Second {
			println("i am being touched!")
			lastPrint = time.Now()
		} else if last && !beingTouched {
			println("oh no, I am getting cold without your touch :(")
		}
		last = beingTouched
		time.Sleep(20 * time.Millisecond)
	}
}
```

[view source](../../examples/examples-grove-single/grove-touch/grove-touch.go)

## Walkthrough

```go
touch := shield.Conn(0).PinInputPulldown()
```

Same shape as the buzzer's `PinOutput()`, opposite direction: this configures
the pin as an input with the internal pull-down enabled, and hands back a
function that reads the pin. Underneath, `touch.GetLevel()` reads a
memory-mapped input register and reports whether the pin currently sits above
the chip's logic threshold.

The loop polls the pin every 20 milliseconds. Polling — checking repeatedly
instead of being notified — is the bread and butter of small embedded
programs. 50 checks per second is glacial for a 160 MHz CPU and instantaneous
for a human finger.

Notice the `last` variable: by remembering the previous reading, the program
detects the *release* (was touched, now isn't) — an **edge**, not a level.
Levels tell you the state of the world; edges tell you something *happened*.

## Run it

```sh
tinygo flash -target=xiao-esp32c3 ./examples/examples-grove-single/grove-touch/
tinygo monitor
```

Rest your finger on the pad: `i am being touched!` about once a second. Lift
it: a heartbroken farewell. You are now a source of input events.

## Experiments

1. Count touches instead of printing on release, and print the running total.
2. The sensor works through thin materials — put a sheet of paper over the pad.
   Does it still trigger? Two sheets?
3. Combine chapters: also plug in the buzzer (connector 1) and make it beep
   while the pad is touched. You already know both halves.

## Takeaways

Inputs need a defined resting state — that's what pull resistors are for — and
the difference between a level ("is touched") and an edge ("was just released")
is the difference between state and event. So far our world is black and white,
3.3V or 0V. Time for shades of gray:
**[Chapter 3 — the light sensor](03-light-sensor.md)**.

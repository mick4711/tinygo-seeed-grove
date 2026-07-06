# 7. Piezo vibration sensor — event detection

**threshold**: a value that splits a continuous signal into two states —
below: nothing, above: something happened.

Remember the buzzer? We put voltage across a piezo disc and it flexed. Piezo
materials run both directions: flex the disc and it *generates* voltage. Same
physics, reversed. The buzzer chapter's loudspeaker is this chapter's
microphone for knocks.

<!-- TODO: photo of Grove piezo vibration sensor module -->

## The concept

Tap the module and the piezo disc deforms for a few milliseconds, producing a
sharp voltage spike that decays immediately. Between taps: a quiet resting
level. The signal is almost always boring, occasionally spectacular.

That shape calls for a different reading strategy. There is no meaningful
"vibration level" to average and report — the information is entirely in the
brief spikes, so we poll fast and compare against a **threshold**: readings
below it are silence, readings above it are an *event*. This turns an analog
sensor into an event source, the same level-versus-event idea as the touch
chapter, except this time *we* draw the line between the two states.

## Wiring

Plug the piezo sensor into connector **0** (`shield.Conn(0)`).

## The code

<!-- code: examples/examples-grove-single/grove-piezo/grove-piezo.go -->
```go
package main

import (
	"time"

	grove "github.com/soypat/seeed-grove"
)

// Define the hardware we are using.
var (
	shield grove.ShieldXiao
)

func main() {
	piezo := shield.Conn(0).ADC()
	// A vibration/knock spikes the analog output above the resting level.
	n := 0
	const threshold = 8000
	for {
		v := piezo.ReadAnalogValue()
		if v > threshold {
			n++
			println("vibration:", v, n)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
```

[view source](../../examples/examples-grove-single/grove-piezo/grove-piezo.go)

## Walkthrough

No sensor wrapper type this time — the raw `grove.ADC` is the whole interface,
because the module is just the disc and a resistor. The loop reads every 10
milliseconds and applies the rule:

```go
if v > threshold {
```

Everything interesting lives in two numbers. The `threshold` (8000 of 65535)
sits above the resting noise but below a real tap — set it too low and
electrical noise "knocks", too high and only hammer blows register. And the
10ms poll period is a bet that a spike stays above threshold long enough for
us to catch it; poll too slowly and short taps slip between readings,
unobserved.

## Run it

```sh
tinygo flash -target=xiao-esp32c3 ./examples/examples-grove-single/grove-piezo/
tinygo monitor
```

Silence in the monitor until you tap the sensor (or the table it rests on) —
then a `vibration:` line with the spike's height and a running count. Note how
the value scales with your enthusiasm.

## Experiments

1. Find your noise floor: drop the threshold to 1000, then 500. At what value
   do phantom knocks appear on their own?
2. One physical knock sometimes prints several lines — the disc rings above
   the threshold across multiple polls. Fix it by ignoring readings for 200ms
   after a detection. (This is called a debounce, and you'll use it forever.)
3. Secret knock lock: record the gaps between knocks with `time.Since`, and
   flash an RGB LED (chapter 8 sneak peek) or beep the buzzer only when the
   rhythm matches shave-and-a-haircut.

## Takeaways

Spiky signals want thresholds, not averages, and a threshold plus a debounce
is the standard recipe for turning raw analog into clean events. That closes
our tour of the analog world — next we command ten LEDs through one wire by
speaking their protocol: **[Chapter 8 — the RGB LED stick](08-rgb-led.md)**.

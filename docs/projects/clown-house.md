# Project: the clown house

**integration**: making separately working parts work *together*.

Every chapter so far plugged in one module, used it, unplugged it. The clown
house plugs in **all eight connectors at once** — sound, rotary, temperature,
display, accelerometer, RGB stick, touch and buzzer — and faces the questions
single-device programs never meet: who goes in which socket, and what do you
do when one of eight devices fails to start?

<!-- TODO: photo of shield fully populated with all modules -->

## The wiring plan

With every socket taken, connector assignment stops being arbitrary. The
constraints we know from the chapters: analog devices *must* land on
connectors 0–2, I2C devices on 3–4, and everything digital can take what's
left:

| Connector | Device | Why here |
|-----------|--------|----------|
| 0 | Sound sensor | needs ADC |
| 1 | Rotary angle | needs ADC |
| 2 | Temperature | needs ADC |
| 3 | OLED display | I2C |
| 4 | Accelerometer | I2C (same bus, different address) |
| 5 | RGB LED stick | digital out |
| 6 | Touch sensor | digital in |
| 7 | Buzzer | digital out |

## The code

<!-- code: examples/grove-clown-house/clown-house.go -->
```go
package main

import (
	"time"

	grove "github.com/soypat/seeed-grove"
	"github.com/soypat/seeed-grove/drivers/ssd1306"
)

const (
	ConnSound   = 0 // ADC pin.
	ConnRotary  = 1 // ADC pin.
	ConnTemp    = 2 // ADC pin.
	ConnDisplay = 3 // I2C
	ConnAccel   = 4 // I2C
	ConnRGB     = 5
	ConnTouch   = 6
	ConnBuzzer  = 7
)

// All hardware used in this project.
var (
	shield grove.ShieldXiao

	// Conn devices:
	temperature grove.SensorTemperature
	rgb         grove.RGBStrip
	display     ssd1306.Device
	piezo       grove.ADC
	touch       grove.PinInput
	rot         grove.SensorRotaryAngle
	sound       grove.SensorSound
	buzzer      grove.ADC
)

func main() {
	time.Sleep(time.Second)
	err := hardwareInit()
	if err != nil {
		for {
			println("hardware initialization failed:", err.Error())
			time.Sleep(time.Second)
		}
	}
	for {

	}
}

func hardwareInit() error {
	const (
		ADCSampling                 = 4
		SoundSamplingRate           = 4000
		DisplayI2CFreq              = 400_000
		DisplayWidth, DisplayHeight = 64, 48
	)
	temperature.ConfigureThermistor(shield.Conn(ConnTemp).ADC(), ADCSampling)
	err := sound.Configure(shield.Conn(ConnSound).ADC(), SoundSamplingRate)
	if err != nil {
		return err
	}

	err = display.ConfigureI2C(shield.Conn(ConnDisplay).I2C(DisplayI2CFreq), ssd1306.Config{
		Buffer: make([]byte, ssd1306.BufferSize(DisplayWidth, DisplayHeight)),
		Width:  DisplayWidth,
		Height: DisplayHeight,
	})
	if err != nil {
		return err
	}
	touch = shield.Conn(ConnTouch).PinInputPulldown()
	rot.Configure(shield.Conn(ConnRotary).ADC(), ADCSampling)
	return nil
}
```

[view source](../../examples/grove-clown-house/clown-house.go)

## Walkthrough

Two things distinguish this program from a chapter example glued eight times.

**All hardware is declared in one place**, as package-level variables with the
connector plan in named constants beside it. When something misbehaves at
3am, `ConnAccel = 4` in one block beats connector numbers scattered through
five hundred lines.

**Initialization is separated and fallible.** `hardwareInit` configures every
device and returns the first error; `main` traps failure in a loop that
prints the reason forever. On a microcontroller there is no exit code, no
stderr, no crash dialog — if you don't *plan* where errors go, they go
nowhere. A device left unplugged shows up in `tinygo monitor` by name instead
of as silent weirdness later.

Then `main` settles into... an empty loop. Deliberate: the house is wired,
the rooms are empty. What the clowns do is your department.

## Run it

```sh
tinygo flash -target=xiao-esp32c3 ./examples/grove-clown-house/
tinygo monitor
```

Silence in the monitor means every configured device answered. Unplug the
display and reset the board to watch `hardwareInit` name the missing tenant.

## Experiments

Furnish the house — each idea uses only chapter knowledge:

1. **Doorbell**: touch pad rings the buzzer. Two devices, five lines, instant
   gratification.
2. **Dashboard**: temperature and knob angle on the OLED, refreshed every
   200ms.
3. **Party mode**: RGB comet whose speed follows the sound sensor's loudness.
4. **Burglar alarm**: accelerometer detects the house moving → siren on the
   buzzer, red flashing lights, `INTRUDER` on the display. Touch pad to
   disarm.
5. Your own clowns. All eight devices are initialized and waiting.

## Takeaways

Integration is a planning discipline: connector constraints decide the wiring,
one visible block declares the hardware, and initialization errors get an
explicit place to land — because on embedded, unhandled means invisible. For
the kit's final act we leave the board and stream data to your computer:
**[the sound recorder](sound-recorder.md)**.

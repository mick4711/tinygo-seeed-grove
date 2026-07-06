# 10. Accelerometer — I2C registers and gravity

**register**: a small named storage location inside a chip, read and written
to control it or fetch its measurements.

The final module is the most sophisticated: the LIS3DH, a chip that measures
acceleration along three axes thousands of times per second. It's also our
window into how I2C devices *really* work under every driver you'll ever call:
as a map of registers.

<!-- TODO: photo of Grove 3-axis accelerometer module -->

## The concept

Inside the LIS3DH a microscopic silicon mass hangs on silicon springs. When
the chip accelerates, the mass lags behind, and the displacement is measured
electrically along X, Y and Z. Here is the beautiful part: sitting still on
your desk, it does **not** read zero. Gravity pulls the mass down exactly as
acceleration would, so Z reads one full g — about 9.8 m/s², reported by our
driver as ~1000 milli-g. An accelerometer at rest is a plumb line; tilt the
board and gravity's 1000 milli-g redistributes among the axes. That one fact
is how your phone knows which way is up.

And how do we talk to it? Like the SSD1306, the LIS3DH is an I2C target — but
where the display swallowed a stream of pixels, the accelerometer exposes a
tidy **register map**: address `0x20` (CTRL_REG1) sets the sampling rate,
`0x23` (CTRL_REG4) the measurement range, `0x28` through `0x2D` hold the
latest X, Y, Z readings as 16-bit values. Configuring the chip is writing
bytes at register addresses; measuring is reading six bytes back. Every I2C
sensor in existence — gyros, barometers, magnetometers — works exactly like
this.

## Wiring

Plug the accelerometer into connector **3** (`shield.Conn(3)`) — the other
I2C socket. Both I2C sockets share the same bus, so the display can stay
plugged into 4: the two chips coexist at different addresses.

## The code

<!-- code: examples/examples-grove-single/grove-accelerometer-3axis/grove-accelerometer.go -->
```go
package main

import (
	"time"

	grove "github.com/soypat/seeed-grove"
	"github.com/soypat/seeed-grove/drivers/lis3dh"
	"tinygo.org/x/drivers"
)

const ConnPos = 3

var (
	shield        grove.ShieldXiao
	accelerometer lis3dh.Device
)

func main() {
	time.Sleep(time.Second)

	accelerometer = lis3dh.New(shield.Conn(ConnPos).I2C(100_000))
	if err := accelerometer.Configure(lis3dh.Config{}); err != nil {
		panic(err)
	}

	for {
		// Acceleration returns the last read values; Update fetches fresh ones.
		if err := accelerometer.Update(drivers.Acceleration); err != nil {
			panic(err)
		}
		x, y, z := accelerometer.Acceleration()
		println("x", x, "y", y, "z", z)
		time.Sleep(500 * time.Millisecond)
	}
}
```

[view source](../../examples/examples-grove-single/grove-accelerometer-3axis/grove-accelerometer.go)

## Walkthrough

```go
accelerometer = lis3dh.New(shield.Conn(ConnPos).I2C(100_000))
if err := accelerometer.Configure(lis3dh.Config{}); err != nil {
```

`Configure` is register writes, nothing more. Open
[drivers/lis3dh](../../drivers/lis3dh/lis3dh.go) and you'll find lines like
`d.writeReg(regCtrl4, 0x80|byte(cfg.Range)<<4)` — bit-packing values into
control registers, positions dictated by the datasheet's register tables. The
empty `Config{}` picks the defaults Seeed calibrated for this module: ±16g
range at 400 Hz.

The read side splits into two calls, a convention shared by all TinyGo
drivers: `Update()` performs the actual I2C transaction (reads the six output
registers), and `Acceleration()` just returns those stored values, scaled to
milli-g. The scaling constant depends on the configured range — at ±16g the
chip resolves 1280 counts per g; at ±2g, 16000. That's the classic sensor
trade: range versus resolution, chosen with one register field.

## Run it

```sh
tinygo flash -target=xiao-esp32c3 ./examples/examples-grove-single/grove-accelerometer-3axis/
tinygo monitor
```

Flat on the desk you should see `x` and `y` near 0 and `z` near **1000** —
you are watching gravity. Stand the board on its edge and the 1000 migrates
to another axis. Shake it and the numbers fly.

## Experiments

1. Tilt meter: OLED on connector 4 (same bus!), draw a dot whose screen
   position follows X and Y tilt. A bubble level, from two chapters.
2. Shake alarm: total motion is `|x|+|y|+|z|` far from 1000. Buzzer on
   connector 1, and you've built the heart of every bike alarm.
3. Free-fall detector: during a (gentle, over-a-pillow) drop, all three axes
   read near *zero* — the mass and chip fall together. Detect it and print.
   Ask yourself why zero, then enjoy having understood Einstein's happiest
   thought.
4. Change `lis3dh.Config{Range: lis3dh.Range2G}` and park the board: how much
   finer is the resting reading than at ±16g?

## Takeaways

Under every polished driver API lies a register map — configuration is writing
bits at documented addresses, measurement is reading them back — and a chip's
defaults are somebody's chosen trade-off (here: range over resolution). You
have now toured every device in the box and every concept it teaches. Time to
use several at once: **[the projects](../projects/README.md)**.

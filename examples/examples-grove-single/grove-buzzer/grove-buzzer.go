package main

import (
	"time"

	grove "github.com/soypat/seeed-grove"
)
/*
Note	Freq	Period	Half Period
do	261.6	3822.629969	1911.314985
re	293.7	3404.834866	1702.417433
me	329.6	3033.980583	1516.990291
fa	349.2	2863.688431	1431.844215
so	392		2551.020408	1275.510204
la	440		2272.727273	1136.363636
te	493.9	2024.701357	1012.350678
do	523.2	1911.314985	955.6574924
*/
const ConnPosition = 0

// Define the hardware we are using.
var shield grove.ShieldXiao

func main() {
	buzzer := shield.Conn(ConnPosition).PinOutput()
	// Toggle the pin to make a square wave; half-period sets the pitch.
	const do = 1911 * time.Microsecond // 261.6 Hz
	const re = 1702 * time.Microsecond // 293.7 Hz
	const me = 1516 * time.Microsecond // 329.6 Hz
	const fa = 1432 * time.Microsecond // 349.2 Hz
	const so = 1275 * time.Microsecond // 392.0 Hz
	const la = 1136 * time.Microsecond // 440.0 Hz
	const te = 1012 * time.Microsecond // 493.9 Hz
	const hd = 956 * time.Microsecond  // 523.2 Hz
	for {
		println("do", do)
		beep(buzzer, do, 200*time.Millisecond)
		time.Sleep(time.Second)

		println("re", re)
		beep(buzzer, re, 200*time.Millisecond)
		time.Sleep(time.Second)

		println("me")
		beep(buzzer, me, 200*time.Millisecond)
		time.Sleep(time.Second)

		println("fa")
		beep(buzzer, fa, 200*time.Millisecond)
		time.Sleep(time.Second)

		println("so")
		beep(buzzer, so, 200*time.Millisecond)
		time.Sleep(time.Second)

		println("la")
		beep(buzzer, la, 200*time.Millisecond)
		time.Sleep(time.Second)

		println("te")
		beep(buzzer, te, 200*time.Millisecond)
		time.Sleep(time.Second)

		println("hd")
		beep(buzzer, hd, 200*time.Millisecond)
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

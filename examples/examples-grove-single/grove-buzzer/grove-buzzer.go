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
const ConnPosition = 2

// Define the hardware we are using.
var shield grove.ShieldXiao

func main() {
	buzzer := shield.Conn(ConnPosition).PinOutput()
	// playScale(buzzer)
	for {
		wail(buzzer)
	}
}

func wail(buzzer grove.PinOutput) {
	println("going up")
	for t := 1000 * time.Microsecond; t >= 300 * time.Microsecond; t -= 20 * time.Microsecond {
		beep(buzzer, t, 50*time.Millisecond)
		time.Sleep(time.Millisecond)
	}
	time.Sleep(50 * time.Millisecond)

	println("going down")
	for t := 300 * time.Microsecond; t <= 1000 * time.Microsecond; t += 20 * time.Microsecond {
		beep(buzzer, t, 50*time.Millisecond)
		time.Sleep(time.Millisecond)
	}
	time.Sleep(50 * time.Millisecond)
}

func playScale(buzzer grove.PinOutput) {
	us := 1* time.Microsecond
	scale := []time.Duration{
		1911 * us, 
		1702 * us, 
		1516 * us, 
		1432 * us, 
		1275 * us, 
		1136 * us, 
		1012 * us, 
		956 * us, 
	}
	for {
		for _, note := range scale {
			println("note", note)
			beep(buzzer, note, 200*time.Millisecond)
			time.Sleep(time.Second)
		}
	}
}

func beep(pin grove.PinOutput, halfPeriod, dur time.Duration) {
	// Toggle the pin to make a square wave; half-period sets the pitch.
	for elapsed := time.Duration(0); elapsed < dur; elapsed += 2 * halfPeriod {
		pin.High()
		time.Sleep(halfPeriod)
		pin.Low()
		time.Sleep(halfPeriod)
	}
}

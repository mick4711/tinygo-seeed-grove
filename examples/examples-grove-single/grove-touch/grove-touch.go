package main

import (
	"time"

	grove "github.com/soypat/seeed-grove"
)

// Define the hardware we are using.
var (
	shield grove.ShieldXiao
	touch       grove.PinInput
	buzzer      grove.PinOutput
)

func main() {
	// Capacitive touch drives the data line high while a finger is present.
	touch = shield.Conn(6).PinInputPulldown()
	buzzer = shield.Conn(7).PinOutput()
	lastPrint := time.Now()
	last := false
	for {
		beingTouched := touch.GetLevel()
		if beingTouched && time.Since(lastPrint) > time.Second {
			println("i am being touched!")
			buzz(buzzer, true)
			lastPrint = time.Now()
		} else if last && !beingTouched {
			println("oh no, I am getting cold without your touch :(")
			buzz(buzzer, false)
		}
		last = beingTouched
		time.Sleep(20 * time.Millisecond)
	}
}

func buzz(pin grove.PinOutput, state bool) {
	if state {
		halfPeriod := 500 * time.Microsecond
		dur := 1000 * time.Millisecond
		for elapsed := time.Duration(0); elapsed < dur; elapsed += 2 * halfPeriod {
			pin.High()
			time.Sleep(halfPeriod)
			pin.Low()
			time.Sleep(halfPeriod)
		}
	} else {
		pin.Low()
	}
}

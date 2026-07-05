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

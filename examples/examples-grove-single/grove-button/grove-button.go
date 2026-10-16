package main

import (
	grove "github.com/soypat/seeed-grove"
)

func main() {
	const (
		ConnPosition = 0
		on, off = true, false
	)

	var (
		shield grove.ShieldXiao
		button grove.PinInput
		count int
		state, currState bool
	)

	button = shield.Conn(ConnPosition).PinInputPulldown()
	for {
		// detect transition from 'off' to 'on' and increment count
		state = button.GetLevel()

		switch state {
		case currState:
			// continue
		case on:
			// transtioned to on
			currState = state
			count++
			println("count =", count)
		case off:
			// transitioned to off
			currState = state
		}
	}
}

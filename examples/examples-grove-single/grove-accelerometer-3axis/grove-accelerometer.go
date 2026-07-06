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

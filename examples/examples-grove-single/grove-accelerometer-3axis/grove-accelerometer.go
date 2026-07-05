package main

import (
	"time"

	grove "github.com/soypat/seeed-grove"
	"tinygo.org/x/drivers/lis3dh"
)

const ConnPos = 3

// Define the hardware we are using.
var (
	shield        grove.ShieldXiao
	accelerometer lis3dh.Device
)

func main() {
	time.Sleep(time.Second)
	accelerometer = lis3dh.New(shield.Conn(ConnPos).I2C(10_000))
	// Seeed's Grove LIS3DHTR module straps SA0 high, so its I2C address is 0x19
	// (Address1), not the driver's 0x18 default.
	err := accelerometer.Configure(lis3dh.Config{Address: lis3dh.Address1})
	if err != nil {
		panic(err)
	}
	for {
		x, y, z := accelerometer.Acceleration()
		println("x", x, "y", y, "z", z)
		time.Sleep(time.Second)
	}
}

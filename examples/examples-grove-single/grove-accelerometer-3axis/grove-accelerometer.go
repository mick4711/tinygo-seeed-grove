package main

import (
	"time"

	grove "github.com/soypat/seeed-grove"
	"github.com/soypat/seeed-grove/drivers/lis3dh"
	"github.com/soypat/seeed-grove/drivers/fonts"
	"github.com/soypat/seeed-grove/drivers/ssd1306"
	"tinygo.org/x/drivers"
)

const (
	ConnPosAcc  = 3
	ConnPosOled  = 4
	width, height = 64, 48
	frameTime     = 30 * time.Millisecond
)

var (
	shield        grove.ShieldXiao
	accelerometer lis3dh.Device
	display ssd1306.Device
	font    = &fonts.Font5x7
)

func main() {
	time.Sleep(time.Second)

	accelerometer = lis3dh.New(shield.Conn(ConnPosAcc).I2C(100_000))
	if err := accelerometer.Configure(lis3dh.Config{}); err != nil {
		panic(err)
	}

	if err := display.ConfigureI2C(shield.Conn(ConnPosOled).I2C(400_000),
		ssd1306.Config{
			Width:  width,
			Height: height,
			Buffer: make([]byte, ssd1306.BufferSize(width, height)),
		}); err != nil {
		panic(err)
	}
	drawer := ssd1306.Drawer{
		DisplayWidth: width,
		Buffer:       display.CurrentDisplayBuffer(),
	}

	drawer.DrawPixel(0, 0, true)
	for {
		// Acceleration returns the last read values; Update fetches fresh ones.
		if err := accelerometer.Update(drivers.Acceleration); err != nil {
			panic(err)
		}
		x, y, z := accelerometer.Acceleration()
		drawPixel(&drawer, x, y)
		println("x", x, "y", y, "z", z)
		time.Sleep(500 * time.Millisecond)
	}
}

func drawPixel(drw *ssd1306.Drawer, xa, ya int32) {
	drw.Clear()
	drw.DrawPixel(0, 0, true)
	x := int(float32(width/2) * (1.0 + float32(xa)/1200.0))
	y := int(float32(height/2) * (1.0 - float32(ya)/1200.0))
	drw.DrawPixel(x, y, true)
	display.Display()
}

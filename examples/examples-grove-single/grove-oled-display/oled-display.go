package main

import (
	"time"

	grove "github.com/soypat/seeed-grove"
	"github.com/soypat/seeed-grove/drivers/ssd1306"
)

const (
	ConnPosition  = 4
	width, height = 64, 48
)

// Define the hardware we are using.
var (
	shield  grove.ShieldXiao
	display ssd1306.Device
)

func main() {
	time.Sleep(time.Second) // wait for USB connection.
	err := display.ConfigureI2C(shield.Conn(ConnPosition).I2C(400_000), ssd1306.Address_128_32, ssd1306.Config{
		Width:  width,
		Height: height,
		Buffer: make([]byte, ssd1306.BufferSize(width, height)),
	})
	if err != nil {
		panic(err)
	}
	drawer := ssd1306.Drawer{
		DisplayWidth: width,
		Buffer:       display.CurrentDisplayBuffer(),
	}
	drawer.Clear()
	display.Display()

	for {
		for x := 0; x < width; x++ {
			for y := 0; y < height; y++ {
				drawer.SetPixel(x, y, true)
				display.Display()
			}
		}
		time.Sleep(time.Second / 2)
		drawer.Clear()
		display.Display()
		time.Sleep(time.Second / 2)
		drawer.PatternSet(0b10101010)
		display.Display()
		time.Sleep(time.Second)
		drawer.Clear()
	}

}

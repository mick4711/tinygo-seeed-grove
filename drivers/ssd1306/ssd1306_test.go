package ssd1306

import (
	"fmt"
	"math/rand"
	"testing"
)

func ExampleAppendFormatDisplay() {
	width := 16
	height := 8
	drawer := Drawer{
		DisplayWidth: width,
		Buffer:       make([]byte, BufferSize(width, height)),
	}
	err := drawer.DrawRectangle(1, 2, 2, 5, true)
	if err != nil {
		panic(err)
	}
	err = drawer.DrawPixel(6, 2, true)
	if err != nil {
		panic(err)
	}
	err = drawer.DrawCircle(7, 4, 3, true)
	if err != nil {
		panic(err)
	}
	draw := AppendFormatDisplay(nil, drawer.Buffer, width, '█', '░')
	fmt.Println(string(draw))
	//output:
	// ░░░░░░░░░░░░░░░░
	// ░░░░░░███░░░░░░░
	// ░██░░██░░█░░░░░░
	// ░██░█░░░░░█░░░░░
	// ░██░█░░░░░█░░░░░
	// ░██░█░░░░░█░░░░░
	// ░██░░█░░░█░░░░░░
	// ░░░░░░███░░░░░░░
}

func ExampleDrawer_DrawFilledCircle() {
	width := 16
	height := 8
	drawer := Drawer{
		DisplayWidth: width,
		Buffer:       make([]byte, BufferSize(width, height)),
	}
	err := drawer.DrawFilledCircle(7, 4, 3, true)
	if err != nil {
		panic(err)
	}
	draw := AppendFormatDisplay(nil, drawer.Buffer, width, '█', '░')
	fmt.Println(string(draw))
	//output:
	// ░░░░░░░░░░░░░░░░
	// ░░░░░░███░░░░░░░
	// ░░░░░█████░░░░░░
	// ░░░░███████░░░░░
	// ░░░░███████░░░░░
	// ░░░░███████░░░░░
	// ░░░░░█████░░░░░░
	// ░░░░░░███░░░░░░░
}

func BenchmarkDrawDisplayCircle(b *testing.B) {
	const width, height = 64, 48
	var display Device
	err := display.ConfigureI2C(mockI2C{}, 0, Config{
		Width:  width,
		Height: height,
		Buffer: make([]byte, BufferSize(width, height)),
	})
	if err != nil {
		b.Fatal(err)
	}
	drawer := Drawer{
		DisplayWidth: width,
		Buffer:       display.CurrentDisplayBuffer(),
	}
	rng := rand.New(rand.NewSource(0))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		x, y := rng.Intn(width-2)+1, rng.Intn(height-2)+1
		radius := min(x, y, width-1-x, height-1-y) // largest circle that fits.
		if err := drawer.DrawCircle(x, y, radius, true); err != nil {
			b.Fatal(err)
		}
		display.Display()
	}
}

type mockI2C struct {
}

func (mockI2C) Tx(addr uint16, w, r []byte) error {
	return nil
}

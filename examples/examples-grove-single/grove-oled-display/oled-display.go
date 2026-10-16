package main

import (
	"fmt"
	"time"

	grove "github.com/soypat/seeed-grove"
	"github.com/soypat/seeed-grove/drivers/fonts"
	"github.com/soypat/seeed-grove/drivers/ssd1306"
)

const (
	ConnTemp  = 0
	ConnOled  = 4
	width, height = 64, 48
	frameTime     = 30 * time.Millisecond
	NumSamples = 8192
)

// Define the hardware we are using.
var (
	shield  grove.ShieldXiao
	display ssd1306.Device
	temp   grove.SensorTemperature
	// rotary grove.SensorRotaryAngle
	font    = &fonts.Font5x7
)

func main() {
	temp.ConfigureThermistor(shield.Conn(ConnTemp).ADC(), NumSamples)
	// rotary.Configure(shield.Conn(0).ADC(), 1)
	time.Sleep(time.Second) // wait for USB connection.
	err := display.ConfigureI2C(shield.Conn(ConnOled).I2C(400_000), ssd1306.Config{
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
	for {
		// splash(&drawer)
		// bounce(&drawer)
		// ping(&drawer)
		// heartbeat(&drawer)
		// loading(&drawer)
		// loadRotary(&drawer, rotary.Fraction())
		temperature(&drawer, temp.ReadTemperature())
	}
}

// display temperature in degC.
func temperature(drw *ssd1306.Drawer, temp float32) {
	drw.Clear()
	border(drw)

	centerText(drw, (height / 2) - 12, "Temp:")
	disp := fmt.Sprintf("%.2f C", temp)
	centerText(drw, (height / 2) + 3, disp)

	display.Display()
	fmt.Printf("time: %s celsius: %.3f\n", time.Now().Format("03 hr - 04 min"), temp)

	time.Sleep(5 * time.Second)
}

// splash shows a framed title card.
func splash(drw *ssd1306.Drawer) {
	drw.Clear()
	border(drw)
	centerText(drw, 6, "Grove")
	centerText(drw, 16, "OLED")
	// Ruler with dot endpoints under the title.
	drw.DrawLine(12, 28, width-13, 28, true)
	drw.DrawFilledCircle(9, 28, 1, true)
	drw.DrawFilledCircle(width-10, 28, 1, true)
	centerText(drw, 34, "64x48")
	display.Display()
	time.Sleep(2500 * time.Millisecond)
}

// bounce animates a ball reflecting off the frame above a caption.
func bounce(drw *ssd1306.Drawer) {
	const radius = 3
	const floor = 35 // separator line above the caption.
	x, y := 20, 12
	dx, dy := 1, 1
	for i := 0; i < 200; i++ {
		drw.Clear()
		border(drw)
		drw.DrawLine(1, floor, width-2, floor, true)
		centerText(drw, floor+3, "bounce!")
		drw.DrawFilledCircle(x, y, radius, true)
		display.Display()

		x, dx = step(x, dx, 1+radius+1, width-2-radius-1)
		y, dy = step(y, dy, 1+radius+1, floor-radius-2)
		time.Sleep(frameTime)
	}
}

// ping expands concentric circles from the center, sonar style.
func ping(drw *ssd1306.Drawer) {
	const cx, cy, maxRadius, spacing = width / 2, 19, 14, 5
	for i := 0; i < 5; i++ {
		for r := 1; r <= maxRadius; r++ {
			drw.Clear()
			border(drw)
			drw.DrawLine(1, 35, width-2, 35, true)
			centerText(drw, 38, "ping...")
			drw.DrawFilledCircle(cx, cy, 1, true)
			for wave := r; wave > 0; wave -= spacing {
				drw.DrawCircle(cx, cy, wave, true)
			}
			display.Display()
			time.Sleep(2 * frameTime)
		}
	}
}

// draws filled circles getting bigger and smaller.
func heartbeat(drw *ssd1306.Drawer) {
	const cx, cy, maxRadius, spacing = width / 2, height / 2, 14, 5
	for i := 0; i < 5; i++ {
		for r := 1; r <= maxRadius; r++ {
			drw.Clear()
			border(drw)
			drw.DrawFilledCircle(cx, cy, r, true)
			// drw.DrawFilledCircle(cx, cy, 1, true)
			// for wave := r; wave > 0; wave -= spacing {
			// 	drw.DrawCircle(cx, cy, wave, true)
			// }
			display.Display()
			time.Sleep(2 * frameTime)
		}
		for r := maxRadius; r >= 1; r-- {
			drw.Clear()
			border(drw)
			drw.DrawFilledCircle(cx, cy, r, true)
			// drw.DrawFilledCircle(cx, cy, 1, true)
			// for wave := r; wave > 0; wave -= spacing {
			// 	drw.DrawCircle(cx, cy, wave, true)
			// }
			display.Display()
			time.Sleep(2 * frameTime)
		}
	}
}

// loading fills a progress bar proportional to rotary fraction.
func loadRotary(drw *ssd1306.Drawer, frac float32) {
	const barX, barY, barW, barH = 6, 20, width - 12, 9
	p := int(frac * 100.0)

	drw.Clear()
	border(drw)
	centerText(drw, 7, "percent")
	drw.DrawRectangle(barX, barY, barW, 1, true)
	drw.DrawRectangle(barX, barY+barH-1, barW, 1, true)
	drw.DrawRectangle(barX, barY, 1, barH, true)
	drw.DrawRectangle(barX+barW-1, barY, 1, barH, true)
	if fill := p * (barW - 4) / 100; fill > 0 {
		drw.DrawRectangle(barX+2, barY+2, fill, barH-4, true)
	}
	// Center "NN%" without building a string on the heap.
	x := (width - (numberWidth(p) + font.StringWidth("%"))) / 2
	x = drawNumber(drw, x, 34, p)
	drw.DrawText(x, 34, "%", font, true)

	display.Display()
	time.Sleep(200 * time.Millisecond)
}

// loading fills a progress bar and prints the percentage under it.
func loading(drw *ssd1306.Drawer) {
	const barX, barY, barW, barH = 6, 20, width - 12, 9
	for p := 0; p <= 100; p++ {
		drw.Clear()
		border(drw)
		centerText(drw, 7, "loading")
		drw.DrawRectangle(barX, barY, barW, 1, true)
		drw.DrawRectangle(barX, barY+barH-1, barW, 1, true)
		drw.DrawRectangle(barX, barY, 1, barH, true)
		drw.DrawRectangle(barX+barW-1, barY, 1, barH, true)
		if fill := p * (barW - 4) / 100; fill > 0 {
			drw.DrawRectangle(barX+2, barY+2, fill, barH-4, true)
		}
		// Center "NN%" without building a string on the heap.
		x := (width - (numberWidth(p) + font.StringWidth("%"))) / 2
		x = drawNumber(drw, x, 34, p)
		drw.DrawText(x, 34, "%", font, true)
		display.Display()
		time.Sleep(frameTime)
	}
	time.Sleep(time.Second)
}

// border draws the 1 pixel screen frame every scene shares.
func border(drw *ssd1306.Drawer) {
	drw.DrawLine(0, 0, width-1, 0, true)
	drw.DrawLine(0, height-1, width-1, height-1, true)
	drw.DrawLine(0, 0, 0, height-1, true)
	drw.DrawLine(width-1, 0, width-1, height-1, true)
}

// centerText draws s horizontally centered at row y.
func centerText(drw *ssd1306.Drawer, y int, s string) {
	drw.DrawText((width-font.StringWidth(s))/2, y, s, font, true)
}

// drawNumber draws n in decimal at (x, y) and returns the next x. Digits are
// sliced from a constant string so nothing is heap allocated.
func drawNumber(drw *ssd1306.Drawer, x, y, n int) int {
	const digits = "0123456789"
	if n >= 10 {
		x = drawNumber(drw, x, y, n/10)
	}
	d := n % 10
	x, _ = drw.DrawText(x, y, digits[d:d+1], font, true)
	return x
}

// numberWidth returns the pixel width drawNumber uses for n.
func numberWidth(n int) int {
	const digits = "0123456789"
	w := font.StringWidth(digits[n%10 : n%10+1])
	for n >= 10 {
		n /= 10
		w += font.StringWidth(digits[n%10 : n%10+1])
	}
	return w
}

// step advances position p by velocity v, reflecting at lo and hi inclusive.
func step(p, v, lo, hi int) (int, int) {
	p += v
	if p <= lo || p >= hi {
		v = -v
	}
	return p, v
}

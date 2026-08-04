// gophercon-demo turns the sound sensor and the OLED display into a
// GopherCon badge: a single screen with the conference title on top, an eight
// band audio spectrum below it, and the year at the foot. The title inverts
// while the room is loud, so a clap lights the badge up.
//
// The bands come from a Goertzel filter bank instead of an FFT. Only eight
// bins are ever drawn, and each one is a two tap resonator run over the sample
// block, so the whole spectrum costs less than an FFT and allocates nothing.
package main

import (
	"time"

	"github.com/chewxy/math32"
	grove "github.com/soypat/seeed-grove"
	"github.com/soypat/seeed-grove/drivers/fonts"
	"github.com/soypat/seeed-grove/drivers/ssd1306"
)

const (
	ConnSound   = 0 // ADC pin.
	ConnDisplay = 3 // I2C.

	width, height = 64, 48
	sampleRate    = 8000 // Hz. Ceiling for software-polled ADC on the C3.
	blockSize     = 128  // Audio samples per frame: 16ms, bins 62.5Hz apart.
	numBands      = 8
)

// Screen layout in pixel rows: title on the top page, bars growing up from
// the baseline, year underneath.
const (
	titleY   = 0
	rulerY   = 9
	barTopY  = 12
	barBaseY = 37 // Bottom row of the bars, the baseline sits below it.
	baseY    = 38
	footerY  = 40

	barWidth = 5
	barGap   = 2
	barsX    = (width - (numBands*(barWidth+barGap) - barGap)) / 2
	barSpan  = barBaseY - barTopY + 1

	// applauseSpan is the bar height that counts as a room clapping.
	applauseSpan = 3 * barSpan / 4
)

// bandBins holds the Goertzel bin of each bar, bin k covering
// k*sampleRate/blockSize Hz. Roughly log spaced from 125Hz to 1.1kHz, the
// range where voice and applause live.
var bandBins = [numBands]int{2, 3, 4, 6, 8, 11, 14, 18}

// Define the hardware we are using.
var (
	shield  grove.ShieldXiao
	display ssd1306.Device
	sound   grove.SensorSound
	font    = &fonts.Font5x7
)

func main() {
	time.Sleep(time.Second) // wait for USB connection.
	err := sound.Configure(shield.Conn(ConnSound).ADC(), sampleRate)
	if err != nil {
		panic(err)
	}
	err = display.ConfigureI2C(shield.Conn(ConnDisplay).I2C(400_000), ssd1306.Config{
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
	// A Goertzel coefficient only depends on its bin, so the bank is built once.
	var coeff [numBands]float32
	for i, k := range bandBins {
		coeff[i] = 2 * math32.Cos(2*math32.Pi*float32(k)/blockSize)
	}

	var block [blockSize]float32
	var bars, peaks [numBands]int
	for frame := 0; ; frame++ {
		record(&block)
		loudest := 0
		for i := range bars {
			h := barHeight(goertzel(block[:], coeff[i]))
			// Bars jump up to the new level and slide back down a pixel per
			// frame, the classic analyzer look. Peak caps hang above them and
			// fall slower so short transients stay readable.
			if h > bars[i] {
				bars[i] = h
			} else if bars[i] > 0 {
				bars[i]--
			}
			if bars[i] >= peaks[i] {
				peaks[i] = bars[i]
			} else if peaks[i] > 0 && frame%3 == 0 {
				peaks[i]--
			}
			if bars[i] > loudest {
				loudest = bars[i]
			}
		}
		draw(&drawer, &bars, &peaks, loudest >= applauseSpan)
	}
}

// record fills block with one frame of audio centered on its own DC bias, so
// the filter bank sees the signal swing instead of the ADC midpoint.
func record(block *[blockSize]float32) {
	sound.Sync() // Drawing the last frame ate into the sampling clock, restart it.
	var sum float32
	for i := range block {
		v := float32(sound.Sample())
		block[i] = v
		sum += v
	}
	bias := sum / blockSize
	for i := range block {
		block[i] -= bias
	}
}

// goertzel returns the squared magnitude of the block's content at the bin
// coeff was built for: what a full FFT would report for that one bin, for the
// cost of a single multiply-accumulate pass.
// See https://en.wikipedia.org/wiki/Goertzel_algorithm.
func goertzel(block []float32, coeff float32) float32 {
	var s1, s2 float32
	for _, x := range block {
		s0 := x + coeff*s1 - s2
		s2, s1 = s1, s0
	}
	return s1*s1 + s2*s2 - coeff*s1*s2
}

// barHeight maps a squared magnitude to a bar height. Hearing is
// logarithmic, so the scale is too: every pxPerLog2 pixels doubles the power,
// starting from the noise floor of a quiet room. Both constants depend on the
// gain the sound sensor's potentiometer is set to; turn floorLog2 up if the
// bars never rest at zero, down if they never leave it.
func barHeight(mag2 float32) int {
	const (
		floorLog2 = 26 // log2 of a quiet room's squared magnitude.
		pxPerLog2 = 2  // 2px per power doubling, so 4px per ~6dB of amplitude.
	)
	if mag2 <= 0 {
		return 0
	}
	h := int((math32.Log2(mag2) - floorLog2) * pxPerLog2)
	if h > barSpan {
		h = barSpan
	} else if h < 0 {
		h = 0
	}
	return h
}

// draw paints the whole screen and pushes the frame out over I2C.
func draw(drw *ssd1306.Drawer, bars, peaks *[numBands]int, applause bool) {
	drw.Clear()
	if applause {
		drw.DrawRectangle(0, titleY, width, font.Height(), true)
	}
	centerText(drw, titleY, "GopherCon", !applause)
	drw.DrawLine(0, rulerY, width-1, rulerY, true)
	for i, h := range bars {
		x := barsX + i*(barWidth+barGap)
		if h > 0 {
			drw.DrawRectangle(x, barBaseY+1-h, barWidth, h, true)
		}
		if p := peaks[i]; p > 0 {
			drw.DrawLine(x, barBaseY-p, x+barWidth-1, barBaseY-p, true)
		}
	}
	drw.DrawLine(0, baseY, width-1, baseY, true)
	centerText(drw, footerY, "2026", true)
	display.Display()
}

// centerText draws s horizontally centered at row y.
func centerText(drw *ssd1306.Drawer, y int, s string, on bool) {
	drw.DrawText((width-font.StringWidth(s))/2, y, s, font, on)
}

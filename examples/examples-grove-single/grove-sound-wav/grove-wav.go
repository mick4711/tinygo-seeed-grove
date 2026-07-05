// grove-wav streams raw microphone audio from a Grove analog sound sensor over
// USB serial as fixed-length frames of signed 16-bit little-endian PCM samples,
// centered on the measured DC bias. It writes no WAV header and never stops; the
// host (see ./capture) locks onto the frame marker and wraps the samples in WAV.
//
// Framing: each frame is a 4-byte marker (0xDEADBEEF) followed by FrameSamples
// audio samples. The marker is non-zero, so it cannot be confused with silence
// (which is 0x0000 samples) — unlike a zero sync word, whose bytes merge with
// adjacent near-zero samples and cause byte-misaligned locks (static).
//
// The ESP32-C3 USB-serial-JTAG port is separate from the boot-ROM UART logs, so
// the stream stays clean as long as this program prints nothing else.
package main

import (
	"encoding/binary"
	"machine"
	"time"

	grove "github.com/soypat/seeed-grove"
)

const (
	// SampleRate must match the value the capture tool uses to build the WAV.
	SampleRate = 8000 // Hz. Ceiling for software-polled ADC on the C3.

	FrameSamples = 128 // audio samples per frame.
	frameBytes   = 4 + FrameSamples*2

	// biasShift sets the DC-bias tracker time constant: 2^biasShift samples
	// (512 → 64ms at 8kHz, a ~2.5Hz high-pass). Long enough to pass voice
	// untouched, short enough to follow supply/temperature drift.
	biasShift = 9
)

// marker prefixes every frame. Must match the capture tool.
var marker = [4]byte{0xDE, 0xAD, 0xBE, 0xEF}

// Define the hardware we are using.
var (
	shield grove.ShieldXiao
	sound  grove.SensorSound
)

func main() {
	time.Sleep(time.Second)
	err := sound.Configure(shield.Conn(0).ADC(), SampleRate)
	if err != nil {
		panic(err)
	}
	var frame [frameBytes]byte
	copy(frame[:], marker[:])
	// One-pole DC-bias tracker, seeded from the first readings so the stream
	// doesn't open with a thump while the estimate converges.
	biasAcc := int32(sound.SampleNoWait(1<<4)>>4) << biasShift
	for {
		sound.Sync()
		for j := 0; j < FrameSamples; j++ {
			raw := int32(sound.Sample())
			bias := biasAcc >> biasShift
			biasAcc += raw - bias
			// Center on the bias and saturate: without the clamp, peaks past
			// int16 range wrap to the opposite sign — a full-scale click.
			v := raw - bias
			if v > 32767 {
				v = 32767
			} else if v < -32768 {
				v = -32768
			}
			binary.LittleEndian.PutUint16(frame[4+j*2:], uint16(int16(v)))
		}
		machine.Serial.Write(frame[:])
	}
}

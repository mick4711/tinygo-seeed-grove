# Project: the sound recorder

**telemetry**: data a device transmits back for storage or analysis elsewhere.

The sound chapter measured loudness and threw the samples away. This project
keeps them: the board streams live microphone audio over the USB cable, a Go
program on your computer catches the stream, and out comes `rec.wav` — a file
you can actually play. Two programs, two computers, one wire. Welcome to
systems programming.

## The concept

The plan sounds trivial — read samples, write them to serial, save on the PC —
but three classic streaming problems stand in the way, and solving them is the
whole project:

**Where does a frame start?** Serial gives the host a raw byte soup with no
boundaries; start reading mid-stream and every 16-bit sample is split down the
middle. So the device sends fixed-size **frames**, each prefixed with the
marker bytes `0xDEADBEEF`. Why not zeros? Silence *is* zeros — a zero marker
would dissolve into the audio around it, while a loud non-zero pattern can be
found and verified by its regular spacing.

**Where is zero?** The module rides its audio on a DC bias (sound chapter),
and the bias wanders with temperature and supply. The device subtracts a
continuously-updated bias estimate from every sample, so the recording is
centered like proper audio.

**When do samples happen?** A WAV at 8000 Hz is a promise about *time*. The
device paces itself with deadlines so samples land 125µs apart on average —
8000 Hz being roughly the ceiling for software-polled ADC reads on this chip.

## The device program

<!-- code: examples/grove-sound-wav/grove-wav.go -->
```go
//go:build tinygo

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
```

[view source](../../examples/grove-sound-wav/grove-wav.go)

The bias tracker deserves a stare:

```go
bias := biasAcc >> biasShift
biasAcc += raw - bias
```

Two integer operations per sample buy a moving average of the last ~512
samples — a one-pole filter that passes voice untouched but tracks slow
drift. Note also the clamp before writing each sample: an overload past int16
range must *saturate*, because wrapping to the opposite sign is a full-scale
click in your ears.

And a subtle constraint: the program prints nothing. The same USB serial
channel carries the audio bytes; one stray `println` would inject garbage
samples into your recording.

## The host program

Plug the sound sensor into connector **0**, flash, then run the capture tool
(it needs the serial port to itself, so `tinygo monitor` must not be running):

```sh
tinygo flash -target=xiao-esp32c3 ./examples/grove-sound-wav/
cd examples/grove-sound-wav/capture
go run . /dev/ttyACM0    # then press the board's reset button
```

Note that's plain `go run` — this program runs on *your computer*, built by
the regular Go compiler, goroutines and filesystem and all. Same language on
both sides of the wire is the quiet superpower of this whole kit.

The capture tool ([capture.go](../../examples/grove-sound-wav/capture/capture.go))
buffers incoming bytes until it finds the marker recurring at exactly the
frame stride — several frames in a row, so a coincidental `0xDEADBEEF` in loud
audio can't fool it. Locked on, it strips markers, accumulates five seconds of
samples, prepends the 44-byte WAV header, and writes `rec.wav`. If a marker
ever fails to appear where expected, the stream slipped: it falls back to
re-locking rather than saving static.

Say something at the sensor while it captures, then:

```sh
mpv rec.wav   # or any audio player
```

That's your voice, sampled by a chip the size of a stamp. 8 kHz telephone
quality — Bell would recognize it.

## Experiments

1. Record 30 seconds instead of 5 (`seconds` in capture.go — device side
   needs no change; it streams forever).
2. Halve `SampleRate` to 4000 on **both** sides and listen. That muffled
   sound is aliasing and lost bandwidth — you can hear the sample theorem.
3. Break the framing on purpose: add a `println("hi")` to the device's main
   loop and watch capture handle the corruption. Now you know why the device
   stays silent.
4. Compute loudness on the host from `pcm` after capture and print a peak
   level, like real recording software.

## Takeaways

Streams need framing you can *find and verify* (non-zero marker + expected
periodicity), signals need their bias removed and their peaks saturated, and
splitting a system between a device and a host — each doing what it's good
at, speaking one agreed protocol — is the shape of nearly every real IoT
product. You've reached the end of the guide; the box is yours now. Go build
something and show us in the
[TinyGo community](https://tinygo.org/community/).

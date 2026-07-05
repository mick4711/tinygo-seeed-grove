// Reads the framed PCM stream from the grove-wav device, locks onto the frame
// marker, and wraps the audio into a WAV file. Run this, THEN reset the board.
//
//	go run . [/dev/ttyACM0]
package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"go.bug.st/serial"
)

// Must match the device firmware (grove-wav.go).
const (
	sampleRate   = 8000
	seconds      = 5
	frameSamples = 128
	frameBytes   = 4 + frameSamples*2
)

// marker prefixes every frame. Non-zero, so silence (0x0000 samples) can never
// forge it — the reason we don't use a zero sync word.
var marker = []byte{0xDE, 0xAD, 0xBE, 0xEF}

func main() {
	port := "/dev/ttyACM0"
	if len(os.Args) > 1 {
		port = os.Args[1]
	}

	// Baud is ignored on USB-CDC/JTAG; the port is raw (no line discipline),
	// so bytes pass through untouched — no UTF-8 decoding to mangle them.
	p, err := serial.Open(port, &serial.Mode{BaudRate: 115200})
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()
	p.SetReadTimeout(2 * time.Second)

	wantSamples := sampleRate * seconds
	pcm := make([]byte, 0, wantSamples*2)

	var buf []byte
	locked := false
	read := make([]byte, 4096)
	for len(pcm) < wantSamples*2 {
		n, err := p.Read(read)
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
		if n == 0 {
			if locked {
				break // stream ended.
			}
			fmt.Print("\rwaiting for device (reset the board)…")
			continue
		}
		buf = append(buf, read[:n]...)

		if !locked {
			off, ok := findLock(buf)
			if !ok {
				continue // need more bytes to confirm a boundary.
			}
			buf = buf[off:] // buf[0] is now a frame marker.
			locked = true
		}

		// Consume whole frames: drop the marker, keep the audio samples. If a
		// marker is ever missing, the stream slipped — fall back to re-locking.
		for len(buf) >= frameBytes {
			if !bytes.Equal(buf[0:4], marker) {
				locked = false
				break
			}
			pcm = append(pcm, buf[4:frameBytes]...)
			buf = buf[frameBytes:]
		}
		fmt.Printf("\rcaptured %d/%d samples (%.0f%%)", len(pcm)/2, wantSamples,
			float32(len(pcm)/2)/float32(wantSamples)*100)
	}
	fmt.Println()

	if len(pcm) > wantSamples*2 {
		pcm = pcm[:wantSamples*2]
	}
	if err := writeWAV("rec.wav", pcm); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote rec.wav (%d samples, %.2fs @ %dHz)", len(pcm)/2, float32(len(pcm)/2)/sampleRate, sampleRate)
}

// findLock returns the byte offset of a frame marker that recurs, aligned, at
// the frame stride for several frames. The marker is non-zero and 4 bytes, so
// neither silence nor a byte-misaligned read can forge it; the periodicity check
// rejects any coincidental marker in the audio. Returns false if more bytes are
// needed to confirm.
func findLock(buf []byte) (int, bool) {
	const verify = 3 // additional frames that must agree.
	need := frameBytes * (verify + 1)
	for start := 0; ; {
		rel := bytes.Index(buf[start:], marker)
		if rel < 0 {
			return 0, false
		}
		off := start + rel
		if off+need > len(buf) {
			return 0, false // candidate found but not enough data to verify yet.
		}
		ok := true
		for k := 1; k <= verify; k++ {
			if !bytes.Equal(buf[off+k*frameBytes:off+k*frameBytes+4], marker) {
				ok = false
				break
			}
		}
		if ok {
			return off, true
		}
		start = off + 1
	}
}

// writeWAV wraps signed-16-bit mono PCM in a canonical 44-byte-header WAV file.
func writeWAV(name string, pcm []byte) error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	le := func(v any) { binary.Write(w, binary.LittleEndian, v) }
	n := uint32(len(pcm))

	io.WriteString(w, "RIFF")
	le(uint32(36 + n))
	io.WriteString(w, "WAVE")
	io.WriteString(w, "fmt ")
	le(uint32(16))             // fmt chunk size.
	le(uint16(1))              // audio format: PCM.
	le(uint16(1))              // channels: mono.
	le(uint32(sampleRate))     // sample rate.
	le(uint32(sampleRate * 2)) // byte rate.
	le(uint16(2))              // block align.
	le(uint16(16))             // bits per sample.
	io.WriteString(w, "data")
	le(n)
	w.Write(pcm)

	return w.Flush()
}

package main

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// buildStream returns a device-style byte stream of framed samples, optionally
// prefixed with junk so the buffer does not start on a frame boundary.
func buildStream(prefix []byte, frames int, sample int16) []byte {
	var b bytes.Buffer
	b.Write(prefix)
	for f := 0; f < frames; f++ {
		b.Write(marker)
		for j := 0; j < frameSamples; j++ {
			binary.Write(&b, binary.LittleEndian, uint16(sample))
		}
	}
	return b.Bytes()
}

// decode runs the same lock+consume logic as main over a full buffer.
func decode(buf []byte) []byte {
	off, ok := findLock(buf)
	if !ok {
		return nil
	}
	buf = buf[off:]
	var pcm []byte
	for len(buf) >= frameBytes {
		if !bytes.Equal(buf[0:4], marker) {
			break
		}
		pcm = append(pcm, buf[4:frameBytes]...)
		buf = buf[frameBytes:]
	}
	return pcm
}

// The bug that produced static: silence (0x0000 samples) next to a zero sync
// word forged a byte-misaligned lock. With a non-zero marker and a mid-frame
// start, the lock must land exactly on a marker and decode to clean zeros.
func TestSilenceNoMisalign(t *testing.T) {
	// Odd-length junk so the first marker is at an odd offset in the buffer.
	stream := buildStream([]byte{0x11, 0x22, 0x33}, 6, 0x0000)
	pcm := decode(stream)

	wantSamples := 5 * frameSamples // first frame may be partially before lock.
	if len(pcm) < wantSamples*2 {
		t.Fatalf("decoded %d samples, want >= %d", len(pcm)/2, wantSamples)
	}
	for i := 0; i+1 < len(pcm); i += 2 {
		if v := binary.LittleEndian.Uint16(pcm[i:]); v != 0 {
			t.Fatalf("sample %d = 0x%04x, want 0 (misaligned lock)", i/2, v)
		}
	}
}

// A loud tone must round-trip byte-exact through lock+consume.
func TestTonePreserved(t *testing.T) {
	var want []int16
	var b bytes.Buffer
	for f := 0; f < 4; f++ {
		b.Write(marker)
		for j := 0; j < frameSamples; j++ {
			s := int16((f*frameSamples + j) * 137) // varied, spans sign.
			want = append(want, s)
			binary.Write(&b, binary.LittleEndian, uint16(s))
		}
	}
	pcm := decode(b.Bytes())
	for i, s := range want {
		if i*2+1 >= len(pcm) {
			t.Fatalf("short decode: %d samples, want %d", len(pcm)/2, len(want))
		}
		if got := int16(binary.LittleEndian.Uint16(pcm[i*2:])); got != s {
			t.Fatalf("sample %d: got %d want %d", i, got, s)
		}
	}
}

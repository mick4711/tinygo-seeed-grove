package lis3dh

import (
	"testing"

	"tinygo.org/x/drivers"
)

// mockI2C answers reads with the WHO_AM_I value so Configure succeeds, and
// otherwise does nothing. It never allocates.
type mockI2C struct{}

func (mockI2C) Tx(addr uint16, w, r []byte) error {
	for i := range r {
		r[i] = whoAmIValue
	}
	return nil
}

func newTestDevice(tb testing.TB) Device {
	d := New(mockI2C{})
	if err := d.Configure(Config{}); err != nil {
		tb.Fatal(err)
	}
	return d
}

// The read/update path must not allocate: the transaction buffers live on the
// Device, so a steady-state read loop should stay at 0 allocs/op.
func TestUpdateNoAllocs(t *testing.T) {
	d := newTestDevice(t)
	if n := testing.AllocsPerRun(100, func() {
		if err := d.Update(drivers.Acceleration); err != nil {
			t.Fatal(err)
		}
		d.Acceleration()
	}); n != 0 {
		t.Fatalf("Update+Acceleration allocated %v times/op, want 0", n)
	}
}

func BenchmarkUpdate(b *testing.B) {
	d := newTestDevice(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := d.Update(drivers.Acceleration); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkAcceleration(b *testing.B) {
	d := newTestDevice(b)
	if err := d.Update(drivers.Acceleration); err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d.Acceleration()
	}
}

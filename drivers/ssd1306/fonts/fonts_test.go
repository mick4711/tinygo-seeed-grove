package fonts

import (
	"strings"
	"testing"

	ssd1306 "github.com/soypat/seeed-grove/drivers/ssd1306"
)

func TestFont5x7A(t *testing.T) {
	buf := make([]byte, ssd1306.BufferSize(8, 8))
	if _, err := ssd1306.DrawText(buf, 0, 0, "A", &Font5x7, 8, true); err != nil {
		t.Fatal(err)
	}
	got := string(ssd1306.AppendFormatDisplay(nil, buf, 8, '#', '.'))
	want := strings.Join([]string{
		"..#.....",
		".#.#....",
		"#...#...",
		"#...#...",
		"#####...",
		"#...#...",
		"#...#...",
		"........",
	}, "\n") + "\n"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestFont5x7StringWidth(t *testing.T) {
	if got := Font5x7.StringWidth("64x48"); got != 5*6 {
		t.Errorf("StringWidth(\"64x48\") = %d, want %d", got, 5*6)
	}
	if got := Font5x7.StringWidth(""); got != 0 {
		t.Errorf("StringWidth(\"\") = %d, want 0", got)
	}
}

func BenchmarkDrawText(b *testing.B) {
	const w, h = 64, 48
	buf := make([]byte, ssd1306.BufferSize(w, h))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		ssd1306.DrawText(buf, 1, 3, "Hello!", &Font5x7, w, true) // unaligned y
	}
}

package ssd1306

import "fmt"

func ExampleAppendFormatDisplay() {
	width := 16
	height := 8
	bufsize := BufferSize(width, height)
	buf := make([]byte, bufsize)
	err := DrawRectangle(buf, 1, 2, 2, 5, width, true)
	if err != nil {
		panic(err)
	}
	err = DrawPixel(buf, 6, 2, width, true)
	if err != nil {
		panic(err)
	}
	draw := AppendFormatDisplay(nil, buf, width, '█', '░')
	fmt.Println(string(draw))
	//output:
	// ░░░░░░░░░░░░░░░░
	// ░░░░░░░░░░░░░░░░
	// ░██░░░█░░░░░░░░░
	// ░██░░░░░░░░░░░░░
	// ░██░░░░░░░░░░░░░
	// ░██░░░░░░░░░░░░░
	// ░██░░░░░░░░░░░░░
	// ░░░░░░░░░░░░░░░░
}

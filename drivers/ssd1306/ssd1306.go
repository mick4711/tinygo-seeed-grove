package ssd1306

import (
	"errors"

	grove "github.com/soypat/seeed-grove"
)

// https://wiki.seeedstudio.com/Grove-OLED-Display-0.66-SSD1306_v1.0

// BufferSize returns the required display buffer length: one byte for the
// leading data-mode byte plus one bit per pixel.
func BufferSize(width, height int) int {
	return 1 + (width*height+7)/8
}

type Device struct {
	bus           grove.I2C
	addr          uint16
	width, height int
	cmdbuf        [2]byte
	displaybuf    []byte
	cmdErr        error
}

func (d *Device) canReset() bool {
	return d.addr != 0 || (d.width != 128 && d.height != 64)
}

type Config struct {
	Width   int
	Height  int
	VccMode VccMode
	// User will allocate their own buffer to use.
	Buffer []byte
}

func (d *Device) ConfigureI2C(bus grove.I2C, addr uint16, config Config) error {
	*d = Device{
		bus:        bus,
		addr:       addr,
		width:      config.Width,
		height:     config.Height,
		displaybuf: d.displaybuf, // keep memory if already set.
		cmdErr:     nil,          // unset.
	}
	bufsize := BufferSize(d.width, d.height)
	if len(config.Buffer) < bufsize {
		return errShortBuffer
	}
	d.displaybuf = config.Buffer[:bufsize]

	d.Command(DISPLAYOFF)
	d.CommandTuple(SETDISPLAYCLOCKDIV, 0x80)

	d.CommandTuple(SETMULTIPLEX, uint8(d.height-1))

	d.CommandTuple(SETDISPLAYOFFSET, 0x0)

	d.Command(SETSTARTLINE | 0x0)

	if config.VccMode == EXTERNALVCC {
		d.CommandTuple(CHARGEPUMP, 0x10)
	} else {
		d.CommandTuple(CHARGEPUMP, 0x14)
	}

	d.CommandTuple(MEMORYMODE, 0)

	if (d.width == 128 && d.height == 64) || (d.width == 64 && d.height == 48) { // 128x64 or 64x48
		d.CommandTuple(SETCOMPINS, 0x12)
		if config.VccMode == EXTERNALVCC {
			d.CommandTuple(SETCONTRAST, 0x9F)
		} else {
			d.CommandTuple(SETCONTRAST, 0xCF)
		}
	} else if d.width == 128 && d.height == 32 { // 128x32
		d.CommandTuple(SETCOMPINS, 0x02)
		d.CommandTuple(SETCONTRAST, 0x8F)
	} else if d.width == 96 && d.height == 16 { // 96x16
		d.CommandTuple(SETCOMPINS, 0x02)
		if config.VccMode == EXTERNALVCC {
			d.CommandTuple(SETCONTRAST, 0x10)
		} else {
			d.CommandTuple(SETCONTRAST, 0xAF)
		}
	} else {
		// fail silently, it might work
		println("there's no configuration for this display's size")
	}

	if config.VccMode == EXTERNALVCC {
		d.CommandTuple(SETPRECHARGE, 0x22)
	} else {
		d.CommandTuple(SETPRECHARGE, 0xF1)
	}
	d.CommandTuple(SETVCOMDETECT, 0x40)
	d.Command(DISPLAYALLON_RESUME)
	d.Command(NORMALDISPLAY)
	d.Command(DEACTIVATE_SCROLL)
	d.Command(DISPLAYON)
	return d.cmdErr
}

func (d *Device) Command(cmd uint8) error {
	d.cmdbuf[0] = 0 // command mode.
	d.cmdbuf[1] = cmd
	err := d.bus.Tx(d.addr, d.cmdbuf[:2], nil)
	if err != nil && d.cmdErr == nil {
		d.cmdErr = err
	}
	return err
}

func (d *Device) CommandTuple(cmd uint8, value uint8) {
	d.Command(cmd)
	d.Command(value)
}

func (d *Device) SetDisplayBuffer(buf []byte) {
	d.displaybuf = buf
}

func (d *Device) Display() error {
	d.displaybuf[0] = 0x40 // data mode.
	return d.bus.Tx(d.addr, d.displaybuf, nil)
}

func (d *Device) CurrentDisplayBuffer() []byte {
	return d.displaybuf
}

var (
	errOutOfRange  = errors.New("ssd1306: rectangle out of range")
	errShortBuffer = errors.New("ssd1306: buffer too short")
)

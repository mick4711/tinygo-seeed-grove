// Package lis3dh drives an ST LIS3DH 3-axis accelerometer over I2C, as found on
// Seeed's Grove 3-Axis Digital Accelerometer (LIS3DHTR) module.
//
// Datasheet: https://www.st.com/resource/en/datasheet/lis3dh.pdf
//
// Reads are done one register at a time (a combined write-address/read-byte
// transaction each). The LIS3DH multi-byte auto-increment read (sub-address MSB
// set) does not survive the ESP32-C3 I2C driver's write-then-read split, so a
// burst read returns the same register repeatedly (x==y==z). Per-register reads
// avoid that entirely.
package lis3dh

import (
	"errors"
	"time"

	grove "github.com/soypat/seeed-grove"
	"tinygo.org/x/drivers"
)

// I2C addresses. The SA0 pin selects between them; Seeed's Grove module straps
// it high, so use AddressHigh (the default when Config.Address is left zero).
const (
	AddressLow  = 0x18 // SA0 low.
	AddressHigh = 0x19 // SA0 high (Seeed Grove LIS3DHTR).
)

// Registers.
const (
	regWhoAmI  = 0x0F
	regTempCfg = 0x1F
	regCtrl1   = 0x20
	regCtrl4   = 0x23
	regOutXL   = 0x28

	whoAmIValue = 0x33
)

// settleDelay matches the reference driver's delay after rate/range writes; the
// device needs time to stabilize or the first reads come back garbage.
const settleDelay = 100 * time.Millisecond

// Range is the full-scale measurement range (CTRL_REG4 bits 5:4).
type Range uint8

const (
	Range2G  Range = 0b00
	Range4G  Range = 0b01
	Range8G  Range = 0b10
	Range16G Range = 0b11
)

// DataRate is the output data rate (CTRL_REG1 bits 7:4).
type DataRate uint8

const (
	Rate1Hz   DataRate = 0b0001
	Rate10Hz  DataRate = 0b0010
	Rate25Hz  DataRate = 0b0011
	Rate50Hz  DataRate = 0b0100
	Rate100Hz DataRate = 0b0101
	Rate200Hz DataRate = 0b0110
	Rate400Hz DataRate = 0b0111
)

// Config holds optional settings; the zero value is valid (Grove defaults).
type Config struct {
	Address uint16   // 0 → AddressHigh.
	Range   Range    // 0 → Range16G (Seeed default).
	Rate    DataRate // 0 → Rate400Hz (Seeed default).
}

// Device wraps an I2C connection to a LIS3DH.
type Device struct {
	bus      grove.I2C
	addr     uint16
	accRange int32   // raw counts per g, per Seeed's calibration.
	data     [6]byte // last values read by Update.
	wbuf     [2]byte
	rbuf     [2]byte
}

var errWrongChip = errors.New("lis3dh: WHO_AM_I mismatch (not a LIS3DH at this address)")

// New creates a Device on the given bus. The bus must already be configured.
// This only builds the object; call Configure to set up the chip.
func New(bus grove.I2C) Device {
	return Device{bus: bus, addr: AddressHigh}
}

// Configure verifies WHO_AM_I and sets up the chip to match Seeed's reference
// driver: all axes enabled at the given rate, block-data-update, normal (non
// high-resolution) mode. Range/Rate default to the Seeed values (16 g, 400 Hz).
func (d *Device) Configure(cfg Config) error {
	if cfg.Address != 0 {
		d.addr = cfg.Address
	}
	if cfg.Range == 0 {
		cfg.Range = Range16G
	}
	if cfg.Rate == 0 {
		cfg.Rate = Rate400Hz
	}

	if id, err := d.readReg(regWhoAmI); err != nil {
		return err
	} else if id != whoAmIValue {
		return errWrongChip
	}

	// TEMP_CFG: ADC and temperature sensor disabled.
	if err := d.writeReg(regTempCfg, 0x00); err != nil {
		return err
	}
	// CTRL_REG1: [ODR(4) | LPen | Zen Yen Xen]. Normal power, all axes enabled.
	if err := d.writeReg(regCtrl1, byte(cfg.Rate)<<4|0x07); err != nil {
		return err
	}
	time.Sleep(settleDelay)
	// CTRL_REG4: [BDU | BLE | FS(2) | HR | ST(2) | SIM]. BDU on, high-res off.
	if err := d.writeReg(regCtrl4, 0x80|byte(cfg.Range)<<4); err != nil {
		return err
	}
	time.Sleep(settleDelay)

	switch cfg.Range {
	case Range16G:
		d.accRange = 1280
	case Range8G:
		d.accRange = 3968
	case Range4G:
		d.accRange = 7282
	default: // Range2G
		d.accRange = 16000
	}
	return nil
}

// Update reads fresh sensor values into the driver. Only Acceleration is
// supported. Call it before Acceleration to refresh the returned values.
func (d *Device) Update(which drivers.Measurement) error {
	if which&drivers.Acceleration == 0 {
		return nil
	}
	for i := range d.data {
		v, err := d.readReg(regOutXL + byte(i))
		if err != nil {
			return err
		}
		d.data[i] = v
	}
	return nil
}

// Acceleration returns the last values read by Update, each in milli-g
// (1000 = 1 g), using Seeed's per-range calibration.
func (d *Device) Acceleration() (x, y, z int32) {
	rx := int16(uint16(d.data[0]) | uint16(d.data[1])<<8)
	ry := int16(uint16(d.data[2]) | uint16(d.data[3])<<8)
	rz := int16(uint16(d.data[4]) | uint16(d.data[5])<<8)
	return d.toMilliG(rx), d.toMilliG(ry), d.toMilliG(rz)
}

func (d *Device) toMilliG(raw int16) int32 {
	return int32(raw) * 1000 / d.accRange
}

func (d *Device) writeReg(reg, val byte) error {
	d.wbuf[0] = reg
	d.wbuf[1] = val
	return d.bus.Tx(d.addr, d.wbuf[:2], nil)
}

func (d *Device) readReg(reg byte) (byte, error) {
	// A combined write-register/read-byte transaction. The ESP32-C3 needs the
	// register write in the same transaction (a read-only Tx NACKs), but rapid
	// back-to-back combined reads race and return the previous read's byte
	// (values walk across axes). A short settle between reads serializes them.
	time.Sleep(time.Millisecond)
	d.wbuf[0] = reg
	err := d.bus.Tx(d.addr, d.wbuf[:1], d.rbuf[:1])
	return d.rbuf[0], err
}

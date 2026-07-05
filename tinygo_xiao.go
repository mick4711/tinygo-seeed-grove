//go:build xiao_esp32c3 || xiao_esp32s3 || xiao_rp2350 || xiao_rp2040 || xiao_ble || xiao_esp32c6

package grove

import "machine"

// dPin maps a Xiao Dx connector index to its board-specific machine.Pin.
// Index N corresponds to Dx label N (e.g. dPin[0] == machine.D0). This is
// required because machine.Pin(N) is the raw GPIO number on some boards
// (e.g. ESP32-C3 where D0 == GPIO2), not the Dx connector.
var dPin = [...]machine.Pin{
	machine.D0, machine.D1, machine.D2, machine.D3, machine.D4, machine.D5,
	machine.D6, machine.D7, machine.D8, machine.D9, machine.D10,
}

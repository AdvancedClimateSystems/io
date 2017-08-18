// Package microchip implements drivers for a few I2C controlled chips produced
// by Microchip.
package microchip

import (
	"fmt"

	"golang.org/x/exp/io/i2c"
)

// The MCP4725 has a 14 bit wide EEPROM to store configuration bits (2 bits)
// and DAC input data (12 bits)
//
// The MCP4275 also has a 19 bits DAC register. The master can read/write the
// DAC register or EEPROM using i2c interface.
//
// The MCP4725 device address contains four fixed bits (1100 = device code) and
// three address bits (A2, A1, A0). The A2 and A1 bits are hard-wired during
// manufacturing, and the A0 bit is determined by the logic state of AO pin.
//
// The MCP4725 has 2 modes of operation: normal mode and power-down mode. This
// driver only supports normal mode.
//
// The datasheet of the device is here:
// http://ww1.microchip.com/downloads/en/DeviceDoc/22039d.pdf
type MCP4725 struct {
	conn *i2c.Device
	vref float64

	Address int
}

// NewMCP4725 returns a new instance of MCP4725.
func NewMCP4725(conn *i2c.Device, vref float64) (*MCP4725, error) {
	return &MCP4725{
		conn: conn,
		vref: vref,
	}, nil
}

// SetVoltage sets voltage of the only channel of the MCP4725. The channel
// parameter is required in the signature of the function to be conform with
// the dac.DAC interface. Because the MCP4725 has only 1 channel it's only
// allowed value is 1.
func (m MCP4725) SetVoltage(v float64, channel int) error {
	code := v * 4096 / m.vref
	return m.SetInputCode(int(code), channel)
}

// SetInputCode sets voltage of the only channel of the MCP4725. The channel
// parameter is required in the signature of the function to be conform with
// the dac.DAC interface. Because the MCP4725 has only 1 channel it's only
// allowed value is 1.
func (m MCP4725) SetInputCode(code, channel int) error {
	if channel != 1 {
		return fmt.Errorf("channel %d is invalid, MCP4725 has only 1 channel", channel)
	}

	if code < 0 || code >= 4096 {
		return fmt.Errorf("digital input code %d is out of range of 0 <= code < 4096", code)
	}

	out := []byte{byte(code >> byte(8)), byte(code & 0xFF)}

	if err := m.conn.Write(out); err != nil {
		return fmt.Errorf("failed to write output code %d: %v", code, err)
	}

	return nil
}

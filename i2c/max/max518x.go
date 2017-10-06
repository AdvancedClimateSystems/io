// Package max contains drivers for IC's produced by Maxim Integrated.
//
// MAX581x
//
// The implemententation for the MAX5813, MAX5814 and MAX5815 only implement
// the REF and CODEn_LOADn commands. The commands CODEn, LOADn, CODEn_LOAD_ALL,
// POWER, SW_CLEAR, SW_RESET, CONFIG, CODE_ALL, LOAD_ALL and CODE_ALL,
// CODE_ALL_LOAD_ALL are not implemented.
package max

import (
	"fmt"
	"math"

	"golang.org/x/exp/io/i2c"
)

const (
	// CODEn_LOADn simultaneously writes data to the selected CODE
	// register(s) while updating selected DAC register(s).
	codenLoadn = 0x30
)

// MAX5813 is a 4 channel DAC with a resolution of 8 bits. The datasheet is
// here: https://datasheets.maximintegrated.com/en/ds/MAX5813-MAX5815.pdf
type MAX5813 struct {
	max581x
}

// NewMAX5813 returns a new instance of MAX5813.
func NewMAX5813(conn *i2c.Device, vref float64) (*MAX5813, error) {
	m := &MAX5813{
		max581x{
			conn:       conn,
			resolution: 8,
		},
	}

	if err := m.SetVref(vref); err != nil {
		return nil, err
	}

	return m, nil
}

// MAX5814 is a 4 channel DAC with a resolution of 10 bits. The datasheet is
// here: https://datasheets.maximintegrated.com/en/ds/MAX5813-MAX5815.pdf
type MAX5814 struct {
	max581x
}

// NewMAX5814 returns a new instance of MAX5814.
func NewMAX5814(conn *i2c.Device, vref float64) (*MAX5814, error) {
	m := &MAX5814{
		max581x{
			conn:       conn,
			resolution: 10,
		},
	}

	if err := m.SetVref(vref); err != nil {
		return nil, err
	}

	return m, nil
}

// MAX5815 is a 4 channel DAC with a resolution of 12 bits. The datasheet is
// here: https://datasheets.maximintegrated.com/en/ds/MAX5813-MAX5815.pdf
type MAX5815 struct {
	max581x
}

// NewMAX5815 returns a new instance of MAX5814.
func NewMAX5815(conn *i2c.Device, vref float64) (*MAX5815, error) {
	m := &MAX5815{
		max581x{
			conn:       conn,
			resolution: 12,
		},
	}

	if err := m.SetVref(vref); err != nil {
		return nil, err
	}

	return m, nil
}

type max581x struct {
	conn       *i2c.Device
	vref       float64
	resolution int
}

// SetVoltage set output voltage of channel. Using the Vref the input code is
// calculated and then SetInputCode is called.
func (m max581x) SetVoltage(v float64, channel int) error {
	code := v * (math.Pow(2, float64(m.resolution)) - 1) / m.vref
	return m.SetInputCode(int(code), channel)
}

// SetInputCode writes the digital input code to the DAC using the CODEn_LOADn
// command.
func (m max581x) SetInputCode(code, channel int) error {
	if channel < 0 || channel > 3 {
		return fmt.Errorf("%d is not a valid channel", channel)
	}

	max := int(math.Pow(2, float64(m.resolution)))
	if code < 0 || code >= max {
		return fmt.Errorf("digital input code %d is out of range of 0 <= code < %d", code, int(max))
	}

	// The requests is 3 bytes long. Byte 1 is the command, byte 2 and 3
	// contain the output code.
	cmd := byte(codenLoadn | channel)
	msb := byte((code >> uint(m.resolution-8)) & 0xFF)
	lsb := byte((code << uint(8-(m.resolution-8))) & 0xFF)

	return m.conn.Write([]byte{cmd, msb, lsb})
}

// Vref sets the global reference for all channels. The device can use either
// an external reference or a internel reference, this depends on the wiring
// of the IC. Allowed values for the internel reference are 2.5V, 2.048V and
// 4.096V. If this function is called with one of these value the internel
// reference is set to this value using the REF command. For any other value
// the channels will use the input reference is equal to the
func (m *max581x) SetVref(v float64) error {
	m.vref = v
	cmd := 0x70

	switch v {
	case 2.5:
		cmd = cmd | 5
	case 2.048:
		cmd = cmd | 6
	case 4.096:
		cmd = cmd | 7
	}

	out := []byte{byte(cmd), 0, 0}
	return m.conn.Write(out)
}

// Conn returns the connection to the I2C device.
func (m *max581x) Conn() *i2c.Device {
	return m.conn
}

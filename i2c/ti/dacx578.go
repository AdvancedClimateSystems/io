// Package ti contains drivers for IC's produced by Texas Instruments.
package ti

import (
	"fmt"
	"math"

	"golang.org/x/exp/io/i2c"
)

const (
	// cmd is the command used to write to DAC input register channel n,
	// and update DAC register channel n. See table 6 of the datasheet.
	cmd = 0x30
)

// DAC5578 is a 8 channel DAC with a resolution of 8 bits. The datasheet is
// here: http://www.ti.com/lit/ds/symlink/dac5578.pdf
type DAC5578 struct {
	dacx578
}

// NewDAC5578 returns a new instance of DAC5578.
func NewDAC5578(conn *i2c.Device, vref float64) *DAC5578 {
	m := &DAC5578{
		dacx578: dacx578{
			conn:       conn,
			resolution: 8,
			vref:       vref,
		},
	}
	return m
}

// DAC6578 is a 8 channel DAC with a resolution of 10 bits. The datasheet is
// here: http://www.ti.com/lit/ds/symlink/dac6578.pdf
type DAC6578 struct {
	dacx578
}

// NewDAC6578 returns a new instance of DAC5578.
func NewDAC6578(conn *i2c.Device, vref float64) *DAC5578 {
	m := &DAC5578{
		dacx578: dacx578{
			conn:       conn,
			resolution: 10,
			vref:       vref,
		},
	}
	return m
}

// DAC7578 is a 8 channel DAC with a resolution of 10 bits. The datasheet is
// here: http://www.ti.com/lit/ds/symlink/dac7578.pdf
type DAC7578 struct {
	dacx578
}

// NewDAC7578 returns a new instance of DAC5578.
func NewDAC7578(conn *i2c.Device, vref float64) *DAC5578 {
	m := &DAC5578{
		dacx578: dacx578{
			conn:       conn,
			resolution: 12,
			vref:       vref,
		},
	}
	return m
}

type dacx578 struct {
	conn       *i2c.Device
	resolution int
	vref       float64
}

// SetVoltage set output voltage of channel. Using the Vref the input code is
// calculated and then SetInputCode is called.
func (d *dacx578) SetVoltage(v float64, channel int) error {
	code := v * ((math.Pow(2, float64(d.resolution)) - 1) / d.vref)
	return d.SetInputCode(int(code), channel)
}

// SetInputCode writes the digital input code to the DAC
func (d *dacx578) SetInputCode(code, channel int) error {
	if channel < 0 || channel > 7 {
		return fmt.Errorf("%d is not a valid channel", channel)
	}

	max := int(math.Pow(2, float64(d.resolution)))
	if code < 0 || code >= max {
		return fmt.Errorf("digital input code %d is out of range of 0 <= code < %d ", code, max)
	}

	// The requests is 3 bytes long. Byte 1 is the command, byte 2 and 3
	// contain the output code.
	cmdAccess := byte(cmd | channel)
	msb := byte((code >> uint(d.resolution-8)) & 0xFF)
	lsb := byte((code << uint(8-(d.resolution-8))) & 0xFF)

	return d.conn.Write([]byte{cmdAccess, msb, lsb})
}

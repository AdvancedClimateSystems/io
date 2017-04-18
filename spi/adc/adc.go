// Package ADC implements a few Analog Digital Converters (ADC). Communication
// with the ADC is done using the Serial Peripheral Interface (SPI) and it
// relies on https://godoc.org/golang.org/x/exp/io/spi package.
package adc

import (
	"fmt"

	"golang.org/x/exp/io/spi"
)

// InputType defines how an ADC samples the input signal. A single-ended input
// samples its input in the range from the ground (0V) to Vref, or the refence
// input. A 10-bits ADC with a reference input of 5V has a precision of (5 -
// 0) / 1024 = 0.0049V = 4.9mV on single-ended inputs.
//
// A (pseudo-)differential output input samples its input between voltage of a
// second pin and Vref, allowing measurements with higher precision. Assume a
// voltage on that second pin is 3V and Vref is 5V. That gives a precision of
// (5 - 3) / 1024 = 0.0020V = 2.0mV for measurements between 3V and 5V. Of
// course, values between 0V and 3V cannot be measured in this case.
type InputType int

const (
	// SingleEnded configures the inputs of an ADC as single-ended.
	SingleEnded InputType = 0
	// PseudoDifferential configures the inputs of an ADC as pseudo-differential.
	PseudoDifferential InputType = 1
)

// ADC is the interface that wraps a Read method.
//
// Read returns the voltage of a channel of the ADC.
type ADC interface {
	Read(channel int) (float64, error)
}

// MCP3008 is 10-bits ADC with 8 single-ended or 4 pseudo-differential inputs.
type MCP3008 struct {
	Conn *spi.Device

	// Vref is the voltage on the reference input of the ADC.
	Vref float64

	InputType InputType
}

// Read returns the voltage of a channel.
func (m MCP3008) Read(channel int) (float64, error) {
	var cmd int

	// The first bit after the start bit will determine if the conversion
	// is done using single-ended or differential input mode. 0 means
	// differential, 1 means single-ended.
	if m.InputType == SingleEnded {
		cmd = 1
	}
	// The bit is then shifted 3 times and the number is incremented with
	// a 3 bits channel.
	cmd = cmd << 3
	cmd += channel

	// The result is shifted 4 times so the high nibble of the byte
	// contains 4 bits of data.
	//
	// 1 1 1 1 x x x x
	// | | | | ------- 4 empty bits.
	// | ------------- 3 bits selecting a channel
	// --------------- The bit defining single-ended or pseudo-differential input mode.
	cmd = cmd << 4

	// The first byte contains a start bit, the second byte contains the
	// actual data and the third byte is another empty byte.
	out := []byte{1, byte(cmd), 0}

	// For every byte send the SPI master reads a byte. Because we send 3
	// bytes we read 3 bytes.
	in := make([]byte, 3)

	if err := m.Conn.Tx(out, in); err != nil {
		return 0, fmt.Errorf("failed to read channel %d: %v", channel, err)
	}

	// The 10-bits measurement are at the end of the 3 byte response.
	//
	// 11111111 11111010 10110111
	//                ^^ ^^^^^^^^
	// To get the base10 value of the channel the second byte is masked
	// with 3:
	//
	//	    11111010
	//          00000011
	//          -------- &
	//          00000010
	//
	// The byte is shifted 8 bits and the last byte is added:
	// 00000010 00000000
	//          10110111
	//          -------- +
	// 00000010 10110111
	//
	// 00000010 10110111 is 696 in base10.
	output := float64(int(in[1]&3)<<8 + int(in[2]))

	return (m.Vref / 1024) * output, nil
}

// Package microchip implements drivers for a few SPI controlled chips produced
// by Microchip.
package microchip

import (
	"fmt"

	"github.com/advancedclimatesystems/io/adc"
	"golang.org/x/exp/io/spi"
)

// MCP3004 is 10-bits ADC with 4 single-ended or 2 pseudo-differential inputs.
// Datasheet: http://ww1.microchip.com/downloads/en/DeviceDoc/21295C.pdf
type MCP3004 struct {
	Conn *spi.Device

	// Vref is the voltage on the reference input of the ADC.
	Vref float64

	InputType adc.InputType
}

// OutputCode queries the channel and returns its digital output code.
func (m MCP3004) OutputCode(channel int) (int, error) {
	if channel < 0 || channel > 3 {
		return 0, fmt.Errorf("channel %d is invalid, ADC has only 4 channels", channel)
	}

	code, err := read10(m.Conn, channel, m.InputType)
	if err != nil {
		return 0, err
	}

	return code, nil
}

// Voltage returns the voltage of a channel.
func (m MCP3004) Voltage(channel int) (float64, error) {
	code, err := m.OutputCode(channel)
	if err != nil {
		return 0, err
	}

	return (m.Vref / 1024) * float64(code), nil
}

// MCP3008 is 10-bits ADC with 8 single-ended or 4 pseudo-differential inputs.
// Datasheet: http://ww1.microchip.com/downloads/en/DeviceDoc/21295C.pdf
type MCP3008 struct {
	Conn *spi.Device

	// Vref is the voltage on the reference input of the ADC.
	Vref float64

	InputType adc.InputType
}

// OutputCode queries the channel and returns its digital output code.
func (m MCP3008) OutputCode(channel int) (int, error) {
	if channel < 0 || channel > 7 {
		return 0, fmt.Errorf("channel %d is invalid, ADC has only 7 channels", channel)
	}

	code, err := read10(m.Conn, channel, m.InputType)
	if err != nil {
		return 0, err
	}

	return code, nil
}

// Voltage returns the voltage of a channel.
func (m MCP3008) Voltage(channel int) (float64, error) {
	code, err := m.OutputCode(channel)
	if err != nil {
		return 0, err
	}

	return (m.Vref / 1024) * float64(code), nil
}

// read10 reads a 10 bits value from an channel of an ADC.
func read10(conn *spi.Device, channel int, inputType adc.InputType) (int, error) {
	var cmd int

	// The first bit after the start bit will determine if the conversion
	// is done using single-ended or differential input mode. 0 means
	// differential, 1 means single-ended.
	if inputType == adc.SingleEnded {
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

	if err := conn.Tx(out, in); err != nil {
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
	return int(in[1]&3)<<8 + int(in[2]), nil
}

// MCP3204 is 12-bits ADC with 4 single-ended or 2 pseudo-differential inputs.
// Datasheet: http://ww1.microchip.com/downloads/en/DeviceDoc/21298e.pdf
type MCP3204 struct {
	Conn *spi.Device

	// Vref is the voltage on the reference input of the ADC.
	Vref float64

	InputType adc.InputType
}

// OutputCode queries the channel and returns its digital output code.
func (m MCP3204) OutputCode(channel int) (int, error) {
	if channel < 0 || channel > 3 {
		return 0, fmt.Errorf("channel %d is invalid, ADC has only 4 channels", channel)
	}

	code, err := read12(m.Conn, channel, m.InputType)
	if err != nil {
		return 0, err
	}

	return code, nil
}

// Voltage returns the voltage of a channel.
func (m MCP3204) Voltage(channel int) (float64, error) {
	code, err := m.OutputCode(channel)
	if err != nil {
		return 0, err
	}

	return (m.Vref / 4096) * float64(code), nil
}

// MCP3208 is 12-bits ADC with 8 single-ended or 4 pseudo-differential inputs.
// Datasheet: http://ww1.microchip.com/downloads/en/DeviceDoc/21298e.pdf
type MCP3208 struct {
	Conn *spi.Device

	// Vref is the voltage on the reference input of the ADC.
	Vref float64

	InputType adc.InputType
}

// OutputCode queries the channel and returns its digital output code.
func (m MCP3208) OutputCode(channel int) (int, error) {
	if channel < 0 || channel > 7 {
		return 0, fmt.Errorf("channel %d is invalid, ADC has only 7 channels", channel)
	}

	code, err := read12(m.Conn, channel, m.InputType)
	if err != nil {
		return 0, err
	}

	return code, nil
}

// Voltage returns the voltage of a channel.
func (m MCP3208) Voltage(channel int) (float64, error) {
	code, err := m.OutputCode(channel)
	if err != nil {
		return 0, err
	}

	return (m.Vref / 4096) * float64(code), nil
}

// read12 reads a 12 bits value from an channel of an ADC.
func read12(conn *spi.Device, channel int, inputType adc.InputType) (int, error) {
	// The start bit.
	cmd := 1
	cmd = cmd << 1

	// The first bit after the start bit will determine if the conversion
	// is done using single-ended or differential input mode. 0 means
	// differential, 1 means single-ended.
	if inputType == adc.SingleEnded {
		cmd = 1
	}
	// The bit is then shifted 3 times and the number is incremented with
	// a 3 bits channel.
	cmd = cmd << 3
	cmd += channel

	// The result is shifted 6 times.
	//
	// x x x x x 1 1 1   1 1 x x x x x x
	//           | | |   | | ------- 3 bits for selecting channel
	//	     | |---------------- 1 bit defining single-ended or pseudo-differential input mode
	//	     |------------------ 1 start bit
	cmd = cmd << 6

	// The data is is in the first 2 bytes, the third byte is an empty byte.
	out := []byte{byte(cmd >> 8), byte(cmd & 0xFF), 0}

	// For every byte send the SPI master reads a byte. Because we send 3
	// bytes we read 3 bytes.
	in := make([]byte, 3)

	if err := conn.Tx(out, in); err != nil {
		return 0, fmt.Errorf("failed to read channel %d: %v", channel, err)
	}

	// The 12-bits measurement is at the end of the 3 byte response.
	//
	// 11111111 11101100 10110111
	//              ^^^^ ^^^^^^^^
	// To get the base10 value of the channel the second byte is masked
	// with 15:
	//
	//	    11101100
	//          00001111
	//          -------- &
	//          00001100
	//
	// The byte is shifted 8 bits and the last byte is added:
	// 00001100 00000000
	//          10110111
	//          -------- +
	// 00001100 10110111
	//
	// 00001100 10110111 is 3255 in base10.
	return int(in[1]&0xF)<<8 + int(in[2]), nil
}

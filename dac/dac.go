// Package dac defines the interface for Digital Analog Converters.
package dac

// DAC is the interface to set the output voltage(s) of a Digital Analog
// Converter.
type DAC interface {
	// SetVoltage sets output voltage of a channel.
	SetVoltage(voltage float64, channel int) error

	// SetInputCode sets output voltage using an number that is between
	// and including 0 - (max resolution of DAC - 1).
	SetInputCode(code, channel int) error
}

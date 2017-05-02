package adc

// InputType defines how an ADC samples the input signal. A single-ended input
// samples its input in the range from the ground (0V) to Vref, that is  the
// reference input. A 10-bits ADC with a reference input of 5V has a precision
// of (5 - 0) / 1024 = 0.0049V = 4.9mV on single-ended inputs.
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

// ADC is the interface that wraps an OutputCode and Voltage method. The first
// returns the digital output code of a channel. The latter returns the voltage
// of a channel.
type ADC interface {
	// OutputCode queries the channel and returns its digital output code.
	OutputCode(channel int) (int, error)
	// Voltage queries the channel of an ADC and returns its voltage.
	Voltage(channel int) (float64, error)
}

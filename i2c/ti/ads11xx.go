// Package ti contains drivers for IC's produces by Texas Instruments.
package ti

import (
	"fmt"
	"math"

	"golang.org/x/exp/io/i2c"
)

type dataRate struct {
	// sps is the data rate samples per second.
	sps int

	bitMask int

	// size is the amount of bits that represent the output code.
	size uint
}

type ads11xx struct {
	Conn *i2c.Device
	Vref float64

	dataRate dataRate
	pga      int

	// dataRates is a map that holds all valid values for data rate.
	dataRates []dataRate
}

func newADS11xx(conn *i2c.Device, vref float64, dataRate, pga int, dataRates []dataRate) (ads11xx, error) {
	a := ads11xx{
		Conn:      conn,
		Vref:      vref,
		dataRates: dataRates,
	}

	if err := a.setDataRate(dataRate); err != nil {
		return a, err
	}

	if err := a.SetPGA(pga); err != nil {
		return a, err
	}

	return a, nil
}

// Voltage queries the channel of an ADC and returns its voltage.
func (a ads11xx) Voltage(channel int) (float64, error) {
	code, err := a.OutputCode(channel)
	if err != nil {
		return 0, err
	}

	max := math.Pow(2, float64(a.dataRate.size))
	return ((a.Vref / max) * float64(code) / float64(a.pga)), nil
}

// OutputCode queries the channel and returns its digital output code. The
// maximum code depends on the selected data rate.  The higher the data rate,
// the lower the number of bits used.
func (a ads11xx) OutputCode(channel int) (int, error) {
	if channel != 1 {
		return 0, fmt.Errorf("channel %d is invalid, ADC has only 1 channel", channel)
	}

	in := make([]byte, 2)
	if err := a.Conn.Read(in); err != nil {
		return 0, fmt.Errorf("failed to read output code: %v", err)
	}

	msb := in[0] & byte(math.Pow(2, float64(a.dataRate.size-8))-1)
	v := (int(msb) << 8) + int(in[1])

	return v, nil
}

// PGA reads the config register of the ADC and returns the current PGA.
func (a *ads11xx) PGA() (int, error) {
	data, err := a.config()
	if err != nil {
		return 0, err
	}

	return int(math.Pow(2, float64(data&0x3))), nil
}

// SetPGA writes the value for the Programmable Gain Amplifier to the ADC.
// Valid values are 1, 2, 4 and 8.
func (a *ads11xx) SetPGA(v int) error {
	if v == 1 || v == 2 || v == 4 || v == 8 {
		a.pga = int(math.Log2(float64(v)))
		return a.setConfig()
	}

	return fmt.Errorf("PGA of %d is invalid, choose 1, 2, 4 or 8", v)
}

// DataRate reads the config register of the ADC returns the current value of
// the data rate.
func (a *ads11xx) DataRate() (int, error) {
	data, err := a.config()
	if err != nil {
		return 0, err
	}

	v := int(data&0xc) >> 2

	for _, d := range a.dataRates {
		if v == d.bitMask {
			return d.sps, nil
		}
	}

	return 0, fmt.Errorf("failed to understand data %x rate value read from config register", v)
}

// SetDataRate writes the value for the data rate to the ADC.
func (a *ads11xx) SetDataRate(r int) error {
	if err := a.setDataRate(r); err != nil {
		return err
	}

	return a.setConfig()
}

func (a *ads11xx) setDataRate(sps int) error {
	var ok bool
	var dataRate dataRate

	for _, rate := range a.dataRates {
		if rate.sps == sps {
			dataRate = rate
			ok = true
			break
		}
	}

	if !ok {
		var rates []int
		for _, rate := range a.dataRates {
			rates = append(rates, rate.sps)
		}
		return fmt.Errorf("%d is an invalid value for data rate, use on of %v", sps, rates)
	}
	a.dataRate = dataRate

	return nil
}

// config reads the config register of the ADC and returns its value.
func (a *ads11xx) config() (byte, error) {
	in := make([]byte, 3)
	if err := a.Conn.Read(in); err != nil {
		return 0, err
	}

	// The first 2 bytes contain the output code, those are ignored. The
	// third bytes contains value of config register.
	return in[2], nil
}

// setConfig writes the settings for the data rate and PGA to the config
// register.
func (a *ads11xx) setConfig() error {
	out := []byte{byte(a.dataRate.bitMask<<2 | a.pga)}
	return a.Conn.Write(out)
}

// ADS1100 is a 16-bit ADC. It's PGA can be set to 1, 2, 4 or 8. Allowed
// values for the data rate are 8, 16, 32 or 128 SPS.
type ADS1100 struct {
	ads11xx
}

// NewADS1100 returns an ADS1100.
func NewADS1100(conn *i2c.Device, vref float64, rate, pga int) (*ADS1100, error) {
	dataRates := []dataRate{
		dataRate{sps: 128, bitMask: 0x0, size: 12},
		dataRate{sps: 32, bitMask: 0x1, size: 14},
		dataRate{sps: 16, bitMask: 0x2, size: 15},
		dataRate{sps: 8, bitMask: 0x3, size: 16},
	}

	inner, err := newADS11xx(conn, vref, rate, pga, dataRates)
	if err != nil {
		return nil, fmt.Errorf("failed to create ADS1100: %v", err)
	}
	return &ADS1100{
		inner,
	}, nil
}

// ADS1110 is a 16-bits ADC. It's PGA can be set to 1, 2, 4 or 8. Allowed
// values are 15, 30, 60 or 240 SPS. The ADS1110 always uses an internal
// voltage reference of 2.048V.
type ADS1110 struct {
	ads11xx
}

// NewADS1110 returns an ADS1110.
func NewADS1110(conn *i2c.Device, rate, pga int) (*ADS1110, error) {
	dataRates := []dataRate{
		dataRate{sps: 240, bitMask: 0x0, size: 12},
		dataRate{sps: 60, bitMask: 0x1, size: 14},
		dataRate{sps: 30, bitMask: 0x2, size: 15},
		dataRate{sps: 15, bitMask: 0x3, size: 16},
	}

	inner, err := newADS11xx(conn, 2.048, rate, pga, dataRates)

	if err != nil {
		return nil, fmt.Errorf("failed to create ADS1110: %v", err)
	}

	return &ADS1110{
		inner,
	}, nil
}

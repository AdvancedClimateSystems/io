package ti

import (
	"errors"
	"fmt"
	"testing"

	"github.com/advancedclimatesystems/io/dac"
	"github.com/advancedclimatesystems/io/iotest"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/io/i2c"
)

func TestDACinterface(t *testing.T) {
	assert.Implements(t, (*dac.DAC)(nil), new(DAC5578))
	assert.Implements(t, (*dac.DAC)(nil), new(DAC6578))
	assert.Implements(t, (*dac.DAC)(nil), new(DAC7578))
}

func TestNewDACX578(t *testing.T) {
	conn, _ := i2c.Open(iotest.NewI2CDriver(iotest.NewI2CConn()), 0x1)

	dac5578 := NewDAC5578(conn, 3)
	assert.Equal(t, 8, dac5578.resolution)

	dac6578 := NewDAC6578(conn, 3)
	assert.Equal(t, 10, dac6578.resolution)

	dac7578 := NewDAC7578(conn, 3)
	assert.Equal(t, 12, dac7578.resolution)
}

func TestDACX578SetVoltage(t *testing.T) {
	data := make(chan []byte, 2)
	c := iotest.NewI2CConn()
	c.TxFunc(func(w, _ []byte) error {
		data <- w
		return nil
	})

	conn, _ := i2c.Open(iotest.NewI2CDriver(c), 0x1)
	m := dacx578{
		conn: conn,
	}

	var tests = []struct {
		resolution int
		vref       float64
		voltage    float64
		channel    int
		expected   []byte
	}{
		{8, 10, 10, 1, []byte{0x31, 0xff, 0}},
		{8, 10, 5, 1, []byte{0x31, 0x7f, 0}},
		{8, 10, 0, 2, []byte{0x32, 0x0, 0}},
		{8, 5, 5, 2, []byte{0x32, 0xff, 0}},
		{8, 20, 10, 2, []byte{0x32, 0x7f, 0}},
		{10, 10, 10, 2, []byte{0x32, 0xff, 0xc0}},
		{10, 10, 5, 2, []byte{0x32, 0x7f, 0xc0}},
		{10, 10, 0, 2, []byte{0x32, 0x00, 0x00}},
		{12, 10, 10, 3, []byte{0x33, 0xff, 0xf0}},
		{12, 10, 5, 3, []byte{0x33, 0x7f, 0xf0}},
		{12, 10, 0, 4, []byte{0x34, 0x00, 0x00}},
	}

	for _, test := range tests {
		m.resolution = test.resolution
		m.vref = test.vref

		err := m.SetVoltage(test.voltage, test.channel)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, test.expected, <-data)
	}
}

func TestDACX578SetVoltageChannelOutOfRange(t *testing.T) {
	c := iotest.NewI2CConn()
	c.TxFunc(func(w, _ []byte) error {
		return nil
	})

	conn, _ := i2c.Open(iotest.NewI2CDriver(c), 0x1)
	m := dacx578{
		conn:       conn,
		resolution: 10,
		vref:       10,
	}

	var tests = []struct {
		channel  int
		expected error
	}{
		{0, nil},
		{7, nil},
		{8, errors.New("8 is not a valid channel")},
		{-1, errors.New("-1 is not a valid channel")},
	}

	for _, test := range tests {
		err := m.SetVoltage(5, test.channel)
		assert.Equal(t, test.expected, err)
	}
}

func TestDACX578SetVoltageOutRange(t *testing.T) {
	c := iotest.NewI2CConn()
	c.TxFunc(func(w, _ []byte) error {
		return nil
	})

	conn, _ := i2c.Open(iotest.NewI2CDriver(c), 0x1)
	m := dacx578{
		conn: conn,
	}

	var tests = []struct {
		vref       float64
		voltage    float64
		resolution int
		expected   error
	}{
		{10, 10, 8, nil},
		{10, 10, 10, nil},
		{10, 10, 12, nil},
		{10, 11, 8, errors.New("digital input code 280 is out of range of 0 <= code < 256 ")},
		{10, 11, 10, errors.New("digital input code 1125 is out of range of 0 <= code < 1024 ")},
		{10, 11, 12, errors.New("digital input code 4504 is out of range of 0 <= code < 4096 ")},
	}

	for _, test := range tests {
		m.resolution = test.resolution
		m.vref = test.vref

		err := m.SetVoltage(test.voltage, 1)
		assert.Equal(t, test.expected, err)
	}
}

func ExampleDAC5578() {
	// We are going to write 5.5 volt to channel 0.
	volts := 5.5
	channel := 0

	dev, err := i2c.Open(&i2c.Devfs{
		Dev: "/dev/i2c-0",
	}, 0x48)

	if err != nil {
		panic(fmt.Sprintf("failed to open device: %v", err))
	}
	defer dev.Close()

	// Create the DAC. The reference voltage is set to 10V.
	dac := NewDAC5578(dev, 10)

	// Write volts to the channel.
	err = dac.SetVoltage(volts, channel)
	if err != nil {
		panic(fmt.Sprintf("failed to set voltage: %v", err))
	}

	// It's also possible to set output of a channel with digital output
	// code. The value must be between 0 and 255.
	if err := dac.SetInputCode(255, channel); err != nil {
		panic(fmt.Sprintf("failed to set voltage using output code: %v", err))
	}
}

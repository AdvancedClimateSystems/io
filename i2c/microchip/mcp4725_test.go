package microchip

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
	assert.Implements(t, (*dac.DAC)(nil), new(MCP4725))
}

func TestMCP4725WithValidVoltages(t *testing.T) {
	data := make(chan []byte, 2)
	c := iotest.NewI2CConn()

	c.TxFunc(func(w, _ []byte) error {
		data <- w
		return nil
	})

	conn, _ := i2c.Open(iotest.NewI2CDriver(c), 0x1)

	var tests = []struct {
		vref     float64
		voltage  float64
		expected []byte
	}{
		{2.7, 1.73, []byte{0xa, 0x40}},
		{2.7, 2.6999, []byte{0x0f, 0xff}},
		{2.7, 0, []byte{0x0, 0x0}},
		{5.5, 1.22, []byte{0x3, 0x8c}},
		{5.5, 0.73, []byte{0x2, 0x1f}},
	}

	for _, test := range tests {
		m, err := NewMCP4725(conn, test.vref)
		assert.Nil(t, err)

		err = m.SetVoltage(test.voltage, 1)
		assert.Equal(t, test.expected, <-data)
		assert.Nil(t, err)
	}
}

func TestMCP4725WithInValidVoltages(t *testing.T) {
	conn, _ := i2c.Open(iotest.NewI2CDriver(iotest.NewI2CConn()), 0x1)
	m, _ := NewMCP4725(conn, 2.7)

	voltages := []float64{-1, 28.1}
	for _, v := range voltages {
		err := m.SetVoltage(v, 1)
		assert.NotNil(t, err)
	}
}

func TestMCP4725WithInvalidChannel(t *testing.T) {
	conn, _ := i2c.Open(iotest.NewI2CDriver(iotest.NewI2CConn()), 0x1)
	m, _ := NewMCP4725(conn, 2.7)

	channels := []int{-1, 0, 2, 28}
	for _, c := range channels {
		err := m.SetVoltage(1, c)
		assert.NotNil(t, err)
	}
}

func TestMCP4725WithFailingConnection(t *testing.T) {
	c := iotest.NewI2CConn()
	c.TxFunc(func(_, _ []byte) error { return errors.New("Is there a officer, problem?") })
	conn, _ := i2c.Open(iotest.NewI2CDriver(c), 0x1)

	m, _ := NewMCP4725(conn, 2.7)
	err := m.SetVoltage(1, 1)

	assert.NotNil(t, err)
}

func ExampleMCP4725() {
	d, err := i2c.Open(&i2c.Devfs{
		Dev: "/dev/i2c-0",
	}, 0x61)

	if err != nil {
		panic(fmt.Sprintf("failed to open device: %v", err))
	}
	defer d.Close()

	// Reference voltage is 2.7V.
	dac, err := NewMCP4725(d, 2.7)

	if err != nil {
		panic(fmt.Sprintf("failed to create MCP4725: %v", err))
	}

	// Set output of channel 1 to 1.3V. The MCP4725 has only 1 channel,
	// select other channels results in an error.
	if err := dac.SetVoltage(3, 1); err != nil {
		panic(fmt.Sprintf("failed to set voltage: %v", err))
	}

	// It's also possible to set output of a channel with digital output
	// code. The value must be in range of 0 till 4096.
	if err := dac.SetInputCode(4095, 1); err != nil {
		panic(fmt.Sprintf("failed to set voltage using output code: %v", err))
	}
}

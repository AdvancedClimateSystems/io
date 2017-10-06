package microchip

import (
	"fmt"
	"testing"

	"github.com/advancedclimatesystems/io/adc"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/io/spi"
	"golang.org/x/exp/io/spi/driver"
)

// testDriver is a mocked driver that implements the driver.Opener interface.
type testDriver struct {
	conn testConn
}

func (d testDriver) Open() (driver.Conn, error) {
	return d.conn, nil
}

// testConn is a mocked connection that implements the spi.Conn interface.
type testConn struct {
	tx func(w, r []byte) error
}

func (c testConn) Configure(k, v int) error { return nil }

func (c testConn) Tx(w, r []byte) error {
	return c.tx(w, r)
}

func (c testConn) Close() error { return nil }

func TestMCP300x(t *testing.T) {
	var tests = []struct {
		resp []byte
		v    float64
	}{
		{[]byte{0, 0}, 0},
		{[]byte{2, 0}, 2.5},
		{[]byte{6, 0}, 2.5},
		{[]byte{255, 255}, 4.9951171875},
	}

	for _, test := range tests {
		c := testConn{
			tx: func(w, r []byte) error {
				assert.Equal(t, []byte{1, 176, 0}, w)

				r[1] = test.resp[0]
				r[2] = test.resp[1]

				return nil
			},
		}

		con, _ := spi.Open(&testDriver{c})
		mcp3004 := MCP3004{
			Conn:      con,
			Vref:      5.0,
			InputType: adc.SingleEnded,
		}

		v, _ := mcp3004.Voltage(3)
		assert.Equal(t, test.v, v)

		mcp3008 := MCP3008{
			Conn:      con,
			Vref:      5.0,
			InputType: adc.SingleEnded,
		}

		v, _ = mcp3008.Voltage(3)
		assert.Equal(t, test.v, v)
	}
}

func TestMCP320x(t *testing.T) {
	var tests = []struct {
		resp []byte
		v    float64
	}{
		{[]byte{0, 0}, 0},
		{[]byte{2, 0}, 0.625},
		{[]byte{6, 0}, 1.875},
		{[]byte{1, 13}, 0.328369140625},
		{[]byte{255, 255}, 4.998779296875},
	}

	for _, test := range tests {
		c := testConn{
			tx: func(w, r []byte) error {
				assert.Equal(t, []byte{4, 192, 0}, w)

				r[1] = test.resp[0]
				r[2] = test.resp[1]

				return nil
			},
		}

		con, _ := spi.Open(&testDriver{c})
		mcp3204 := MCP3204{
			Conn:      con,
			Vref:      5.0,
			InputType: adc.PseudoDifferential,
		}

		v, _ := mcp3204.Voltage(3)
		assert.Equal(t, test.v, v)

		mcp3208 := MCP3208{
			Conn:      con,
			Vref:      5.0,
			InputType: adc.PseudoDifferential,
		}

		v, _ = mcp3208.Voltage(3)
		assert.Equal(t, test.v, v)

	}
}

// TestWithInvalidChannels calls adc.OutputCode with a channel that isn't in
// the range of the ADC.
func TestMCP3x0xWithInvalidChannels(t *testing.T) {
	var tests = []struct {
		adc     adc.ADC
		channel int
	}{
		{MCP3004{}, -1},
		{MCP3004{}, 4},
		{MCP3008{}, -1},
		{MCP3008{}, 8},
		{MCP3204{}, -1},
		{MCP3204{}, 4},
		{MCP3208{}, -1},
		{MCP3208{}, 8},
	}

	for _, test := range tests {
		_, err := test.adc.OutputCode(test.channel)
		assert.NotNil(t, err)
	}
}

// TestMCP3x0xWithFailingConnection test if all ADC's return errors when the
// connection fails.
func TestMCP3x0xWithFailingConnection(t *testing.T) {
	c := testConn{
		tx: func(w, r []byte) error {
			return fmt.Errorf("some error occured")
		},
	}
	con, _ := spi.Open(&testDriver{c})

	adcs := []adc.ADC{
		MCP3004{
			Conn: con,
		},
		MCP3008{
			Conn: con,
		},
		MCP3204{
			Conn: con,
		},
		MCP3208{
			Conn: con,
		},
	}

	for _, adc := range adcs {
		_, err := adc.Voltage(3)
		assert.NotNil(t, err)
	}
}

func ExampleMCP3008() {
	conn, err := spi.Open(&spi.Devfs{
		Dev:      "/dev/spidev32766.0",
		Mode:     spi.Mode0,
		MaxSpeed: 3600000,
	})

	if err != nil {
		panic(fmt.Sprintf("failed to open SPI device: %s", err))
	}

	defer conn.Close()

	a := MCP3008{
		Conn: conn,
		Vref: 5.0,

		// Optional, default value is SingleEnded.
		InputType: adc.SingleEnded,
	}

	// Voltage the voltage on channel 3.
	v, err := a.Voltage(3)
	if err != nil {
		panic(fmt.Sprintf("failed to read channel 3 of MCP3008: %s", err))
	}

	fmt.Printf("read %f Volts from channel 3", v)
}

package adc

import (
	"fmt"
	"testing"

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

func TestMCP3008(t *testing.T) {
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
				assert.Equal(t, []byte{1, 240, 0}, w)

				r[1] = test.resp[0]
				r[2] = test.resp[1]

				return nil
			},
		}

		con, _ := spi.Open(&testDriver{c})
		m := MCP3008{
			Conn:      con,
			Vref:      5.0,
			InputType: SingleEnded,
		}

		v, _ := m.Read(7)
		assert.Equal(t, test.v, v)
	}
}

func ExampleMCP3008() {
	conn, err := spi.Open(&spi.Devfs{
		Dev:      "/dev/spidev32766.0",
		Mode:     spi.Mode0,
		MaxSpeed: 5000000,
	})

	if err != nil {
		panic(fmt.Sprintf("failed to open SPI device: %s", err))
	}

	defer conn.Close()

	a := MCP3008{
		Conn: conn,
		Vref: 5.0,
	}

	// Read the voltage on channel 3.
	v, err := a.Read(3)
	if err != nil {
		panic(fmt.Sprintf("failed to read channel 3 of MCP3008: %s", err))
	}
	fmt.Printf("read %f Volts from channel 3", v)
}

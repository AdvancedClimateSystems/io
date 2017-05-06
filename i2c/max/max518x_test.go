package max

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/io/i2c"
	"golang.org/x/exp/io/i2c/driver"
)

type testDriver struct {
	conn testConn
}

func (d testDriver) Open(_ int, _ bool) (driver.Conn, error) {
	return d.conn, nil
}

// testConn is a mocked connection that implements the spi.Conn interface.
type testConn struct {
	tx func(w, r []byte) error
}

func (c testConn) Tx(w, r []byte) error {
	return c.tx(w, r)
}

func (c testConn) Close() error { return nil }

func TestMAX581xSetVref(t *testing.T) {
	data := make(chan []byte, 2)
	c := testConn{
		tx: func(w, r []byte) error {
			data <- w
			return nil
		},
	}

	conn, _ := i2c.Open(&testDriver{c}, 0x1)

	m := max581x{
		conn:       conn,
		resolution: 8,
	}

	var tests = []struct {
		vref     float64
		expected []byte
	}{
		{2.5, []byte{0x75, 0, 0}},
		{2.048, []byte{0x76, 0, 0}},
		{4.096, []byte{0x77, 0, 0}},
	}

	for _, test := range tests {
		m.SetVref(test.vref)
		assert.Equal(t, test.expected, <-data)
	}

}

func TestMAX581xSetVoltage(t *testing.T) {
	data := make(chan []byte, 2)
	c := testConn{
		tx: func(w, r []byte) error {
			data <- w
			return nil
		},
	}

	conn, _ := i2c.Open(&testDriver{c}, 0x1)

	m := max581x{
		conn: conn,
	}

	var tests = []struct {
		resolution int
		vref       float64
		voltage    float64
		channel    int
		expected   []byte
	}{
		{8, 2.5, 2.5, 1, []byte{0x30, 0xff, 0}},
		{8, 2.5, 0, 2, []byte{0x31, 0x0, 0}},
		{8, 5, 2.5, 2, []byte{0x31, 0x7f, 0}},
		{10, 2.5, 2.5, 2, []byte{0x31, 0xff, 0xc0}},
		{12, 2.5, 2.5, 3, []byte{0x32, 0xff, 0xf0}},
		{12, 10, 2, 3, []byte{0x32, 0x33, 0x30}},
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

// Package iotest contains some test helpers for code relying on
// golang.org/x/exp/io/i2c.
//
//  func TestMCP4725(t *testing.T)  {
//	data := make(chan []byte, 2)
//	c := iotest.NewI2CConn()
//
//	// Set the TxFunc.
//	c.TxFunc(func(w, _ []byte) error {
//		data <- w
//		return nil
//	})
//
//	conn, _ := i2c.Open(iotest.NewI2CDriver(c), 0x1)
//	dac, _  := microchip.NewMCP4725(conn, 5.5)
//
//         // Under the hood SetInputCode calls c.Tx which in turn calls TxFunc defined earlier.
//	dac.SetInputCode(0x539, 1)
//
//	assert.Equal(t, []byte{0x5, 0x39}, <-data)
// }
package iotest

import (
	"golang.org/x/exp/io/i2c/driver"
)

// I2CDriver implements the i2c.Device interface.
type I2CDriver struct {
	conn I2CConn
}

// NewI2CDriver creates a new I2CDriver.
func NewI2CDriver(c I2CConn) *I2CDriver {
	return &I2CDriver{conn: c}
}

// Open returns a type that implements the driver.Conn interface.
func (d I2CDriver) Open(_ int, _ bool) (driver.Conn, error) {
	return d.conn, nil
}

// I2CConn implements the driver.Conn interface.
type I2CConn struct {
	tx    func(w, r []byte) error
	close func() error
}

// NewI2CConn creates a new I2CConn.
func NewI2CConn() I2CConn {
	c := I2CConn{}

	c.TxFunc(func(_, _ []byte) error {
		return nil
	})

	c.CloseFunc(func() error {
		return nil
	})

	return c
}

// Tx calls the TxFunc.
func (c I2CConn) Tx(w, r []byte) error {
	return c.tx(w, r)
}

// TxFunc sets TxFunc which is called when Tx is called.
func (c *I2CConn) TxFunc(f func(w, r []byte) error) {
	c.tx = f
}

// Close calls the CloseFunc.
func (c I2CConn) Close() error {
	return c.close()
}

// CloseFunc sets the CloseFunc. CloseFunc is called when Close is called.
func (c *I2CConn) CloseFunc(f func() error) {
	c.close = f
}

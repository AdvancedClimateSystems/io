package ti

import (
	"fmt"
	"math"
	"testing"

	"github.com/advancedclimatesystems/io/iotest"
	"github.com/stretchr/testify/assert"

	"golang.org/x/exp/io/i2c"
)

// TestADS11xxPGA tests if configuring the devices works as expected.
func TestADS11xxPGA(t *testing.T) {
	data := make(chan []byte, 1)
	c := iotest.NewI2CConn()
	c.TxFunc(func(w, _ []byte) error {
		data <- w
		return nil
	})

	conn, err := i2c.Open(iotest.NewI2CDriver(c), 0x1)
	a, err := NewADS1100(conn, 5.0, 128, 2)

	// Test if config register is written correctly
	assert.Equal(t, []byte{0x1}, <-data)

	// Test with invalig value for PGA.
	assert.NotNil(t, a.SetPGA(18))

	// Test with valid value for PGA.
	assert.Nil(t, a.SetPGA(8))
	assert.Equal(t, []byte{0x3}, <-data)

	c.TxFunc(func(_, r []byte) error {
		copy(r, []byte{0x0, 0x0, 0x3})
		return nil
	})

	pga, err := a.PGA()
	assert.Nil(t, err)
	assert.Equal(t, 8, pga)
}

func TestADS11xxDataRate(t *testing.T) {
	data := make(chan []byte, 1)
	c := iotest.NewI2CConn()
	c.TxFunc(func(w, _ []byte) error {
		data <- w
		return nil
	})

	conn, _ := i2c.Open(iotest.NewI2CDriver(c), 0x1)
	a, _ := NewADS1100(conn, 5.0, 128, 2)

	assert.Equal(t, []byte{0x1}, <-data)

	assert.Nil(t, a.SetDataRate(32))
	assert.Equal(t, []byte{0x5}, <-data)

	// Test with invalid value for data rate.
	assert.NotNil(t, a.SetDataRate(18))

	c.TxFunc(func(_, r []byte) error {
		copy(r, []byte{0x0, 0x0, 0x5})
		return nil
	})

	d, err := a.DataRate()
	assert.Nil(t, err)
	assert.Equal(t, 32, d)
}

func TestADS1100Voltage(t *testing.T) {
	data := make(chan []byte, 1)
	c := iotest.NewI2CConn()

	conn, _ := i2c.Open(iotest.NewI2CDriver(c), 0x1)
	ads, _ := NewADS1100(conn, 5.0, 128, 2)
	c.TxFunc(func(w, r []byte) error {
		copy(r, <-data)
		return nil
	})

	tests := []struct {
		dataRate int
		pga      int
		response []byte
		expected float64
	}{
		{8, 1, []byte{0xff, 0xff}, 4.99992},
		{8, 2, []byte{0xff, 0xff}, 2.49996},
		{8, 4, []byte{0xff, 0xff}, 1.24998},
		{8, 8, []byte{0xff, 0xff}, 0.62499},
		{16, 1, []byte{0xff, 0xff}, 4.99985},
		{16, 2, []byte{0xff, 0xff}, 2.49992},
		{32, 2, []byte{0xff, 0xff}, 2.49985},
		{32, 8, []byte{0x00, 0x37}, 0.0021},
		{128, 2, []byte{0xff, 0xff}, 2.49939},
	}

	for _, test := range tests {
		ads.setDataRate(test.dataRate)
		ads.pga = test.pga

		data <- test.response
		v, _ := ads.Voltage(1)
		assert.Equal(t, test.expected, round(v))
	}
}

func TestADS1110Voltage(t *testing.T) {
	data := make(chan []byte, 1)
	c := iotest.NewI2CConn()

	conn, _ := i2c.Open(iotest.NewI2CDriver(c), 0x1)
	ads, _ := NewADS1110(conn, 240, 2)
	c.TxFunc(func(w, r []byte) error {
		copy(r, <-data)
		return nil
	})

	tests := []struct {
		dataRate int
		pga      int
		response []byte
		expected float64
	}{
		{15, 1, []byte{0xff, 0xff}, 2.04797},
		{15, 2, []byte{0x3f, 0x9f}, 0.25448},
		{30, 4, []byte{0x3f, 0x9f}, 0.25448},
		{30, 8, []byte{0x00, 0xae}, 0.00136},
		{60, 2, []byte{0x11, 0x2e}, 0.27487},
		{60, 1, []byte{0x11, 0x2e}, 0.54975},
		{240, 1, []byte{0x0b, 0x77}, 1.4675},
		{240, 8, []byte{0xc0, 0x83}, 0.00819},
	}

	for _, test := range tests {
		ads.setDataRate(test.dataRate)
		ads.pga = test.pga

		data <- test.response
		v, _ := ads.Voltage(1)
		assert.Equal(t, test.expected, round(v))
	}
}

func round(f float64) float64 {
	shift := math.Pow(10, 5)
	return math.Floor((f*shift)+0.5) / shift
}

func ExampleADS1100() {
	d, err := i2c.Open(&i2c.Devfs{
		Dev: "/dev/i2c-0",
	}, 0x1c)

	if err != nil {
		panic(fmt.Sprintf("failed to open device: %v", err))
	}
	defer d.Close()

	// 4.048 is Vref, 16 is the data rate and the PGA is set to 1.
	adc, err := NewADS1100(d, 4.048, 16, 1)

	if err != nil {
		panic(fmt.Sprintf("failed to create ADS1100: %v", err))
	}

	// Retrieve voltage of channel 1...
	v, err := adc.Voltage(1)

	if err != nil {
		panic(fmt.Sprintf("failed to read channel 1 of ADS1100: %s", err))
	}

	// ...read the raw value of channel 1. PGA has not been applied.
	c, err := adc.OutputCode(1)

	if err != nil {
		panic(fmt.Sprintf("failed to read channel 1 of ADS1100: %s", err))
	}

	fmt.Printf("channel 1 reads %f or digital output code  %d", v, c)
}

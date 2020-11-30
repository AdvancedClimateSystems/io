[![godoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/AdvancedClimateSystems/io/i2c/microchip)

# Microchip

Package microchip implements drivers for I<sup>2</sup>C controlled IC's
produced by [Microchip](http://www.microchip.com).

Drivers for the following IC's are implemented:

* [MCP4725](http://www.microchip.com/wwwproducts/DevicePrint/en/MCP4725?httproute=True)

Sample usage:


```go
package main

import (
	"fmt"

	"github.com/advancedclimatesystems/io/i2c/microchip"
	"golang.org/x/exp/io/i2c"
)

func main() {
	d, err := i2c.Open(&i2c.Devfs{
		Dev: "/dev/i2c-1",
	}, 0x60)

	if err != nil {
		panic(fmt.Sprintf("failed to open device: %v", err))
	}
	defer d.Close()

	// Reference voltage is 2.7V.
	dac, err := microchip.NewMCP4725(d, 2.7)

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
```

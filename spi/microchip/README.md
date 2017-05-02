[![godoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/AdvancedClimateSystems/io/spi/adc)

# Microchip

Package microchip implements drivers for a few SPI controlled chips produced by
[Microchip](http://www.microchip.com/). This package relies on
[x/exp/io/spi](https://godoc.org/golang.org/x/exp/io/spi).

MCP3x0x is a family of Analog Digital Converters (ADC).
Currently the package contains drivers for the following ADC:

* [MCP3004](http://www.microchip.com/wwwproducts/en/MCP3004)
* [MCP3008](http://www.microchip.com/wwwproducts/en/MCP3008)
* [MCP3204](http://www.microchip.com/wwwproducts/en/MCP3204)
* [MCP3208](http://www.microchip.com/wwwproducts/en/MCP3208)

Sample usage::

``` go
package main

import (
	"fmt"

	"golang.org/x/exp/io/spi"
	"github.com/advancedclimatesystems/io/spi/microchip"
)

func main() {
	conn, err := spi.Open(&spi.Devfs{
		Dev:      "/dev/spidev32766.0",
		Mode:     spi.Mode0,
		MaxSpeed: 3600000,
	})

	if err != nil {
		panic(fmt.Sprintf("failed to open SPI device: %s", err))
	}

	defer conn.Close()

	adc := microchip.MCP3008{
		Conn: conn,
		Vref: 5.0,
	}

        // Read the voltage of channel 3...
	v, err := adc.Voltage(3)

	if err != nil {
		panic(fmt.Sprintf("failed to read channel 3 of MCP3008: %s", err))
	}

        // ...or read the raw value of channel 3.
        c, err := adc.OutputCode(3)

	fmt.Printf("channel 3 reads %f Volts or digital output code %d", v, c)
}
```

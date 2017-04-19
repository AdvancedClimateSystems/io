[![godoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/AdvancedClimateSystems/io/spi/adc)

# ADC

Package ADC implements a few Analog Digital Converters (ADC). Communication
with the ADC is done using the Serial Peripheral Interface (SPI) and it relies
on the [x/exp/io/spi](https://godoc.org/golang.org/x/exp/io/spi) package.

An example::

``` go
package main

import (
	"fmt"

	"golang.org/x/exp/io/spi"
	"github.com/advancedclimatesystems/io/spi/adc"
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

	a := adc.MCP3008{
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
```
## Supported ADC's

* [MCP3004](http://www.microchip.com/wwwproducts/en/MCP3004)
* [MCP3008](http://www.microchip.com/wwwproducts/en/MCP3008)
* [MCP3204](http://www.microchip.com/wwwproducts/en/MCP3204)
* [MCP3208](http://www.microchip.com/wwwproducts/en/MCP3208)

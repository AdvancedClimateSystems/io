[![godoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/AdvancedClimateSystems/io/i2c/max)

# MAX

Package microchip implements drivers for a few I<sub>2</sub>C controlled IC's
produced by [Maxim Integrated](https://www.maximintegrated.com). This package
relies on [x/exp/io/spi](https://godoc.org/golang.org/x/exp/io/i2c).

Drivers for the following IC's are implemented:

* [MAX5813](https://www.maximintegrated.com/en/products/analog/data-converters/digital-to-analog-converters/MAX5813.html)
* [MAX5814](https://www.maximintegrated.com/en/products/analog/data-converters/digital-to-analog-converters/MAX5814.html)
* [MAX5815](https://www.maximintegrated.com/en/products/analog/data-converters/digital-to-analog-converters/MAX5815.html)

Sample usage:

```go
package main

import (
	"fmt"

	"github.com/advancedclimatesystems/io/i2c/max"
	"golang.org/x/exp/io/i2c"
)

func main() {
	d, err := i2c.Open(&i2c.Devfs{
		Dev: "/dev/i2c-0",
	}, 0x1c)

	if err != nil {
		panic(fmt.Sprintf("failed to open device: %v", err))
	}
	defer d.Close()


	// 2.5 is the input reference of the DAC.
	dac, err := max.NewMAX5813(d, 2.5)

	if err != nil {
		panic(fmt.Sprintf("failed to create MAX5813: %v", err))
	}

        // Set output of channel 1 to 1.3V.
        if err := dac.SetVoltage(1.3, 1); err != nil {
		panic(fmt.Sprintf("failed to set voltage: %v", err))
        }

        // It's also possible to set output of a channel with digital output code.
        if err := dac.SetInutCode(128, 1); err != nil {
		panic(fmt.Sprintf("failed to set voltage using output code: %v", err))
        }
}
```

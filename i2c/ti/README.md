[![godoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/AdvancedClimateSystems/io/i2c/ti)

# Texas Instruments

Package ti implements drivers for I<sup>2</sup>C controlled IC's
produced by [Texas Instruments](http://www.ti.com).

Drivers for the following IC's are implemented:

* [ADS1100](http://www.ti.com/lit/ds/symlink/ads1100.pdf)
* [ADS1110](http://www.ti.com/lit/ds/symlink/ads1110.pdf)
* [DAC5578](http://www.ti.com/product/dac5578)
* [DAC6578](http://www.ti.com/product/dac6578)
* [DAC7578](http://www.ti.com/product/dac7578)

Sample usage:


```go
package main

import (
	"fmt"

	"github.com/advancedclimatesystems/io/i2c/ti"
	"golang.org/x/exp/io/i2c"
)

func main() {
	// We are going to write 5.5 volt to channel 0.
	volts := 5.5
	channel := 0

	dev, err := i2c.Open(&i2c.Devfs{
		Dev: "/dev/i2c-0",
	}, 0x48)

	if err != nil {
		panic(fmt.Sprintf("failed to open device: %v", err))
	}
	defer dev.Close()

	// Create the DAC. The reference voltage is set to 10V.
	dac := ti.NewDAC5578(dev, 10)

	// Write volts to the channel.
	if err = dac.SetVoltage(volts, channel); err != nil {
		panic(fmt.Sprintf("failed to set voltage: %v", err))
	}

	// It's also possible to set output of a channel with digital output
        // code. Because the DAC5578 has a resolution of 8 bits the value must
        // be between 0 and 255.
	if err := dac.SetInputCode(255, channel); err != nil {
		panic(fmt.Sprintf("failed to set voltage using output code: %v", err))
	}
}
```

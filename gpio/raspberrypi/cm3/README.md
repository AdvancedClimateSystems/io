[![godoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/AdvancedClimateSystems/io/raspberrypi/cm3")

# Compute Module 3

Package cm3 implements drivers for the GPIO of the [cm3](https://www.raspberrypi.org/documentation/hardware/computemodule/datasheets/rpi_DATA_CM3plus_1p0.pdf).

Sample usage:


```go
package main

import (
	"log"
	"time"

	"github.com/advancedclimatesystems/io/gpio/raspberrypi/cm3"
	"github.com/advancedclimatesystems/io/gpio"
)

func main() {
	outPin, _ := cm3.NewPin(16)
	_ = outPin.SetDirection(gpio.OutDirection)

	inPin, _ := cm3.NewPin(17)
	_ = inPin.SetDirection(gpio.InDirection)
	_ = inPin.SetEdge(gpio.RisingEdge, func(p *gpio.Pin) {
		log.Printf("wow")
	})

	for i := 0; i < 4; i++ {
		_ = outPin.SetHigh()
		time.Sleep(1000 * time.Millisecond)
		_ = outPin.SetLow()
		time.Sleep(1000 * time.Millisecond)
	}
}
```

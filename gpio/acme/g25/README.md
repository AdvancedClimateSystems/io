[![godoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/AdvancedClimateSystems/io/acme/g25)

# Aria G25

Package g25 implements drivers for the GPIO of the [Aria G25](https://www.acmesystems.it/aria) produced by [Acme Systems](https://www.acmesystems.it/).

Sample usage:


```go
package main

import (
	"log"
	"time"

	"github.com/advancedclimatesystems/io/acme/g25"
	"github.com/advancedclimatesystems/io/gpio"
)

func main() {
	outPin, _ := g25.NewPin("N16")
	_ = outPin.SetDirection(gpio.OutDirection)

	inPin, _ := g25.NewPin("N20")
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

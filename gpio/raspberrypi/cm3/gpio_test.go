package cm3

import (
	"log"
	"time"

	"github.com/advancedclimatesystems/io/gpio"
)

func ExampleNewPin() {
	outPin, _ := NewPin(16)
	_ = outPin.SetDirection(gpio.OutDirection)

	inPin, _ := NewPin(17)
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

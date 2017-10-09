package g25

import (
	"errors"
	"log"
	"testing"
	"time"

	"github.com/advancedclimatesystems/io/gpio"
	"github.com/stretchr/testify/assert"
)

func TestGetKernelID(t *testing.T) {
	tests := []struct {
		kv  int
		id  string
		err error
		kid int
	}{
		{2, "N2", nil, 96},
		{3, "N2", nil, 64},
		{2, "E2", nil, 118},
		{3, "E2", nil, 86},

		{2, "N999", errors.New("id N999 not known"), 0},
		{3, "N999", errors.New("id N999 not known"), 0},
		{2, "W0", errors.New("id W0 not known"), 0},
		{3, "W0", errors.New("id W0 not known"), 0},

		{1, "N2", nil, 96},
		{4, "N2", nil, 64},
	}
	for _, test := range tests {
		kid, err := getkernelID(test.kv, test.id)
		assert.Equal(t, test.kid, kid)
		assert.Equal(t, test.err, err)
	}
}

func ExampleNewPin() {
	outPin, _ := NewPin("N16")
	_ = outPin.SetDirection(gpio.OutDirection)

	inPin, _ := NewPin("N20")
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

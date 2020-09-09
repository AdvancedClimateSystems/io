// +build linux

// Package cm3 contains GPIO drivers for the Raspberry PI Compute Module 3
package cm3

import (
	"fmt"

	"github.com/advancedclimatesystems/io/gpio"
)

var w gpio.Watcher

// NewPin creates a new pin. The available IDs can be found here:
// https://www.raspberrypi.org/documentation/hardware/computemodule/datasheets/rpi_DATA_CM3plus_1p0.pdf
func NewPin(id int) (gpio.GPIO, error) {
	if err := setupWatcher(); err != nil {
		return nil, err
	}

	// The file created by export is always called pio, followed by ic pin ID,
	// but without the first character. So exporting N2 gives an file called pioC0.
	gpio := gpio.NewPin(id, fmt.Sprintf("gpio%v", id), w)
	if err := gpio.Export(); err != nil {
		return nil, err
	}
	return gpio, nil
}

// setupWatcher creates a new watcher and starts it, if its not already running.
func setupWatcher() error {
	// A Watcher only needs to be setup once, but an error can't be handled in an
	// init function.
	var err error
	if w == nil {
		w, err = gpio.NewWatcher()
		if err != nil {
			return err
		}
		go func() {
			err = w.Watch()
			defer w.Close()
		}()
	}
	return err
}

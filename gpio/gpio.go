// +build linux

// Package gpio contains an interface and implementation for controlling GPIO
// pins via the sysfs interface.
//
// This packages does not contain any vendor specific implementations of GPIO
// pins, however the Pin struct in this package can be embedded in another
// struct which implements vendor specific functionality.
package gpio

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

// basePath is where the GPIO pins can be found.
const basePath = "/sys/class/gpio"

// Edge describes on what edge a function should be called.
type Edge string

const (
	// RisingEdge is triggered when the values goes from low to high.
	RisingEdge Edge = "rising"
	// FallingEdge is triggered when the values goes from high to low.
	FallingEdge Edge = "falling"
	// BothEdge is triggered when the values goes from high to low or from low to high.
	BothEdge Edge = "both"
	// NoneEdge never triggers.
	NoneEdge Edge = "none"
)

// EdgeEvent is a type of function that can be used as a vcallback to watcher
type EdgeEvent func(pin *Pin)

// Direction is the direction of the dataflow.
type Direction string

const (
	// InDirection means a value can be read from the pin.
	InDirection Direction = "in"
	// OutDirection means a value can written to the pin.
	OutDirection Direction = "out"
)

// GPIO is an interface for GPIO pins.
type GPIO interface {
	Value() (int, error)
	SetHigh() error
	SetLow() error

	Direction() (Direction, error)
	SetDirection(d Direction) error

	Edge() (Edge, error)
	SetEdge(edge Edge, f EdgeEvent) error

	ActiveLow() (bool, error)
	SetActiveLow(invert bool) error

	Export() error
	Unexport() error
}

// Pin is an implementation of the GPIO interface. It can be embedded in vendor
// specific implentations.
type Pin struct {
	KernelID int
	// The kernel ID is often needed as []byte to write to a file.
	kernelIDByte []byte
	pinBase      string
	rwHelper     rwHelper
	w            Watcher
}

// NewPin creates an instance of Pin.
// The kernelID is the ID used to expose the pin. The filename is the name of
// the folder that contains files such as value and edge. This folder gets
// created when the in is exported and is often named gpio<kernelID>.
func NewPin(kernelID int, pinBase string, w Watcher) *Pin {
	return &Pin{
		KernelID:     kernelID,
		kernelIDByte: []byte(strconv.Itoa(kernelID)),
		pinBase:      pinBase,
		rwHelper:     new(baseReaderWriter),
		w:            w,
	}
}

// Direction returns the curent direction of the pin.
func (p *Pin) Direction() (Direction, error) {
	b := make([]byte, 3)
	n, err := p.read(b, "direction")
	if err != nil {
		return OutDirection, err
	}
	if n == 0 {
		return OutDirection, errors.New("not enough bytes to read")
	}
	if string(b[:n]) == "out" {
		return OutDirection, nil
	}
	if string(b[:n-1]) == "in" {
		return InDirection, nil
	}
	return OutDirection, fmt.Errorf("not a known direction: '%v'", string(b[:n]))
}

// SetDirection configures the pin as an input or output.
func (p *Pin) SetDirection(d Direction) error {
	data := []byte(d)
	return p.write(data, "direction")
}

// Value returns the value of the pin. The pin must be in the 'in' direction.
func (p *Pin) Value() (int, error) {
	b := make([]byte, 1)
	n, err := p.read(b, "value")
	if err != nil {
		return 0, err
	}
	if n != 1 {
		return 0, fmt.Errorf("expected 1 byte, got %v", n)
	}
	if string(b[:n]) == "1" {
		return 1, nil
	}
	if string(b[:n]) == "0" {
		return 0, nil
	}
	return 0, fmt.Errorf("not a known value: '%v'", string(b[:n]))
}

// SetLow writes a 0 to the Pin
func (p *Pin) SetLow() error {
	data := []byte("0")
	return p.write(data, "value")
}

// SetHigh writes a 1 to the Pin. It also sets the pins direction to output.
func (p *Pin) SetHigh() error {
	data := []byte("1")
	return p.write(data, "value")
}

// ActiveLow returns true if the the pin is inverted, i.e. it is true when
// the value is low
func (p *Pin) ActiveLow() (bool, error) {
	b := make([]byte, 1)
	n, err := p.read(b, "active_low")
	if err != nil {
		return false, err
	}
	if n != 1 {
		return false, fmt.Errorf("expected 1 byte, got %v", n)
	}
	if string(b[:n]) == "1" {
		return true, nil
	}
	if string(b[:n]) == "0" {
		return false, nil
	}
	return false, fmt.Errorf("not a known value: '%v'", string(b[:n]))
}

// SetActiveLow inverts the pins value, i.e. it is true when
// the value is low.
func (p *Pin) SetActiveLow(invert bool) error {
	var data []byte

	data = []byte("0")
	if invert {
		data = []byte("1")
	}

	return p.write(data, "active_low")
}

// Edge returns the current edge of the pin.
func (p *Pin) Edge() (Edge, error) {
	b := make([]byte, 8)
	n, err := p.read(b, "edge")
	if err != nil {
		return NoneEdge, err
	}
	if n == 0 {
		return NoneEdge, errors.New("not enough bytes to read")
	}

	switch string(b[:n-1]) {
	case "rising":
		return RisingEdge, nil
	case "falling":
		return FallingEdge, nil
	case "both":
		return BothEdge, nil
	case "none":
		return NoneEdge, nil
	default:
		return NoneEdge, fmt.Errorf("not a known value: '%v'", string(b[:n]))
	}
}

// SetEdge sets an edge and sets up event handing for given edge. An edge can
// only be set on a pin with the 'in' direction.
func (p *Pin) SetEdge(e Edge, f EdgeEvent) error {
	b := []byte(e)
	valF, err := os.OpenFile(fmt.Sprintf("%v/%v/value", basePath, p.pinBase), os.O_RDWR, 0777)
	if err != nil {
		return err
	}
	// Wrap the callback function, so that the pin can be used as a parameter.
	callback := func() {
		f(p)
	}
	if err = p.w.AddEvent(int(valF.Fd()), callback); err != nil {
		return err
	}
	return p.write(b, "edge")
}

// Export exports the pin, if it wasn't exported already.
func (p *Pin) Export() error {
	err := p.rwHelper.writeFromBase(p.kernelIDByte, "export")
	// The 'device or resource busy' error indicates the pin has already been
	// exported. Checking for specific error is a bit weird in Go. Maybe proper
	// error handling will come with Go 2.0 ....
	if fmt.Sprintf("%v", err) == fmt.Sprintf("write %v/export: device or resource busy", basePath) {
		return nil
	}
	return err
}

// Unexport unexports the pin.
func (p *Pin) Unexport() error {
	return p.rwHelper.writeFromBase(p.kernelIDByte, "unexport")
}

func (p *Pin) read(b []byte, file string) (int, error) {
	return p.rwHelper.readFromBase(b, fmt.Sprintf("%v/%v", p.pinBase, file))
}

func (p *Pin) write(b []byte, file string) error {
	return p.rwHelper.writeFromBase(b, fmt.Sprintf("%v/%v", p.pinBase, file))
}

// rwHelper is a seperate interface for interacting with files. This makes it
// possible to create mocks for reading and writing files, which is needed to
// write proper tests.
type rwHelper interface {
	readFromBase(b []byte, pathFromBase string) (int, error)
	writeFromBase(b []byte, pathFromBase string) error
}

// baseReaderWriter has methods to read/write gpio-related files.
type baseReaderWriter struct{}

// readFromBase reads data from a file into b.
func (baseReaderWriter) readFromBase(b []byte, pathFromBase string) (int, error) {
	f, err := os.OpenFile(fmt.Sprintf("%v/%v", basePath, pathFromBase), os.O_RDONLY, 0777)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return f.Read(b)
}

// readFromBase writeFromBase writes data to a file.
func (baseReaderWriter) writeFromBase(b []byte, pathFromBase string) error {
	f, err := os.OpenFile(fmt.Sprintf("%v/%v", basePath, pathFromBase), os.O_WRONLY, 0777)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(b)
	return err
}

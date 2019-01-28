// +build linux

// Package g25 contains GPIO drivers for the Acme Systems Aria G25
//
// The Aria G25 contains up to 60 GPIO pins. This package implements all GPIO
// operations such as getting/setting the value, setting the direction and
// changing the active low. The package provides a mapping between the
// pinnumber, the atmel ID and the kernel ID.
// https://www.acmesystems.it/aria
package g25

import (
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/advancedclimatesystems/io/gpio"
)

var w gpio.Watcher

// NewPin creates a new pin with a kernel ID based on the pin name found
// here: https://www.acmesystems.it/aria. It assumes the kernel has version 3.1x
// if this is not the case, use the NewPinV26 instead.
func NewPin(id string) (gpio.GPIO, error) {
	k, err := getKernelVersion()
	if err != nil {
		return nil, err
	}

	kernelID, err := getkernelID(k, id)
	if err != nil {
		return nil, err
	}

	if err := setupWatcher(); err != nil {
		return nil, err
	}

	// The file created by export is always called pio, followed by ic pin ID,
	// but without the fist character. So exporting N2 gives an file called pioC0.
	gpio := gpio.NewPin(kernelID, fmt.Sprintf("pio%v", g25Id[id].icPin[1:]), w)
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

// getKernelID returns the corrent kernel ID, based on the kernel version and id.
func getkernelID(k int, id string) (int, error) {
	pinID, ok := g25Id[id]
	if !ok {
		return 0, fmt.Errorf("id %v not known", id)
	}

	var kernelID int
	if k < 3 {
		kernelID = pinID.kernelID26
	} else {
		kernelID = pinID.kernelID31
	}
	return kernelID, nil
}

// getkernelVersion get s the kernel version from /proc/version
func getKernelVersion() (int, error) {
	data, err := ioutil.ReadFile("/proc/version")
	if err != nil || len(data) < 15 {
		return 0, err
	}
	// The 14th byte should be the major version, which is what we are loooking for.
	return strconv.Atoi(string(data[14]))
}

// g25Id provides a mapping between the "pin name" and all other indentifiers
// of a pin.
var g25Id = map[string]struct {
	// Names such as N2, N3, E10, E6
	pinName string
	// Names such as PC0, PC1, PC30, PC26. These are used to figure out the
	// name of the file created by exporting the pin.
	icPin string
	// Kernel ID for version 2.6.
	kernelID26 int
	// Kernel ID for version 3.1.
	kernelID31 int
}{
	"N2":  {"N2", "PC0", 96, 64},
	"N3":  {"N3", "PC1", 97, 65},
	"N4":  {"N4", "PC2", 98, 66},
	"N5":  {"N5", "PC3", 99, 67},
	"N6":  {"N6", "PC4", 100, 68},
	"N7":  {"N7", "PC5", 101, 69},
	"N8":  {"N8", "PC6", 102, 70},
	"N9":  {"N9", "PC7", 103, 71},
	"N10": {"N10", "PC8", 104, 72},
	"N11": {"N11", "PC9", 105, 73},
	"N12": {"N12", "PC10", 106, 74},
	"N13": {"N13", "PC11", 107, 75},
	"N14": {"N14", "PC12", 108, 76},
	"N15": {"N15", "PC13", 109, 77},
	"N16": {"N16", "PC14", 110, 78},
	"N17": {"N17", "PC15", 111, 79},
	"N18": {"N18", "PC16", 112, 80},
	"N19": {"N19", "PC17", 113, 81},
	"N20": {"N20", "PC18", 114, 82},
	"N21": {"N21", "PC19", 115, 83},
	"N22": {"N22", "PC20", 116, 84},
	"N23": {"N23", "PC21", 117, 85},

	"E2":  {"E2", "PC22", 118, 86},
	"E3":  {"E3", "PC23", 119, 87},
	"E4":  {"E4", "PC24", 120, 88},
	"E5":  {"E5", "PC25", 121, 89},
	"E6":  {"E6", "PC26", 122, 90},
	"E7":  {"E7", "PC27", 123, 91},
	"E8":  {"E8", "PC28", 124, 92},
	"E9":  {"E9", "PC29", 125, 93},
	"E10": {"E10", "PC30", 126, 94},
	"E11": {"E11", "PC31", 127, 95},

	"S2":  {"S2", "PA21", 53, 21},
	"S3":  {"S3", "PA20", 52, 20},
	"S4":  {"S4", "PA19", 51, 19},
	"S5":  {"S5", "PA18", 50, 18},
	"S6":  {"S6", "PA17", 49, 17},
	"S7":  {"S7", "PA16", 48, 16},
	"S8":  {"S8", "PA15", 47, 15},
	"S9":  {"S9", "PA14", 46, 14},
	"S10": {"S10", "PA13", 45, 13},
	"S11": {"S11", "PA12", 44, 12},
	"S12": {"S12", "PA11", 43, 11},
	"S15": {"S15", "PA8", 40, 8},
	"S16": {"S16", "PA7", 39, 7},
	"S17": {"S17", "PA6", 38, 6},
	"S18": {"S18", "PA5", 37, 5},
	"S19": {"S19", "PA4", 36, 4},
	"S20": {"S20", "PA3", 35, 3},
	"S21": {"S21", "PA2", 34, 2},
	"S22": {"S22", "PA1", 33, 1},
	"S23": {"S23", "PA0", 32, 0},

	"W9":  {"W9", "PA22", 54, 22},
	"W10": {"W10", "PA23", 55, 23},
	"W11": {"W11", "PA24", 56, 24},
	"W12": {"W12", "PA25", 57, 25},
	"W13": {"W13", "PA26", 58, 26},
	"W14": {"W14", "PA27", 59, 27},
	"W15": {"W15", "PA28", 60, 28},
	"W16": {"W16", "PA29", 61, 29},
	"W17": {"W17", "PA30", 62, 30},
	"W18": {"W18", "PA31", 63, 31},
	"W20": {"W20", "PB11", 75, 43},
	"W21": {"W21", "PB12", 76, 44},
	"W22": {"W22", "PB13", 77, 45},
	"W23": {"W23", "PB14", 78, 46},
}

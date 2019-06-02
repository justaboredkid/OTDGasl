/*
Program for parsing ASL using OTDG.

OTDG IS NOT ONLY DESIGNED FOR ASL. OTDG = Open Tactile Data Glove

Copyright (C) 2019  Henry Lo

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.

*/

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/justaboredkid/OTDGasl/asllibs"

	"github.com/paypal/gatt"
	"github.com/paypal/gatt/examples/option"
	"github.com/shantanubhadoria/go-kalmanfilter/kalmanfilter"
	"github.com/warthog618/gpio"
)

var debug *bool
var glove asllibs.Hand
var dict []asllibs.ASLdict // slice of ASLdict, not
var oldTime time.Time = time.Now()

var myFilterData = new(kalmanfilter.FilterData)

// Checks and reading GPIO
func check(e error) {
	if e != nil {
		panic(e)
	}
}

func bleServer() {
	device, err := gatt.NewDevice(option.DefaultServerOptions...)
	if err != nil {
		log.Fatalf("Failed to open device, err: %s", err)
	}

	// Register optional handlers.
	// note: implement a ID filter here to avoid tempering
	device.Handle(
		gatt.PeripheralConnected(func(p gatt.Peripheral, err error) { fmt.Println("Connect: ", p.ID()) }),
		gatt.PeripheralDisconnected(func(p gatt.Peripheral, err error) { fmt.Println("Disconnect: ", p.ID()) }),
	)

	// A mandatory handler for monitoring device state.
	onStateChanged := func(device gatt.Device, s gatt.State) {
		fmt.Printf("State: %s\n", s)
		switch s {
		case gatt.StatePoweredOn:
			// Setup GAP and GATT services for Linux implementation.
			// OS X doesn't export the access of these services.
			s := asllibs.OrientationData() // no effect on OS X
			device.AddService(s)

			// Advertise device name and service's UUIDs.
			device.AdvertiseNameAndServices("OTDGasl", []gatt.UUID{s.UUID()})

			// Advertise as an OpenBeacon iBeacon
			device.AdvertiseIBeacon(gatt.MustParseUUID("d86e828c-658a-4373-9d4c-8a26c5cc73fd"), 1, 2, -59)

		}
	}

	device.Init(onStateChanged)
	select {}
}

func pinRead(pin *gpio.Pin) bool {
	pin.PullUp()
	pin.Input()
	return bool(pin.Read())
}

func init() {
	err := gpio.Open()
	check(err)

	bleServer()

	debug = flag.Bool("debug", false, "Enables Debug Logging")
	flag.Parse()

	// Go thorugh all files and folders in "data/"
	err = filepath.Walk("data/",
		func(path string, _ os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Check if path is folder
			fi, err := os.Stat(path)
			check(err)
			if fi.IsDir() {
				return nil
			}

			f, err := ioutil.ReadFile(path)
			check(err)
			fmt.Printf("%v loaded\n", path[5:])
			json.Unmarshal(f, &dict)
			return nil
		})
	check(err)
}

// Main recognition
func main() {

	if *debug {
		for i := range dict {
			fmt.Printf("ID: %v loaded\n", dict[i].ID)
		}
	}

	fmt.Println("ASL Parsing [STARTED]")

	// Graceful shutdown handling
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)

	go func() {
		select {
		case sig := <-c:
			gpio.Close()
			fmt.Printf("Got %s signal. Exiting...\n", sig)
			os.Exit(0)
		}
	}()

	// NOTE TO SELF: use location for signs pointing downwards (ie Q)
	for {
		glove = asllibs.Hand{
			Pinky:     pinRead(gpio.NewPin(2)),
			Ring:      pinRead(gpio.NewPin(3)),
			Middle:    pinRead(gpio.NewPin(4)),
			Index:     pinRead(gpio.NewPin(5)),
			Thumb:     pinRead(gpio.NewPin(6)),
			PalmLeft:  pinRead(gpio.NewPin(7)),
			PalmRight: pinRead(gpio.NewPin(8)),
			BackThumb: pinRead(gpio.NewPin(9)),
			BackRing:  pinRead(gpio.NewPin(10)),
			BetwIM:    pinRead(gpio.NewPin(11)),
			BetwMR:    pinRead(gpio.NewPin(12)),
			BetwRP:    pinRead(gpio.NewPin(13)),
			Angle:     -1,
			Motion:    "",
		}
		time.Sleep(200 * time.Millisecond)
		for i := range dict {
			if glove == dict[i].Hand {
				fmt.Printf("%v [MATCH]\n", dict[i].ID)
			} else {
				if *debug {
					fmt.Printf("%v [NOT MATCHED]\n", dict[i].ID)
				}
			}
		}
	}

}

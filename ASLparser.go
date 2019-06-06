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
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/grandcat/zeroconf"
	"github.com/justaboredkid/OTDGasl/asllibs"
	"github.com/warthog618/gpio"
)

var o asllibs.Orientation
var debug *bool
var parse bool
var glove asllibs.Hand
var dict []asllibs.ASLdict // slice of ASLdict, not
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Checks and reading GPIO
func check(e error) {
	if e != nil {
		panic(e)
	}
}

// Main recognition
func buttonRead() {
	fmt.Println("ASL Parsing [STARTED]")

	// NOTE TO SELF: use location for signs pointing downwards (ie Q)
	go func() {
		for {
			// Graceful shutdown handling
			if parse {
				gpio.Close()
				fmt.Printf("ASL parsing [STOPPED]")
				return
			}
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
				Angle:     o,
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
	}()
}

func pinRead(pin *gpio.Pin) bool {
	pin.PullUp()
	pin.Input()
	return bool(pin.Read())
}

func init() {
	err := gpio.Open()
	check(err)

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

func main() {
	if *debug {
		for i := range dict {
			fmt.Printf("ID: %v loaded\n", dict[i].ID)
		}
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	select {
	case <-sig:
		// Exit by user
		gpio.Close()
	case <-time.After(time.Second * 120):
		// Exit by timeout
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var conn, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print(err)
			return
		}

		go func(c *websocket.Conn) {
			_, msg, err := c.ReadMessage()
			if err != nil {
				c.Close()
				fmt.Printf("[ERR] %v\n", err)
			}

			if msg == nil {
				fmt.Println("[ERR] No message from client")
			}

			err = json.Unmarshal(msg, &o)
			if err != nil {
				c.Close()
				fmt.Println(err)
			}
		}(conn)
		buttonRead()
	})

	go func() {
		server, err := zeroconf.Register("OTDGasl", "_OTDGmain._tcp", "local.", 443, []string{"txtv=0", "lo=1", "la=2"}, nil)
		if err != nil {
			panic(err)
		}

		defer server.Shutdown()
	}()

	err := http.ListenAndServeTLS(":443", "certs/server.pem", "certs/key.pem", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

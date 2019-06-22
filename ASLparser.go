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
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/gorilla/websocket"
	"github.com/justaboredkid/OTDGasl/asllibs"
	"github.com/warthog618/gpio"
)

var o asllibs.Orientation
var debug *bool
var noOrien *bool
var local *bool
var parse bool
var err error
var glove asllibs.Hand
var dict []asllibs.ASLdict
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
	log.Println("ASL Parsing [STARTED]")
	if !parse {
		log.Println("[ERR] Parse bool not true.")
	}
	// NOTE TO SELF: use location for signs pointing downwards (ie Q)

	if *noOrien {
		log.Println("[WARN] Orientation sensors disabled, interperetation limited")
		for {
			if !parse {
				gpio.Close()
				log.Printf("ASL parsing [STOPPED]")
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
				Angle: asllibs.Orientation{
					Alpha: 0,
					Beta:  0,
					Gamma: 0,
				},
				Motion: "",
				Dom:    true,
			}

			time.Sleep(200 * time.Millisecond)
			for i := range dict {
				if glove == dict[i].Hand {
					log.Printf("%v [MATCH]\n", dict[i].ID)
				}
			}
		}
	}

	go func() {
		for {
			// Graceful shutdown handling
			if !parse {
				gpio.Close()
				log.Printf("ASL parsing [STOPPED]")
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
				Angle: asllibs.Orientation{
					Alpha: 0,
					Beta:  0,
					Gamma: o.Gamma,
				},
				Motion: "",
				Dom:    true,
			}
			time.Sleep(200 * time.Millisecond)
			for i := range dict {
				if glove == dict[i].Hand {
					log.Printf("%v [MATCH]\n", dict[i].ID)
				}
			}
		}
	}()
}

func pinRead(pin *gpio.Pin) bool {
	pin.Input()
	pin.PullUp()
	if *debug && !bool(pin.Read()) == true {
		log.Printf("[INFO] %v button pressed\n", pin)
	}
	return !bool(pin.Read())
}

func init() {
	debug = flag.Bool("debug", false, "Enables Debug Logging")
	noOrien = flag.Bool("noOrien", false, "Ignores orientation and WS connect. Will limit interpretation")
	local = flag.Bool("local", false, "Starts in http only. DO NOT USE IN PRODUCTION")
	flag.Parse()

	logFile, err := os.OpenFile("log.txt", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}

	if *debug {
		lw := io.MultiWriter(os.Stdout, logFile)
		log.SetOutput(lw)
	} else {
		log.SetOutput(os.Stdout)
	}

	if *local {
		log.Println("[WARN] HTTP only mode. If you are seeing this message in production, you left the -local flag enabled")
	}

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
			log.Printf("[ASL FILE] %v loaded\n", path[5:])
			json.Unmarshal(f, &dict)
			return nil
		})
	check(err)

	go func() {
		sigchan := make(chan os.Signal, 10)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan
		log.Println("[INFO] Program exiting")

		if parse {
			parse = false
			gpio.Close()
		}
		os.Exit(0)
	}()

}

// By Artem Co on stackOverflow
func keepAlive(c *websocket.Conn, timeout time.Duration) {
	lastResponse := time.Now()
	c.SetPongHandler(func(msg string) error {
		lastResponse = time.Now()
		return nil
	})

	go func() {
		for {
			err := c.WriteMessage(websocket.PingMessage, []byte("keepalive"))
			if err != nil {
				return
			}
			time.Sleep(timeout / 2)
			if time.Now().Sub(lastResponse) > timeout {
				c.Close()
				return
			}
		}
	}()
}

func main() {
	if *debug {
		for i := range dict {
			log.Printf("[ASL LOAD] ID: %v loaded\n", dict[i].ID)
		}
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		gpio.Open()
		check(err)

		var conn, err = upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		keepAlive(conn, 2*time.Second)
		log.Printf("[WS] %v CONNECTED", conn.RemoteAddr())

		go func() {
			for {
				_, msg, err := conn.ReadMessage()
				if err != nil {
					log.Printf("[ERR] %v\n", err)

					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						log.Printf("[WS] Connection Closed")
					}
					break
				}

				if msg == nil {
					log.Println("[ERR] Null message from client")
				} else {
					err = json.Unmarshal(msg, &o)
					if *debug {
						// WS debug here
					}
					if err != nil {
						log.Println(err)
					}

				}
			}
		}()

		parse = true
		buttonRead()
	})
	// deal with this later, deadline is approaching
	/* 	go func() {
		info := []string{"OTDG main glove"}
		service, err := mdns.NewMDNSService("otdgtest", "_otdgmain._tcp", "", "", 8000, nil, info)
		if err != nil {
			log.Fatal("[ERR] mDNS Service: ", err)
		}

		_, err = mdns.NewServer(&mdns.Config{Zone: service})
		log.Println("[mDNS] Server Started")
		if err != nil {
			log.Fatal("[ERR] mDNS Server: ", err)
		}
	}() */

	if *noOrien {
		gpio.Open()
		parse = true
		buttonRead()
	} else {
		if *local {
			log.Println("[HTTP (insecure)] Server Started")
			err = http.ListenAndServe(":80", nil)
		} else {
			log.Println("[HTTPS] Server Started")
			err = http.ListenAndServeTLS(":443", "certs/server.pem", "certs/key.pem", nil)
		}

		if err != nil {
			log.Fatal("[ERR] ListenAndServe: ", err)
		}
	}
}

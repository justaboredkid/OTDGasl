package asllibs

import (
	"fmt"

	"github.com/paypal/gatt"
)

// OrientationData - Orientation Characteristic
func OrientationData() *gatt.Service {
	s := gatt.NewService(gatt.MustParseUUID("100cff88-99cf-43ae-9289-ef07516ef6f4"))
	c := s.AddCharacteristic(gatt.UUID16(0x2A19))

	c.AddDescriptor(gatt.UUID16(0x2901)).SetValue([]byte("Orientation receiver"))

	c.HandleNotifyFunc(func(r gatt.Request, n gatt.Notifier) {
		fmt.Printf("%v has sent data", r.Central.ID())
	})

	return s
}

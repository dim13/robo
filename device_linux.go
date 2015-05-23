package robo

import "log"

func NewDevice() (dev Device) {
	dev, err := NewLP("/dev/usb/lp0")
	if err != nil {
		log.Fatal(err)
	}
	return dev
}

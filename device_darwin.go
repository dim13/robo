package robo

import "log"

func NewDevice() Device {
	dev, err := NewUSB()
	if err != nil {
		log.Fatal(err)
	}
	return dev
}

package robo

import "log"

func NewDevice() Device {
	dev, err := NewLP("/dev/ulpt0")
	if err != nil {
		log.Fatal(err)
	}
	return dev
}

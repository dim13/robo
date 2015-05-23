package robo

import "bufio"

type Device interface {
	Close()
	Handle() *bufio.ReadWriter
}

var devices = []string{
	"/dev/usb/lp0", // Linux
	"/dev/ulpt0",   // OpenBSD
}

func NewDevice() (dev Device) {
	var err error

	for _, d := range devices {
		if dev, err = NewLP(d); err == nil {
			return
		}
	}

	if dev, err = NewUSB(); err == nil {
		return
	}

	panic(err)
}

package robo

import (
	"bufio"
	"log"
	"runtime"
)

type Device interface {
	Close()
	Handle() *bufio.ReadWriter
}

func NewDevice() (dev Device) {
	var err error

	if runtime.GOOS == "linux" {
		dev, err = NewLP("/dev/usb/lp0")
	} else {
		dev, err = NewUSB()
	}

	if err != nil {
		log.Fatal(err)
	}

	return
}

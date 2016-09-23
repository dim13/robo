package robo

import (
	"io"
	"log"
)

func NewDevice() io.ReadWriteCloser {
	dev, err := NewUSB()
	if err != nil {
		log.Fatal(err)
	}
	return dev
}

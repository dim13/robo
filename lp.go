package main

import (
	"bufio"
	"os"
)

type LPDevice struct {
	*os.File
}

func NewLPDevice(path string) (Devicer, error) {
	f, err := os.OpenFile(path, os.O_RDWR, 0666)
	return LPDevice{f}, err
}

func (d LPDevice) Close() {
	d.File.Close()
}

func (d LPDevice) Handle() *bufio.ReadWriter {
	r := bufio.NewReader(d.File)
	w := bufio.NewWriter(d.File)
	return bufio.NewReadWriter(r, w)
}

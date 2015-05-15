package main

import (
	"bufio"
	"os"
)

type LP struct {
	*os.File
}

func NewLP(path string) (Devicer, error) {
	f, err := os.OpenFile(path, os.O_RDWR, 0666)
	return LP{f}, err
}

func (d LP) Close() {
	d.File.Close()
}

func (d LP) Handle() *bufio.ReadWriter {
	r := bufio.NewReader(d.File)
	w := bufio.NewWriter(d.File)
	return bufio.NewReadWriter(r, w)
}

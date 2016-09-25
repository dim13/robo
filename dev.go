package robo

import (
	"bufio"
	"io"
	"os"
	"syscall"
)

type Device struct {
	io.ReadWriteCloser
}

// Open is implemented in according GOOS files

// LP represents Line Printer
type LP struct {
	*os.File
}

func NewLP(path string) (LP, error) {
	f, err := os.OpenFile(path, os.O_RDWR, 0666)
	return LP{f}, err
}

func (d LP) Close() error {
	return d.File.Close()
}

func (d LP) SetNonblock() {
	fd := d.File.Fd()
	syscall.SetNonblock(int(fd), true)
}

func (d LP) Handle() *bufio.ReadWriter {
	r := bufio.NewReader(d.File)
	w := bufio.NewWriter(d.File)
	return bufio.NewReadWriter(r, w)
}

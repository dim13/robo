package robo

import (
	"bufio"
	"io"
	"os"
	"syscall"
)

const (
	NUL = byte(0x00) // Null
	ETX = byte(0x03) // End of Text
	ESC = byte(0x1B) // Escape
	FS  = byte(0x1C) // File Separator
)

type Device interface {
	io.ReadWriteCloser
	ReadString() (string, error)
	WriteString(string) error
	Command([]byte) error
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

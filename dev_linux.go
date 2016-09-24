package robo

import "io"

func Open() (io.ReadWriteCloser, error) {
	return NewLP("/dev/usb/lp0")
}

package robo

import "io"

type Device interface {
	io.ReadWriteCloser
}

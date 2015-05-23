package robo

import "bufio"

type Device interface {
	Close()
	Handle() *bufio.ReadWriter
}

package main

import "bufio"

type Device interface {
	Close()
	Handle() *bufio.ReadWriter
}

package main

import "bufio"

type Devicer interface {
	Close()
	Handle() *bufio.ReadWriter
}

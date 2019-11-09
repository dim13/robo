package main

import (
	"fmt"

	"github.com/dim13/robo"
)

func main() {
	dev := robo.NewDevice()
	defer dev.Close()
	handle := dev.Handle()
	fmt.Println(robo.Version(handle))
}

package main

import (
	"fmt"

	"dim13.org/robo"
)

func main() {
	dev := robo.NewDevice()
	defer dev.Close()
	handle := dev.Handle()
	fmt.Println(robo.Version(handle))
}

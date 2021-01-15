package main

import (
	"fmt"
	"log"

	"github.com/dim13/robo"
)

func main() {
	dev, err := robo.NewDevice()
	if err != nil {
		log.Fatal(err)
	}
	defer dev.Close()
	handle := dev.Handle()
	fmt.Println(robo.Version(handle))
}

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
	fmt.Println(robo.Version(dev.ReadWriter))
}

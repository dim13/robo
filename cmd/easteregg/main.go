package main

import (
	"log"

	"github.com/dim13/robo"
)

func main() {
	dev, err := robo.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer dev.Close()

	for _, cmd := range Easteregg {
		robo.Send(dev, cmd)
	}
}

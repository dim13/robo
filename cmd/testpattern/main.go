package main

import (
	"log"

	"github.com/dim13/robo"
)

func main() {
	dev, err := robo.NewRobo()
	if err != nil {
		log.Fatal(err)
	}
	defer dev.Close()

	dev.Initialize(robo.MediaPen, robo.Portrait)
	dev.UpperRight(robo.A4)
	dev.TestPattern()
}

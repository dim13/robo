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
	defer robo.Home(dev)
	if err := robo.Initialize(dev, 113, robo.Portrait); err != nil {
		log.Fatal("init: ", err)
	}
	robo.A4.Sub(robo.Margin).UpperRight(dev)
	robo.DrawMarks(dev)
}

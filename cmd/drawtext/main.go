package main

import (
	"flag"
	"log"
	"os"

	"github.com/dim13/robo"
)

var scale = flag.Float64("scale", 1.0, "font scale")

func main() {
	flag.Parse()

	dev, err := robo.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer dev.Close()

	defer robo.Home(dev.Writer)

	robo.Initialize(dev, 113, robo.Portrait)
	robo.A4.Sub(robo.Margin).UpperRight(dev)
	robo.Triple{U: 100, V: 100, W: 100}.Factor(dev)
	robo.Print(dev, os.Stdin, robo.Unit(*scale))
}

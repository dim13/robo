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

	robo.Initialize(dev.ReadWriter, 113, robo.Portrait)
	robo.A4.UpperRight(dev.Writer)
	robo.Triple{100, 100, 100}.Factor(dev.Writer)
	robo.Print(dev.Writer, os.Stdin, robo.Unit(*scale))
}

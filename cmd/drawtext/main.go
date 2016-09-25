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

	r, err := robo.NewRobo()
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()

	robo.Initialize(handle, 113, robo.Portrait)
	robo.A4.UpperRight(handle.Writer)
	robo.Triple{100, 100, 100}.Factor(handle.Writer)
	robo.Print(handle.Writer, os.Stdin, robo.Unit(*scale))
	r.Home()
}

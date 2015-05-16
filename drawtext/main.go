package main

import (
	"flag"

	"dim13.org/robo"
)

var scale = flag.Float64("scale", 1.0, "font scale")

func main() {
	flag.Parse()

	dev := robo.NewDevice()
	defer dev.Close()

	handle := dev.Handle()
	defer robo.Home(handle.Writer)

	robo.Initialize(handle, 113, robo.Portrait)
	robo.A4.UpperRight(handle.Writer)
	robo.Triple{100, 100, 100}.Factor(handle.Writer)
	robo.PrintStdin(handle.Writer, robo.Unit(*scale))
}

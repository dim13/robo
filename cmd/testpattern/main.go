package main

import "github.com/dim13/robo"

func main() {
	dev := robo.NewDevice()
	defer dev.Close()

	handle := dev.Handle()
	defer robo.Home(handle.Writer)

	robo.Initialize(handle, 113, robo.Portrait)
	robo.A4.UpperRight(handle.Writer)
	robo.TestPattern(handle.Writer)
}

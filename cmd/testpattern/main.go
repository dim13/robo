package main

import "dim13.org/robo"

func main() {
	dev := robo.NewDevice()
	defer dev.Close()

	handle := dev.Handle()
	defer robo.Home(handle.Writer)

	robo.Initialize(handle, 113, robo.Portrait)
	robo.A4.UpperRight(handle.Writer)
	robo.TestPattern(handle.Writer)
}

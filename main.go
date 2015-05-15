package main

func main() {
	dev := NewDevice()
	defer dev.Close()

	handle := dev.Handle()
	defer Home(handle.Writer)

	Initialize(handle, 113, Portrait)
	A4.UpperRight(handle.Writer)
	PrintStdin(handle.Writer)
	//DrawPic(handle.Writer)
}

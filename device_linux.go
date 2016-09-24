package robo

func NewDevice() (Device, error) {
	return NewLP("/dev/usb/lp0")
}

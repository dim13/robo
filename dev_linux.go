package robo

func Open() (Device, error) {
	return NewLP("/dev/usb/lp0")
}

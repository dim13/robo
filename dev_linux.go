package robo

const LPDevice = "/dev/usb/lp0"

func Open() (Device, error) {
	return NewLP(LPDevice)
}

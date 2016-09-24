package robo

func Open() (Device, error) {
	dev, err := NewLP("/dev/usb/lp0")
	if err != nil {
		return Device{}, nil
	}
	return Device{dev}
}

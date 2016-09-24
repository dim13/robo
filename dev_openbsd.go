package robo

func Open() (Device, error) {
	// note: ulpt* doesn't support read(), thus this dev is broken atm.
	dev, err := NewLP("/dev/ulpt0")
	if err != nil {
		return Device{}, nil
	}
	return Device{dev}
}

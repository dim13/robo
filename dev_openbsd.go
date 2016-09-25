package robo

func Open() (Device, error) {
	// note: ulpt* doesn't support read(), thus this dev is broken atm.
	return NewLP("/dev/ulpt0")
}

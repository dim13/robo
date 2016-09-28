package robo

const LPDevice = "/dev/ulpt0"

func Open() (Device, error) {
	// note: ulpt* doesn't support read(), thus this dev is broken atm.
	return NewLP(LPDevice)
}

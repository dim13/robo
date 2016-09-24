package robo

func Open() (Device, error) {
	return NewUSB()
}

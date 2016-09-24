package robo

func NewDevice() (Device, error) {
	return NewUSB()
}

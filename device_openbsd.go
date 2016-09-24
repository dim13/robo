// note: ulpt* doesn't support read(), thus this dev is broken atm.
package robo

func NewDevice() (Device, error) {
	return NewLP("/dev/ulpt0")
}

// note: ulpt* doesn't support read(), thus this dev is broken atm.
// use instead ugen* by disabling ulpt at boot time
// see config(8) for details
package robo

import "log"

func NewDevice() Device {
	dev, err := NewLP("/dev/ugen0.00")
	if err != nil {
		log.Fatal(err)
	}
	return dev
}

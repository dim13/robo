package robo

import "io"

func Open() (io.ReadWriteCloser, error) {
	// note: ulpt* doesn't support read(), thus this dev is broken atm.
	return NewLP("/dev/ulpt0")
}

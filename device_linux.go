package robo

import (
	"io"
	"os"
)

type USB struct {
	io.ReadWriteCloser
}

const (
	graphtec            = 0x0b4d
	craftrobo           = 0x110a
	craftrobolite       = 0x111a
	silhouette          = 0x111c
	silhouette_sd       = 0x111d
	silhouette_cameo    = 0x1121
	silhouette_portrait = 0x1123
)

func Open() (USB, error) {
	fs, err := os.Open("/dev/usb/lp0")
	if err != nil {
		return USB{}, err
	}
	return USB{
		ReadWriteCloser: fs,
	}, nil
}

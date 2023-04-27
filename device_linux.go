package robo

import (
	"bufio"
	"os"
)

type USB struct {
	done func() error
	*bufio.ReadWriter
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
		ReadWriter: bufio.NewReadWriter(bufio.NewReader(fs), bufio.NewWriter(fs)),
		done:       fs.Close,
	}, nil
}

func (d USB) Close() error {
	d.ReadWriter.Flush()
	d.done()
	return nil
}

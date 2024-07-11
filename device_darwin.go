package robo

import (
	"io"

	"github.com/google/gousb"
)

type USB struct {
	ctx  *gousb.Context
	dev  *gousb.Device
	intf *gousb.Interface
	io.Reader
	io.Writer
	done func()
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
	ctx := gousb.NewContext()
	dev, err := ctx.OpenDeviceWithVIDPID(graphtec, craftrobo)
	if err != nil {
		return USB{}, err
	}
	intf, done, err := dev.DefaultInterface()
	if err != nil {
		return USB{}, err
	}
	in, err := intf.InEndpoint(2)
	if err != nil {
		return USB{}, err
	}
	out, err := intf.OutEndpoint(1)
	if err != nil {
		return USB{}, err
	}
	return USB{
		ctx:    ctx,
		dev:    dev,
		intf:   intf,
		done:   done,
		Reader: in,
		Writer: out,
	}, nil
}

func (d USB) Close() error {
	d.done()
	d.dev.Close()
	d.ctx.Close()
	return nil
}

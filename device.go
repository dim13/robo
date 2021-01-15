package robo

import (
	"bufio"

	"github.com/google/gousb"
)

type USB struct {
	ctx  *gousb.Context
	dev  *gousb.Device
	intf *gousb.Interface
	*bufio.ReadWriter
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

func match(desc *gousb.DeviceDesc) bool {
	if desc.Vendor == graphtec {
		switch desc.Product {
		case craftrobo, craftrobolite,
			silhouette, silhouette_sd,
			silhouette_cameo, silhouette_portrait:
			return true
		}
	}
	return false
}

func NewDevice() (USB, error) {
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
		ctx:        ctx,
		dev:        dev,
		intf:       intf,
		done:       done,
		ReadWriter: bufio.NewReadWriter(bufio.NewReader(in), bufio.NewWriter(out)),
	}, nil
}

func (d USB) Close() error {
	d.ReadWriter.Flush()
	d.done()
	d.dev.Close()
	d.ctx.Close()
	return nil
}

package robo

import (
	"bufio"
	"errors"
	"log"

	"github.com/google/gousb"
)

type USB struct {
	ctx  *gousb.Context
	dev  *gousb.Device
	done func()
	intf *gousb.Interface
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

func NewUSB() (USB, error) {
	ctx := gousb.NewContext()
	devs, err := ctx.OpenDevices(match)
	if err != nil {
		log.Fatal(err)
	}
	if len(devs) != 1 {
		for _, dev := range devs {
			dev.Close()
		}
		return USB{}, errors.New("Cannot find Craft ROBO")
	}
	intf, done, err := devs[0].DefaultInterface()
	return USB{ctx, devs[0], done, intf}, nil
}

func (d USB) Close() {
	d.done()
	d.dev.Close()
	d.ctx.Close()
}

func (d USB) Handle() *bufio.ReadWriter {
	in, err := d.intf.InEndpoint(2)
	if err != nil {
		panic(err)
	}
	out, err := d.intf.OutEndpoint(1)
	if err != nil {
		panic(err)
	}
	return bufio.NewReadWriter(bufio.NewReader(in), bufio.NewWriter(out))
}

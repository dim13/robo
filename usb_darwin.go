package robo

import (
	"bufio"
	"errors"
	"log"

	"github.com/kylelemons/gousb/usb"
)

type USB struct {
	ctx *usb.Context
	dev *usb.Device
}

var (
	graphtec            = usb.ID(0x0b4d)
	craftrobo           = usb.ID(0x110a)
	craftrobolite       = usb.ID(0x111a)
	silhouette          = usb.ID(0x111c)
	silhouette_sd       = usb.ID(0x111d)
	silhouette_cameo    = usb.ID(0x1121)
	silhouette_portrait = usb.ID(0x1123)
	debug               = 3
)

func init() {
	// bump timeouts
	usb.DefaultReadTimeout *= 60
	usb.DefaultWriteTimeout *= 300
}

func match(desc *usb.Descriptor) bool {
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
	ctx := usb.NewContext()
	ctx.Debug(debug)
	devs, err := ctx.ListDevices(match)
	if err != nil {
		log.Fatal(err)
	}
	if len(devs) != 1 {
		for _, dev := range devs {
			dev.Close()
		}
		return USB{}, errors.New("Cannot find Craft ROBO")
	}
	return USB{ctx, devs[0]}, nil
}

func (d USB) Close() {
	d.dev.Close()
	d.ctx.Close()
}

func (d USB) Handle() *bufio.ReadWriter {
	var (
		r *bufio.Reader
		w *bufio.Writer
	)

	for _, c := range d.dev.Configs {
		for _, i := range c.Interfaces {
			for _, s := range i.Setups {
				for _, ep := range s.Endpoints {
					e, err := d.dev.OpenEndpoint(
						c.Config,
						i.Number,
						s.Number,
						ep.Address)
					if err != nil {
						log.Fatal(err)
					}
					switch ep.Direction() {
					case usb.ENDPOINT_DIR_OUT:
						w = bufio.NewWriter(e)
					case usb.ENDPOINT_DIR_IN:
						r = bufio.NewReader(e)
					}
				}
			}
		}
	}

	return bufio.NewReadWriter(r, w)
}

package main

import (
	"bufio"
	"log"

	"github.com/kylelemons/gousb/usb"
)

type Device struct {
	ctx *usb.Context
	dev *usb.Device
}

var (
	vendor  = usb.ID(0x0b4d)
	product = usb.ID(0x110a)
	debug   = 3
)

func CC100(desc *usb.Descriptor) bool {
	return desc.Vendor == vendor && desc.Product == product
}

func NewDevice() (d Device) {
	ctx := usb.NewContext()
	ctx.Debug(debug)
	devs, err := ctx.ListDevices(CC100)
	if err != nil {
		log.Fatal(err)
	}
	if len(devs) != 1 {
		for _, dev := range devs {
			dev.Close()
		}
		log.Fatal("wrong number of devices")
	}
	return Device{ctx: ctx, dev: devs[0]}
}

func (d Device) Close() {
	d.dev.Close()
	d.ctx.Close()
}

func (d Device) Handle() *bufio.ReadWriter {
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

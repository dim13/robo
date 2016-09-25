package robo

import (
	"bufio"
	"errors"
	"io"

	"github.com/kylelemons/gousb/usb"
)

func Open() (Device, error) {
	return NewUSB()
}

type USB struct {
	ctx *usb.Context
	dev *usb.Device
	io.Writer
	io.Reader
}

var (
	graphtec            = usb.ID(0x0b4d)
	craftrobo           = usb.ID(0x110a)
	craftrobolite       = usb.ID(0x111a)
	silhouette          = usb.ID(0x111c)
	silhouette_sd       = usb.ID(0x111d)
	silhouette_cameo    = usb.ID(0x1121)
	silhouette_portrait = usb.ID(0x1123)
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
	u := USB{}
	u.ctx = usb.NewContext()
	devs, err := u.ctx.ListDevices(match)
	if err != nil {
		return USB{}, err
	}
	if len(devs) != 1 {
		for _, dev := range devs {
			dev.Close()
		}
		return USB{}, errors.New("Cannot find Craft ROBO")
	}
	u.dev = devs[0]

	for _, c := range u.dev.Configs {
		for _, i := range c.Interfaces {
			for _, s := range i.Setups {
				for _, ep := range s.Endpoints {
					e, err := u.dev.OpenEndpoint(
						c.Config,
						i.Number,
						s.Number,
						ep.Address)
					if err != nil {
						return USB{}, err
					}
					switch ep.Direction() {
					case usb.ENDPOINT_DIR_OUT:
						u.Writer = e
					case usb.ENDPOINT_DIR_IN:
						u.Reader = e
					}
				}
			}
		}
	}

	return u, nil
}

func (d USB) Close() error {
	d.dev.Close()
	d.ctx.Close()
	return nil
}

// ReadString reads until End of Text
func (d USB) ReadString() (string, error) {
	buf := bufio.NewReader(d.Reader)
	resp, err := buf.ReadString(ETX)
	if err != nil {
		return "", err
	}
	return resp[:len(resp)-1], nil
}

// WriteString terminates transfer with End of Text
func (d USB) WriteString(s string) error {
	buf := bufio.NewWriter(d.Writer)
	buf.WriteString(s)
	buf.WriteByte(ETX)
	return buf.Flush()
}

// Command prefixes transfer with Escape
func (d USB) Command(b []byte) error {
	buf := bufio.NewWriter(d.Writer)
	buf.WriteByte(ESC)
	buf.Write(b)
	return buf.Flush()
}

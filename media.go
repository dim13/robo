package robo

import (
	"fmt"
	"io"
)

type Media struct {
	Descr        string // Description
	ID           int    // FW111
	Speed        int    // !10
	Force        int    // FX10,0
	CutterOffset int    // FC18, depends on blade, 0 for pens
}

var MediaID = make(map[int]Media)

func init() {
	for _, m := range Medias {
		MediaID[m.ID] = m
	}
}

func (m Media) String() string {
	return fmt.Sprintf("%v: Speed %2d, Force %2d %s",
		m.ID, m.Speed, m.Force, m.Descr)
}

func (m Media) Apply(c io.Writer) {
	Unit(m.ID).Media(c)
	Unit(m.Speed).Speed(c)
	Unit(m.Force).Force(c)
	Unit(m.CutterOffset).CutterOffset(c)
}

var Medias = []Media{
	{
		Descr:        "Card without Craft Paper Backing",
		ID:           100,
		Speed:        10,
		Force:        27,
		CutterOffset: 18,
	},
	{
		Descr:        "Card with Craft Paper Backing",
		ID:           101,
		Speed:        10,
		Force:        27,
		CutterOffset: 18,
	},
	{
		Descr:        "Vinyl Sticker",
		ID:           102,
		Speed:        10,
		Force:        10,
		CutterOffset: 18,
	},
	{
		Descr:        "Film Labels",
		ID:           106,
		Speed:        10,
		Force:        14,
		CutterOffset: 18,
	},
	{
		Descr:        "Magnetic Sheet",
		ID:           107,
		Speed:        10,
		Force:        12,
		CutterOffset: 18,
	},
	{
		Descr:        "Thick Media",
		ID:           111,
		Speed:        10,
		Force:        27,
		CutterOffset: 18,
	},
	{
		Descr:        "Thin Media",
		ID:           112,
		Speed:        10,
		Force:        2,
		CutterOffset: 18,
	},
	{
		Descr:        "Pen",
		ID:           113,
		Speed:        10,
		Force:        10,
		CutterOffset: 0,
	},
	{
		Descr:        "Custom",
		ID:           300,
		Speed:        10,
		Force:        10,
		CutterOffset: 18,
	},
}

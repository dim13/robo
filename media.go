package main

import (
	"bufio"
	"fmt"
)

// Overcut: depends on blade, 0 for pens

type Media struct {
	Descr   string // Description
	ID      int    // FW111
	Speed   int    // !10
	Force   int    // FX10,0
	Overcut int    // FC ?
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

func (m Media) Apply(c *bufio.Writer) {
	Unit(m.ID).Media(c)
	Unit(m.Speed).Speed(c)
	Unit(m.Force).Force(c)
	Unit(m.Overcut).Overcut(c)
}

var Medias = []Media{
	Media{
		Descr:   "Card without Craft Paper Backing",
		ID:      100,
		Speed:   10,
		Force:   27,
		Overcut: 18,
	},
	Media{
		Descr:   "Card with Craft Paper Backing",
		ID:      101,
		Speed:   10,
		Force:   27,
		Overcut: 18,
	},
	Media{
		Descr:   "Vinyl Sticker",
		ID:      102,
		Speed:   10,
		Force:   10,
		Overcut: 18,
	},
	Media{
		Descr:   "Film Labels",
		ID:      106,
		Speed:   10,
		Force:   14,
		Overcut: 18,
	},
	Media{
		Descr:   "Magnetic Sheet",
		ID:      107,
		Speed:   10,
		Force:   12,
		Overcut: 18,
	},
	Media{
		Descr:   "Thick Media",
		ID:      111,
		Speed:   10,
		Force:   27,
		Overcut: 18,
	},
	Media{
		Descr:   "Thin Media",
		ID:      112,
		Speed:   10,
		Force:   2,
		Overcut: 18,
	},
	Media{
		Descr:   "Pen",
		ID:      113,
		Speed:   10,
		Force:   10,
		Overcut: 0,
	},
	Media{
		Descr:   "Custom",
		ID:      300,
		Speed:   10,
		Force:   10,
		Overcut: 18,
	},
}

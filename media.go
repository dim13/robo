package main

import "fmt"

type Media struct {
	Descr     string // Description
	ID        int    // FW111
	Speed     int    // !10
	Thickness int    // FX10,0
	FC        int    // FC ?
}

var MediaMap = make(map[int]Media)

func init() {
	for _, m := range Medias {
		MediaMap[m.ID] = m
	}
}

func (m Media) String() string {
	return fmt.Sprintf("%v: Speed %2d, Thickness %2d %s",
		m.ID, m.Speed, m.Thickness, m.Descr)
}

var Medias = []Media{
	Media{
		Descr:     "Card without Craft Paper Backing",
		ID:        100,
		Speed:     10,
		Thickness: 27,
		FC:        18,
	},
	Media{
		Descr:     "Card with Craft Paper Backing",
		ID:        101,
		Speed:     10,
		Thickness: 27,
		FC:        18,
	},
	Media{
		Descr:     "Vinyl Sticker",
		ID:        102,
		Speed:     10,
		Thickness: 10,
		FC:        18,
	},
	Media{
		Descr:     "Film Labels",
		ID:        106,
		Speed:     10,
		Thickness: 14,
		FC:        18,
	},
	Media{
		Descr:     "Magnetic Sheet",
		ID:        107,
		Speed:     10,
		Thickness: 12,
		FC:        18,
	},
	Media{
		Descr:     "Thick Media",
		ID:        111,
		Speed:     10,
		Thickness: 27,
		FC:        18,
	},
	Media{
		Descr:     "Thin Media",
		ID:        112,
		Speed:     10,
		Thickness: 2,
		FC:        18,
	},
	Media{
		Descr:     "Pen",
		ID:        113,
		Speed:     10,
		Thickness: 10,
		FC:        0,
	},
	Media{
		Descr:     "Custom",
		ID:        300,
		Speed:     10,
		Thickness: 10,
		FC:        18,
	},
}

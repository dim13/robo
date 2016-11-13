package robo

// Overcut: depends on blade, 0 for pens

type Media struct {
	Descr   string // Description
	ID      Unit   // FW111
	Speed   Unit   // !10
	Force   Unit   // FX10,0
	Overcut Unit   // FC ?
}

func (r Robo) SetMedia(m Media) {
	r.Media(m.ID)
	r.Speed(m.Speed)
	r.Force(m.Force)
	r.Overcut(m.Overcut)
}

var (
	MediaCardWithoutCraftPaperBacking = Media{
		Descr:   "Card without Craft Paper Backing",
		ID:      100,
		Speed:   10,
		Force:   27,
		Overcut: 18,
	}
	MediaCardWithCraftPaperBacking = Media{
		Descr:   "Card with Craft Paper Backing",
		ID:      101,
		Speed:   10,
		Force:   27,
		Overcut: 18,
	}
	MediaVinylSticker = Media{
		Descr:   "Vinyl Sticker",
		ID:      102,
		Speed:   10,
		Force:   10,
		Overcut: 18,
	}
	MediaFilmLables = Media{
		Descr:   "Film Labels",
		ID:      106,
		Speed:   10,
		Force:   14,
		Overcut: 18,
	}
	MediaMagneticSheet = Media{
		Descr:   "Magnetic Sheet",
		ID:      107,
		Speed:   10,
		Force:   12,
		Overcut: 18,
	}
	MediaThick = Media{
		Descr:   "Thick Media",
		ID:      111,
		Speed:   10,
		Force:   27,
		Overcut: 18,
	}
	MediaThin = Media{
		Descr:   "Thin Media",
		ID:      112,
		Speed:   10,
		Force:   2,
		Overcut: 18,
	}
	MediaPen = Media{
		Descr:   "Pen",
		ID:      113,
		Speed:   10,
		Force:   10,
		Overcut: 0,
	}
	MediaCustom = Media{
		Descr:   "Custom",
		ID:      300,
		Speed:   10,
		Force:   10,
		Overcut: 18,
	}
)

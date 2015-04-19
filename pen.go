package main

type Pen struct {
	Speed int
	Force int
	Cap   string
}

var pens = map[string]Pen{
	"pen": Pen{
		Speed: 10,
		Force: 10,
		Cap:   "pen",
	},
	"thin": Pen{
		Speed: 10,
		Force: 2,
		Cap:   "blue",
	},
	"thick": Pen{
		Speed: 10,
		Force: 27,
		Cap:   "yellow",
	},
}

package main

import "log"

func (c Cutter) DrawAtom() {
	base := Point{2000, 2000}
	for i := 0; i < 3; i++ {
		c.Ellipse(0, base,
			Polar{500, 0}, Polar{200, 3600},
			600*float64(i))
	}
	c.Circle(base, Polar{100, 0}, Polar{100, 3600})
}

func (c Cutter) DrawLines() {
	for i := 0; i < 9; i++ {
		c.LineType(LineStyle(i))
		c.Move(Point{100 * float64(i), 0})
		c.Draw(Point{100 * float64(i), 1000})
	}
}

func (c Cutter) DrawCircles() {
	base := Point{1000, 1000}
	for i := 1; i <= 3; i++ {
		c.Circle(base,
			Polar{100 * float64(i), 0},
			Polar{100 * float64(i), 3600})
	}
}

func (c Cutter) MustMarks(p Point) {
	if !c.SearchMarks(p) {
		log.Fatal("marks not found")
	}
}

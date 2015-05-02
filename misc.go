package main

func (c Cutter) DrawAtom() {
	for i := 0; i < 3; i++ {
		c.Ellipse(0, Point{2000, 2000},
			Polar{500, 0}, Polar{200, 3600}, 600*i)
	}
	c.Circle(Point{2000, 2000}, Polar{100, 0}, Polar{100, 3600})
}

func (c Cutter) DrawLines() {
	for i := 0; i < 9; i++ {
		c.LineType(LineStyle(i))
		c.Move(Point{100 * i, 0})
		c.Draw(Point{100 * i, 1000})
	}
}

func (c Cutter) DrawCircles() {
	for i := 1; i < 10; i++ {
		c.Circle(Point{1000, 1000},
			Polar{100 * i, 0},
			Polar{100 * i, 3600})
	}
	c.Move(Point{0, 1000})
	c.Draw(Point{2000, 1000})
	c.Move(Point{1000, 0})
	c.Draw(Point{1000, 2000})
}

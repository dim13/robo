package main

func (c Cutter) TestPattern() {
	c.Factor(100)
	c.Offset(Origin)
	c.WriteLowerLeft(Origin)

	c.Move(Point{510, 637})
	c.Bezier(1, Point{510, 637}, Point{439, 637},
		Point{383, 580}, Point{383, 510})
	c.Bezier(1, Point{383, 510}, Point{383, 439},
		Point{439, 383}, Point{510, 383})
	c.Bezier(1, Point{510, 383}, Point{580, 383},
		Point{637, 439}, Point{637, 510})
	c.Bezier(1, Point{637, 510}, Point{637, 580},
		Point{580, 637}, Point{510, 637})
	c.Move(Point{764, 764})
	c.Draw(Point{256, 764})
	c.Draw(Point{256, 256})
	c.Draw(Point{764, 256})
	c.Draw(Point{764, 764})
	c.Move(Point{2, 510})
	c.Draw(Point{1018, 510})
	c.Move(Point{510, 1018})
	c.Draw(Point{510, 2})
	c.Move(Origin)
}

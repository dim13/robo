package robo

func (r Robo) TestPattern() {
	r.Factor(100, 100, 100)
	r.Offset(Point{0, 0})
	r.LowerLeft(Point{0, 0})
	r.Move(Point{510, 637})
	r.Bezier(Point{510, 637}, Point{439, 637}, Point{383, 580}, Point{383, 510})
	r.Bezier(Point{383, 510}, Point{383, 439}, Point{439, 383}, Point{510, 383})
	r.Bezier(Point{510, 383}, Point{580, 383}, Point{637, 439}, Point{637, 510})
	r.Bezier(Point{637, 510}, Point{637, 580}, Point{580, 637}, Point{510, 637})
	r.Line(Point{764, 764}, Point{256, 764}, Point{256, 256}, Point{764, 256}, Point{764, 764})
	r.Line(Point{2, 510}, Point{1018, 510})
	r.Line(Point{510, 1018}, Point{510, 2})
	r.Move(Point{0, 0})
}

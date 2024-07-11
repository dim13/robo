package robo

import "io"

func TestPattern(c io.Writer) {
	Triple{100, 100, 100}.Factor(c)
	Point{0, 0}.Offset(c)
	Point{0, 0}.LowerLeft(c)
	Point{510, 637}.Move(c)
	Path{
		Point{510, 637}, Point{439, 637},
		Point{383, 580}, Point{383, 510},
	}.Bezier(c, 1)
	Path{
		Point{383, 510}, Point{383, 439},
		Point{439, 383}, Point{510, 383},
	}.Bezier(c, 1)
	Path{
		Point{510, 383}, Point{580, 383},
		Point{637, 439}, Point{637, 510},
	}.Bezier(c, 1)
	Path{
		Point{637, 510}, Point{637, 580},
		Point{580, 637}, Point{510, 637},
	}.Bezier(c, 1)
	Path{
		Point{764, 764}, Point{256, 764},
		Point{256, 256}, Point{764, 256},
		Point{764, 764},
	}.Line(c)
	Path{
		Point{2, 510}, Point{1018, 510},
	}.Line(c)
	Path{
		Point{510, 1018}, Point{510, 2},
	}.Line(c)
	Point{0, 0}.Move(c)
}

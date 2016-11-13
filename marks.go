package robo

import "bufio"

/*
	Landscape		Portrait
	┌─  H  ─┐		┌─  W  ─┐
	│x1    3│		│2    1x│

	W   →			    ↓   H

	│2			       3│
	└─			       ─┘
*/

/*
	Type1			Type2

	 │     │
	─┘     └─		┌─     ─┐
				│       │

				│
	─┐			└─
	 │
*/

func DrawMarks(c *bufio.Writer, offset, size Point, length int) {
	Point{600, 3800}.Move(c)
	Point{200, 3800}.Draw(c)
	Point{200, 3400}.Draw(c)

	Point{200, 600}.Move(c)
	Point{200, 200}.Draw(c)
	Point{600, 200}.Draw(c)

	Point{4840, 200}.Move(c)
	Point{5240, 200}.Draw(c)
	Point{5240, 600}.Draw(c)
}

func (r Robo) DrawMarks() {
	r.Line(Point{600, 3800}, Point{200, 3800}, Point{200, 3400})
	r.Line(Point{200, 600}, Point{200, 200}, Point{600, 200})
	r.Line(Point{4840, 200}, Point{5240, 200}, Point{5240, 600})
}

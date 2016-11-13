package robo

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

func (r Robo) DrawMarks() {
	r.Line(Point{600, 3800}, Point{200, 3800}, Point{200, 3400})
	r.Line(Point{200, 600}, Point{200, 200}, Point{600, 200})
	r.Line(Point{4840, 200}, Point{5240, 200}, Point{5240, 600})
}

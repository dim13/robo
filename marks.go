package main

/*	Landscape		Portrait
	+- H -+			+- W -+
	|x1  3|			|2  1x|

	W ->			   |  H
				   v
	|2			     3|
	+-			     -+
*/

/*
	Type1

	 |	|
	-+	+-


	-+
	 |

	Type2

	+-	-+
	|	 |


	|
	+-
*/

func (c Cutter) DrawMarks(offset, size Point, length int) {
	c.Move(Point{600, 3800})
	c.Draw(Point{200, 3800})
	c.Draw(Point{200, 3400})

	c.Move(Point{200, 600})
	c.Draw(Point{200, 200})
	c.Draw(Point{600, 200})

	c.Move(Point{4840, 200})
	c.Draw(Point{5240, 200})
	c.Draw(Point{5240, 600})
}

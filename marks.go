package main

/*	Portrait
	+--
	| 3


	| 1  2 |
	+--  --+
	Offset:  10mm (200 pt)
	Length:  20mm (400 pt)
	Width:  190mm (3800 pt)
	Height: 262mm (5240 pt)
*/

/*	Landscape
	+--        +
	| 3

	| 1      2 |
	+--      --+
	1-2 262 mm
	1-3 190 mm
	offset 10x10 mm
	length 20 mm
*/

/*	Landscape		Portrait
	+- H -+			+- W -+
	|x1  3|			|2  1x|

	W ->			   |  H
				   v
	|2			     3|
	+-			     -+
*/

func (c Cutter) DrawMarks() (string, error) {
	c.Move(Point{600, 3800})
	c.Draw(Point{200, 3800})
	c.Draw(Point{200, 3400})

	c.Move(Point{200, 600})
	c.Draw(Point{200, 200})
	c.Draw(Point{600, 200})

	c.Move(Point{4840, 200})
	c.Draw(Point{5240, 200})
	c.Draw(Point{5240, 600})

	return c.readResponse()
}

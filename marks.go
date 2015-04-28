package main

/*
	Marks
	+--    --+
	|        |



	|
	+--

	Offset:  10mm (200 pt)
	Length:  20mm (400 pt)
	Width:  190mm (3800 pt)
	Height: 262mm (5240 pt)
*/

// 5440 x 4000
// Origin: 200x200

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

func (c Cutter) SearchMarks() {
	//cmd := []string{"TB99", "TB55,1", "TB123,3800,5240"}
	cmd := []string{"TB99", "TB55,1", "TB23,3800,5240"}
	c.Raw(cmd)
}

package main

func main() {
	dev := NewDevice()
	defer dev.Close()

	cu := NewCutter(dev.Handle())
	p := pens["pen"]

	//cu.Version()
	cu.Orientation(Portrait)
	cu.WriteUpperRight(A4)
	cu.Speed(p.Speed)
	cu.Force(p.Force)

	/*
		cu.TestPattern()
		cu.Draw(Point{4000,0})
		cu.Draw(Point{4000,4000})
		cu.Draw(Point{0,4000})
		cu.Draw(Point{0,0})
	*/

	/*
		for i := 0; i < 5; i++ {
			cu.Move(Point{1000*i,0})
			cu.Draw(Point{1000*i,4000})

			cu.Move(Point{0, 1000*i})
			cu.Draw(Point{4000, 1000*i})
		}
	*/

	defer cu.Home()
	defer cu.LineType(Solid)

	for i := 0; i < 9; i++ {
		cu.LineType(LineStyle(i))
		cu.Move(Point{100 * i, 0})
		cu.Draw(Point{100 * i, 1000})
	}

}

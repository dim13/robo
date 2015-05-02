package main

func main() {
	dev := NewDevice()
	defer dev.Close()

	cu := NewCutter(dev.Handle(), Portrait)

	defer cu.Home()
	//defer cu.LineType(Solid)
	//cu.Raw([]string{"L100,1,400,100"})

	cu.TestCut()
	//cu.TestPattern()
	//cu.EasterEgg()
	//cu.DrawMarks()
	//cu.Move(Point{300, 300})
	//cu.SearchMarks()
	//cu.Bezier(1, Point{0,0},Point{0,1000},Point{0,0},Point{1000,0})

	/*
		for i := 0; i < 9; i++ {
			cu.LineType(LineStyle(i))
			cu.Move(Point{100 * i, 0})
			cu.Draw(Point{100 * i, 1000})
		}
	*/
}

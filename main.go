package main

func main() {
	dev := NewDevice()
	defer dev.Close()

	cu := NewCutter(dev.Handle(), Portrait)

	defer cu.Home()
	//defer cu.LineType(Solid)
	//cu.Raw([]string{"L100,1,400,100"})

	//cu.TestCut()
	//cu.TestPattern()
	//cu.EasterEgg()
	//cu.DrawMarks()
	//cu.Move(Point{300, 300})
	//cu.SearchMarks()
	//cu.Bezier(1, Point{0,0},Point{0,1000},Point{0,0},Point{1000,0})

	/*
		for i := 1; i < 10; i++ {
			cu.Circle(Point{1000, 1000},
				Polar{100 * i, 0},
				Polar{100 * i, 3600})
		}
	*/

	//cu.Move(Point{500,500})

	cu.Circle(Point{2000, 2000}, Polar{100, 0}, Polar{100, 3600})
	for i := 0; i < 3; i++ {
		cu.Ellipse(0, Point{2000, 2000},
			Polar{500, 0}, Polar{200, 3600}, 600*i)
	}

	/*
		for i := 0; i < 9; i++ {
			cu.LineType(LineStyle(i))
			cu.Move(Point{100 * i, 0})
			cu.Draw(Point{100 * i, 1000})
		}
	*/
}

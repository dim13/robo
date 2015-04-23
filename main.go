package main

import "fmt"

func main() {
	dev := NewDevice()
	defer dev.Close()

	cu := NewCutter(dev.Handle())
	p := pens["pen"]

	v, _ := cu.Version()
	fmt.Println("Version:", v)

	cu.Orientation(Portrait)
	cu.WriteUpperRight(A4)
	cu.Speed(p.Speed)
	cu.Force(p.Force)

	defer cu.Home()
	defer cu.LineType(Solid)

	cu.Orientation(Landscape)
	cu.TestCut()
	//cu.TestPattern()
	cu.Move(Point{1000, 1000})
	cu.Home()

	/*
		for i := 0; i < 9; i++ {
			cu.LineType(LineStyle(i))
			cu.Move(Point{100 * i, 0})
			cu.Draw(Point{100 * i, 1000})
		}
	*/
}

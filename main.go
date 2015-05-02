package main

import "fmt"

func main() {
	dev := NewDevice()
	defer dev.Close()

	c := NewCutter(dev.Handle(), Portrait)

	defer c.Home()
	//defer c.LineType(Solid)
	//c.Raw([]string{"L100,1,400,100"})

	//c.TestCut()
	//c.TestPattern()
	//c.EasterEgg()
	//c.DrawMarks(Point{}, Point{}, 0)
	//c.WriteUpperRight(Point{4000,5440})
	fmt.Println(c.SearchMarks(Point{5240, 3800}, 400))
	//c.Bezier(1, Point{0,0},Point{0,1000},Point{0,0},Point{1000,0})
}

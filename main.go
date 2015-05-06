package main

import (
	"fmt"
	"log"
)

func main() {
	//dev, err := NewDevice()
	dev, err := NewLPDevice("/dev/usb/lp0")
	if err != nil {
		log.Fatal(err)
	}
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
	//fmt.Println(c.UnknownFQ5())
	//c.Bezier(1, Point{0,0},Point{0,1000},Point{0,0},Point{1000,0})
	//c.DrawCircles()

	fmt.Println("Gin", c.Gin())
	fmt.Println("Call Gin", c.CallGin())

	fmt.Println("Offset", c.ReadOffset())
	fmt.Println("Upper Right", c.ReadUpperRight())
	fmt.Println("Lower Left", c.ReadLowerLeft())
	//fmt.Println(c.StatusWord())

	if c.SearchMarks(Point{19 * CM, 18 * CM}, 2*CM) {
		fmt.Println("Reg Marks ok")
	}

}

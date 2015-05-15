package main

import (
	"flag"
	"log"
	"runtime"
)

var (
	dev Devicer
	err error
	cmd = flag.String("cmd", "", "command")
)

func main() {
	flag.Parse()

	if runtime.GOOS == "linux" {
		dev, err = NewLP("/dev/usb/lp0")
	} else {
		dev, err = NewUSB()
	}

	if err != nil {
		log.Fatal(err)
	}

	defer dev.Close()

	handle := dev.Handle()

	//c := NewCutter(handle, Portrait, 0)

	defer Home(handle.Writer)
	//defer c.LineType(Solid)
	//c.Raw([]string{"L100,1,400,100"})

	//c.TestCut()
	//c.TestPattern()
	//c.EasterEgg()
	//c.DrawMarks(Point{}, Point{}, 0)
	//c.WriteUpperRight(Point{4000,5440})
	//fmt.Println(c.UnknownFQ5())
	//c.Bezier(1, Point{0,0},Point{0,1000},Point{0,0},Point{1000,0})

	//fmt.Println("Gin", c.Gin())
	//fmt.Println("Call Gin", c.CallGin())

	//c.MustMarks(Point{18 * CM, 19 * CM}, Type2)

	Initialize(handle, 113, Portrait)
	A4.UpperRight(handle.Writer)
	PrintStdin(handle.Writer)
	//DrawPic(handle.Writer)
}

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
		dev, err = NewLPDevice("/dev/usb/lp0")
	} else {
		dev, err = NewDevice()
	}

	if err != nil {
		log.Fatal(err)
	}

	defer dev.Close()

	c := NewCutter(dev.Handle(), Portrait, 0)

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

	//fmt.Println("Gin", c.Gin())
	//fmt.Println("Call Gin", c.CallGin())

	//c.MustMarks(Point{18 * CM, 19 * CM}, Type2)
	if *cmd != "" {
		c.Send(*cmd)
		log.Println(c.returnString())
	} else {
		c.DrawPic()
	}

}

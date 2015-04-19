package main

import (
	"bufio"
	"fmt"
	"log"
)

type Point struct {
	X, Y int
}

/*
	A4 Cutting area

	5mm	y	5mm
	+-----------------+
	|  		  | 5mm
	| +-------------+ |
	| |		| |
	. .		. . x
	. .		. .
	. .		. .
	| |		| |
	| +-------------+ |
	|  		  |
	|  		  | 20mm
	+-----------------+

	Default size: 20000,  4000
	1000 pt == 50 mm

	A4: 210x297 mm => 4200x5940
	Usable: 4000x5440 pt
*/

var (
	A4     = Point{5440, 4000} // Portrait
	Origin = Point{0, 0}
)

func (p Point) String() string {
	return fmt.Sprintf("%v,%v", p.X, p.Y)
}

type Cutter struct {
	*bufio.ReadWriter
}

func NewCutter(io *bufio.ReadWriter) Cutter {
	return Cutter{io}
}

func (c Cutter) EOT() {
	defer c.Flush()
	c.WriteByte(0x03)
}

func (c Cutter) Home() {
	defer c.EOT()
	c.WriteString("H")
}

func (c Cutter) Draw(p Point) {
	defer c.EOT()
	fmt.Fprint(c, "D", p)
}

func (c Cutter) Move(p Point) {
	defer c.EOT()
	fmt.Fprint(c, "M", p)
}

type LineStyle int

const (
	Solid LineStyle = iota
	Dots
	ShortDash
	Dash
	LongDash
	DashDot
	DashLongDot
	DashDoubleDot
	DashLongDoubleDot
)

func (c Cutter) LineType(n LineStyle) {
	defer c.EOT()
	fmt.Fprint(c, "L", n)
}

func (c Cutter) Factor(p, q, r int) {
	defer c.EOT()
	fmt.Fprintf(c, "&%v,%v,%v", p, q, r)
}

func (c Cutter) Offset(p Point) {
	defer c.EOT()
	fmt.Fprint(c, "^", p)
}

func (c Cutter) WriteLowerLeft(p Point) {
	defer c.EOT()
	fmt.Fprint(c, "\\", p)
}

func (c Cutter) WriteUpperRight(p Point) {
	defer c.EOT()
	fmt.Fprint(c, "Z", p)
}

func (c Cutter) Version() {
	c.WriteString("FG")
	c.Flush()
	ans, err := c.ReadString(0x03)
	if err != nil {
		log.Println(err)
	}
	log.Println(ans)
}

func (c Cutter) ReadUpperRight() {
	c.WriteString("U")
	c.Flush()
	ans, err := c.ReadString(0x03)
	if err != nil {
		log.Println(err)
	}
	log.Println(ans)
}

func (c Cutter) Speed(n int) {
	defer c.EOT()
	fmt.Fprint(c, "!", n)
}

func (c Cutter) Force(n int) {
	defer c.EOT()
	fmt.Fprint(c, "FX", n)
}

func (c Cutter) Initialize() {
	c.WriteString("\x1b\x04") // Initialize ???
	c.Flush()
}

func (c Cutter) Status() {
	c.WriteString("\x1b\x05") // Status ???
	c.Flush()
	ans, err := c.ReadString(0x03)
	if err != nil {
		log.Println(err)
	}
	switch ans[:1] {
	case "0":
		log.Println("Ready")
	case "1":
		log.Println("Moving")
	default:
		log.Println("Unknown", ans)
	}
}

func (c Cutter) Bezier(a int, p0, p1, p2, p3 Point) {
	defer c.EOT()
	fmt.Fprintf(c, "BZ%v,%v,%v,%v,%v", a, p0, p1, p2, p3)
}

type Page int

const (
	Portrait Page = iota
	Landscape
)

func (c Cutter) Orientation(l Page) {
	defer c.EOT()
	fmt.Fprint(c, "FN", l)
}

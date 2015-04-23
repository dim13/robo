package main

import (
	"bufio"
	"fmt"
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

func (c Cutter) TestCut() {
	defer c.EOT()
	c.WriteString("FH")
}

type StepDirection int

const (
	stepEnd StepDirection = 1 << iota >> 1
	StepDown
	StepUp
	StepLeft
	StepRight
)

func (c Cutter) step(dir StepDirection) {
	fmt.Fprintf(c, "\x1e\x00%c", dir)
	c.Flush()
}

func (c Cutter) Step(dir StepDirection) {
	c.step(dir)
	c.step(stepEnd)
}

// CR returns carret to home on same line
func (c Cutter) CR() {
	defer c.EOT()
	c.WriteString("TT")
}

// Home retuns carret to home position
func (c Cutter) Home() {
	defer c.EOT()
	c.WriteString("H")
}

func (c Cutter) SetOrigin() {
	defer c.EOT()
	c.WriteString("FJ")
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

func (c Cutter) readResponse() (string, error) {
	ans, err := c.ReadString(0x03)
	if err != nil {
		return "", err
	}
	return ans[:len(ans)-1], nil
}

// Version requests hardware version
func (c Cutter) Version() (string, error) {
	c.WriteString("FG")
	c.Flush()
	return c.readResponse()
}

func (c Cutter) ReadUpperRight() (string, error) {
	c.WriteString("U")
	c.Flush()
	return c.readResponse()
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

func (c Cutter) Status() (string, error) {
	c.WriteString("\x1b\x05") // Status ???
	c.Flush()
	ans, err := c.readResponse()
	if err != nil {
		return "", err
	}
	switch ans {
	case "0":
		return "Ready", nil
	case "1":
		return "Moving", nil
	default:
		return "Unknown " + ans, nil
	}
}

func (c Cutter) Bezier(a int, p0, p1, p2, p3 Point) {
	defer c.EOT()
	fmt.Fprintf(c, "BZ%v,%v,%v,%v,%v", a, p0, p1, p2, p3)
}

type Orientation int

const (
	Portrait Orientation = iota
	Landscape
)

func (c Cutter) Orientation(l Orientation) {
	defer c.EOT()
	fmt.Fprint(c, "FN", l)
}

type OnOff int

const (
	On OnOff = iota
	Off
)

// TrackEnhancing moves paper back and forth for better traction
// Not for thickness less then 19
func (c Cutter) TrackEnhancing(state OnOff) {
	c.EOT()
	fmt.Fprint(c, "FY", state)
}

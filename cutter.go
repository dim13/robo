package main

import (
	"bufio"
	"fmt"
	"time"
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

func (c Cutter) Emit() {
	defer c.Flush()
	c.WriteByte(0x03) // send End Of Text
}

func (c Cutter) TestCut() {
	defer c.Emit()
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
	fmt.Fprintf(c, "\x1b\x00%c", dir)
	c.Flush()
}

func (c Cutter) Step(dir StepDirection) {
	c.step(dir)
	c.step(stepEnd)
}

// CR returns carret to home on same line
func (c Cutter) CR() {
	defer c.Emit()
	c.WriteString("TT")
}

// Home retuns carret to home position
func (c Cutter) Home() {
	defer c.Emit()
	c.WriteString("H")
}

func (c Cutter) SetOrigin() {
	defer c.Emit()
	c.WriteString("FJ")
}

func (c Cutter) Draw(p Point) {
	defer c.Emit()
	fmt.Fprint(c, "D", p)
}

func (c Cutter) Move(p Point) {
	defer c.Emit()
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
	defer c.Emit()
	fmt.Fprint(c, "L", n)
}

func (c Cutter) LineScale(n int) {
	defer c.Emit()
	fmt.Fprint(c, "B", n)
}

func (c Cutter) Factor(p, q, r int) {
	defer c.Emit()
	fmt.Fprintf(c, "&%v,%v,%v", p, q, r)
}

func (c Cutter) Offset(p Point) {
	defer c.Emit()
	fmt.Fprint(c, "^", p)
}

func (c Cutter) WriteLowerLeft(p Point) {
	defer c.Emit()
	fmt.Fprint(c, "\\", p)
}

func (c Cutter) WriteUpperRight(p Point) {
	defer c.Emit()
	fmt.Fprint(c, "Z", p)
}

// CuttingArea ???
func (c Cutter) CuttingArea(p Point) {
	defer c.Emit()
	fmt.Fprint(c, "FU", p)
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

// MediaType (Meida ID)
func (c Cutter) MediaType(n int) {
	defer c.Emit()
	fmt.Fprint(c, "FW", n)
}

func (c Cutter) ReadUpperRight() (string, error) {
	c.WriteString("U")
	c.Flush()
	return c.readResponse()
}

// Speed 10..100 mm/s
func (c Cutter) Speed(n int) {
	if n >= 1 && n <= 10 {
		defer c.Emit()
		fmt.Fprint(c, "!", n)
	}
}

func (c Cutter) Thickness(n int) {
	if n >= 1 && n <= 30 {
		defer c.Emit()
		fmt.Fprint(c, "FX", n, ",0")
	}
}

func (c Cutter) Force(n int) {
	c.Thickness(n)
}

func (c Cutter) UnknownFC(n int) {
	defer c.Emit()
	fmt.Fprint(c, "FC", n)
}

func (c Cutter) UnknownFE(n int) {
	defer c.Emit()
	fmt.Fprint(c, "FE", n)
}

func (c Cutter) UnknownTB71() (string, error) {
	fmt.Fprint(c, "TB71")
	c.Emit()
	return c.readResponse()
}

func (c Cutter) UnknownFA() (string, error) {
	fmt.Fprint(c, "FA")
	c.Emit()
	return c.readResponse()
}

func (c Cutter) UnknownTB51() {
	defer c.Emit()
	fmt.Fprint(c, "TB51,400")
}

// Updater Version ???
func (c Cutter) UpdaterVersion() (string, error) {
	c.WriteString("\x1b\x01")
	c.Flush()
	return c.readResponse()
}

// Update starts update sequence
// Send raw S-Record data after
func (c Cutter) Update() (bool, error) {
	c.WriteString("CC1VERUP")
	c.Flush()
	ans, err := c.readResponse()
	return ans == "\x00", err
}

// Initialize ???
func (c Cutter) Initialize() {
	c.WriteString("\x1b\x04")
	c.Flush()
}

func (c Cutter) Ready() bool {
	c.WriteString("\x1b\x05")
	c.Flush()
	ans, _ := c.readResponse()
	return ans == "0"
}

func (c Cutter) Wait() {
	for !c.Ready() {
		time.Sleep(100 * time.Millisecond)
	}
}

func (c Cutter) Bezier(a int, p0, p1, p2, p3 Point) {
	defer c.Emit()
	fmt.Fprintf(c, "BZ%v,%v,%v,%v,%v", a, p0, p1, p2, p3)
}

type Orientation int

const (
	Portrait Orientation = iota
	Landscape
)

func (c Cutter) Orientation(l Orientation) {
	defer c.Emit()
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
	c.Emit()
	fmt.Fprint(c, "FY", state)
}

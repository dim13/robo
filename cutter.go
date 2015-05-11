package main

import (
	"bufio"
	"fmt"
	"time"
)

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
	A4     = Point{272 * MM, 200 * MM} // Portrait
	Origin = Point{0, 0}
)

type Cutter struct {
	*bufio.ReadWriter
}

func NewCutter(io *bufio.ReadWriter, o Orientation, rmlen Unit) Cutter {
	c := Cutter{io}
	c.Initialize()
	if !c.Ready() {
		fmt.Println("not ready")
		time.Sleep(time.Second)
	}

	c.GoHome() // Home
	fmt.Println(craftRobo, "Ver.", c.Version())

	pen := MediaID[113]
	c.MediaType(pen.ID)
	c.Speed(pen.Speed)
	c.Force(pen.Thickness)
	c.Overcut(pen.Overcut)

	c.TrackEnhancing(On)
	c.UnknownFE(0)

	fmt.Println("Calibration", c.GetCalibration())
	fmt.Println("FA", c.UnknownFA())
	if rmlen > 0 {
		c.RegMarkLen(rmlen)
	}
	c.Orientation(o)

	return Cutter{io}
}

const (
	NUL = 0x00
	ETX = 0x03 // End of Text
	ESC = 0x1B // Escape
	FS  = 0x1C // File Separator
)

func (c Cutter) Add(a ...interface{}) {
	fmt.Fprint(c, a...)
	c.WriteByte(ETX)
}

func (c Cutter) Send(a ...interface{}) {
	c.Add(a...)
	c.Flush()
}

type StepDirection byte

const (
	StepStop StepDirection = 1 << iota >> 1
	StepDown
	StepUp
	StepRight
	StepLeft
)

func (c Cutter) Esc(bytes ...byte) {
	c.WriteByte(ESC)
	for _, b := range bytes {
		c.WriteByte(b)
	}
	c.Flush()
}

func (c Cutter) Step(dir StepDirection) {
	c.Esc(NUL, byte(dir))
}

func (c Cutter) TestCut() {
	c.Send("FH")
}

func (c Cutter) TestCut2() {
	c.Send("FI")
}

// GoHome returns carret to home on same line
func (c Cutter) GoHome() {
	c.Send("TT")
}

// Home retuns carret to home position
func (c Cutter) Home() {
	c.Send("H")
}

func (c Cutter) SetCurrentOrigin() {
	c.Send("FJ")
}

func (c Cutter) SetOrigin(n int) {
	c.Send("SO", n)
}

func (c Cutter) Draw(p Point) {
	c.Send("D", p)
}

func (c Cutter) Move(p Point) {
	c.Send("M", p)
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
	Custom1 // 2 args ? --a-- b --a--
	Custom2 // 3 args ? --a-- b -c- b -c- b --a--
	Custom3 // 3 args ? --a-- b -c- b --a--
)

func (c Cutter) LineType(n LineStyle) {
	c.Send("L", n)
}

func (c Cutter) LineScale(n int) {
	c.Send("B", n)
}

func (c Cutter) Factor(t Triple) {
	c.Send("&", t)
}

func (c Cutter) Offset(p Point) {
	c.Send("^", p)
}

func (c Cutter) WriteLowerLeft(p Point) {
	c.Send("\\", p)
}

func (c Cutter) WriteUpperRight(p Point) {
	c.Send("Z", p)
}

// CuttingArea ???
func (c Cutter) CuttingArea(p Point) {
	c.Send("FU", p)
}

func (c Cutter) Response() string {
	ans, err := c.ReadString(ETX)
	if err != nil {
		panic(err)
	}
	return ans[:len(ans)-1]
}

// Version requests hardware version
func (c Cutter) Version() string {
	c.Send("FG")
	return c.returnString()
}

// MediaType (Meida ID)
func (c Cutter) MediaType(n int) {
	c.Send("FW", n)
}

// Speed 10..100 mm/s
func (c Cutter) Speed(n int) {
	if n >= 1 && n <= 10 {
		c.Send("!", n)
	}
}

func (c Cutter) Thickness(n int) {
	if n >= 1 && n <= 30 {
		c.Send("FX", n, ",0")
	}
}

func (c Cutter) Force(n int) {
	c.Thickness(n)
}

func (c Cutter) Overcut(n int) {
	c.Send("FC", n)
}

func (c Cutter) UnknownFE(n int) {
	c.Send("FE", n)
}

func (c Cutter) returnString() string {
	return c.Response()
}

func (c Cutter) returnUnit() Unit {
	return NewUnit(c.Response())
}

func (c Cutter) returnPoint() Point {
	return NewPoint(c.Response())
}

func (c Cutter) returnTriple() Triple {
	return NewTriple(c.Response())
}

func (c Cutter) RegMarkLen(n Unit) {
	c.Send("TB51,", n)
}

func (c Cutter) Calibrate() {
	c.Send("TB70")
}

// Sensor position
func (c Cutter) GetCalibration() Point {
	c.Send("TB71")
	return c.returnPoint()
}

// Emited after auto calibration
func (c Cutter) UnknownFQ5() Unit {
	c.Send("FQ5")
	return c.returnUnit()
}

func (c Cutter) SetCalibration(p Point) {
	if p.X > 40 || p.Y > 40 || p.X < -40 || p.Y < -40 {
		return
	}
	c.Send("TB72,", p)
}

// Arg: percent +/- 2.00% -> +/- 200
func (c Cutter) DistanseCorrection(n int) {
	if n > 200 || n < -200 {
		return
	}
	c.Send(c, "FB", n, ",0")
}

func (c Cutter) UnknownFA() Point {
	c.Send("FA")
	return c.returnPoint()
}

// VersionUpgrade
func (c Cutter) BootUpgrade() (string, error) {
	c.Esc(1)
	return c.ReadString(' ')
}

// Upgrade starts update sequence
// Send raw S-Record data after
func (c Cutter) Upgrade() bool {
	c.Send("CC1VERUP")
	return c.returnString() == string(NUL)
}

// Educated Guss, not tested
func (c Cutter) EnableDebug() {
	c.Send("FPGRFCC1")
}

// Initialize ???
func (c Cutter) Initialize() {
	c.Esc(4)
}

func (c Cutter) Ready() bool {
	c.Esc(5)
	return c.returnUnit() == 0
}

func (c Cutter) Wait() {
	for !c.Ready() {
		time.Sleep(100 * time.Millisecond)
	}
}

func (c Cutter) Bezier(a int, p0, p1, p2, p3 Point) {
	c.Send("BZ", a, ",", p0, ",", p1, ",", p2, ",", p3)
}

type Orientation int

const (
	Portrait Orientation = iota
	Landscape
)

func (c Cutter) Orientation(l Orientation) {
	c.Send("FN", l)
}

type OnOff int

const (
	On OnOff = iota
	Off
)

// TrackEnhancing moves paper back and forth for better traction
// Not for thickness less then 19
func (c Cutter) TrackEnhancing(state OnOff) {
	c.Send("FY", state)
}

func (c Cutter) SearchMarks(p Point) bool {
	c.Send("TB99")
	c.Send("TB55,1")
	c.Send("TB123,", p)
	return c.returnUnit() == 0
}

func (c Cutter) ManualSearchMarks(p Point) bool {
	c.Send("TB99")
	c.Send("TB55,1")
	c.Send("TB23,", p)
	return c.returnUnit() == 0
}

func (c Cutter) Circle(p Point, start, end Polar) {
	c.Send("W", p, ",",
		start.R, ",", end.R, ",",
		start.Theta, ",", end.Theta)
}

// Not supported?
func (c Cutter) Curve(a int, ph Path) {
	c.Add("Y", a)
	for _, p := range ph {
		c.Add(",", p)
	}
	c.Flush()
}

func (c Cutter) Ellipse(a int, p Point, start, end Polar, theta Unit) {
	c.Send(")", a, ",", p, ",", start.R, ",", end.R, ",",
		start.Theta, ",", end.Theta, ",", theta)
}

func (c Cutter) Gin() Triple {
	c.Send("G")
	return c.returnTriple()
}

func (c Cutter) CallGin() Triple {
	c.Send("C")
	return c.returnTriple()
}

func (c Cutter) ReadOffset() Point {
	c.Send("?")
	return c.returnPoint()
}

func (c Cutter) ReadLowerLeft() Point {
	c.Send("[")
	return c.returnPoint()
}

func (c Cutter) ReadUpperRight() Point {
	c.Send("U")
	return c.returnPoint()
}

func (c Cutter) StatusWord() string {
	c.Send("@")
	return c.returnString()
}

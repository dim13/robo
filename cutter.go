package main

import (
	"bufio"
	"fmt"
	"time"
)

type Point struct {
	X, Y float64
}

func (p Point) Add(u Point) Point {
	return Point{p.X + u.X, p.Y + u.Y}
}

func (p Point) AddX(u float64) Point {
	return Point{p.X + u, p.Y}
}

func (p Point) AddY(u float64) Point {
	return Point{p.X, p.Y + u}
}

type Polar struct {
	R, Theta float64
}

type Path []Point

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

type Triple struct {
	U, V, W float64
}

var (
	A4     = Point{5440, 4000} // Portrait
	Origin = Point{0, 0}
)

func (p Point) String() string {
	return fmt.Sprintf("%v,%v", p.X, p.Y)
}

func (t Triple) String() string {
	return fmt.Sprintf("%v,%v,%v", t.U, t.V, t.W)
}

type Cutter struct {
	*bufio.ReadWriter
}

func NewCutter(io *bufio.ReadWriter, o Orientation) Cutter {
	c := Cutter{io}
	c.Initialize()
	if !c.Ready() {
		fmt.Println("not ready")
		time.Sleep(time.Second)
	}

	c.GoHome() // Home
	v, _ := c.Version()
	fmt.Println(craftRobo, "Ver.", v)

	pen := MediaID[113]
	c.MediaType(pen.ID)
	c.Speed(pen.Speed)
	c.Force(pen.Thickness)
	c.CutterOffset(pen.Offset)

	c.TrackEnhancing(On)
	c.UnknownFE(0)

	fmt.Println("Calibration", c.GetCalibration())
	fmt.Println("FA", c.UnknownFA())

	c.Orientation(o)

	return Cutter{io}
}

const (
	NUL = 0x00
	ETX = 0x03 // End of Text
	ESC = 0x1b
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

func (c Cutter) SetOrigin(p Point) {
	c.Send("SO", p)
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

func (c Cutter) readResponse() (string, error) {
	ans, err := c.ReadString(ETX)
	if err != nil {
		return "", err
	}
	return ans[:len(ans)-1], nil
}

// Version requests hardware version
func (c Cutter) Version() (string, error) {
	c.Send("FG")
	return c.readResponse()
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

func (c Cutter) CutterOffset(n int) {
	c.Send("FC", n)
}

func (c Cutter) UnknownFE(n int) {
	c.Send("FE", n)
}

func (c Cutter) parseDigit() (n int) {
	s, _ := c.readResponse()
	fmt.Sscanf(s, "%v", &n)
	return
}

func (c Cutter) parsePoint() (p Point) {
	s, _ := c.readResponse()
	fmt.Sscanf(s, "%v,%v", &p.X, &p.Y)
	return
}

func (c Cutter) parseTriple() (t Triple) {
	s, _ := c.readResponse()
	fmt.Sscanf(s, "%v,%v,%v", &t.U, &t.V, &t.W)
	return
}

func (c Cutter) RegMarkLen(n int) {
	c.Send("TB51,", n)
}

func (c Cutter) Calibrate() {
	c.Send("TB70")
}

// Sensor position
func (c Cutter) GetCalibration() Point {
	c.Send("TB71")
	return c.parsePoint()
}

// Emited after auto calibration
func (c Cutter) UnknownFQ5() int {
	c.Send("FQ5")
	return c.parseDigit()
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
	return c.parsePoint()
}

// VersionUpgrade
func (c Cutter) BootUpgrade() (string, error) {
	c.Esc(1)
	return c.ReadString(' ')
}

// Upgrade starts update sequence
// Send raw S-Record data after
func (c Cutter) Upgrade() (bool, error) {
	c.Send("CC1VERUP")
	ans, err := c.readResponse()
	return ans == string(NUL), err
}

// Initialize ???
func (c Cutter) Initialize() {
	c.Esc(4)
}

func (c Cutter) Ready() bool {
	c.Esc(5)
	return c.parseDigit() == 0
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

func (c Cutter) SearchMarks(p Point, l int) bool {
	c.RegMarkLen(l)
	c.Send("TB99")
	c.Send("TB55,1")
	c.Send("TB123,", p)
	return c.parseDigit() == 0
}

func (c Cutter) ManualSearchMarks(p Point, l int) bool {
	c.RegMarkLen(l)
	c.Send("TB99")
	c.Send("TB55,1")
	c.Send("TB23,", p)
	return c.parseDigit() == 0
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

func (c Cutter) Ellipse(a int, p Point, start, end Polar, theta float64) {
	c.Send(")", a, ",", p, ",", start.R, ",", end.R, ",",
		start.Theta, ",", end.Theta, ",", theta)
}

func (c Cutter) Gin() Triple {
	c.Send("G")
	return c.parseTriple()
}

func (c Cutter) CallGin() Triple {
	c.Send("C")
	return c.parseTriple()
}

func (c Cutter) ReadOffset() Point {
	c.Send("?")
	return c.parsePoint()
}

func (c Cutter) ReadLowerLeft() Point {
	c.Send("[")
	return c.parsePoint()
}

func (c Cutter) ReadUpperRight() Point {
	c.Send("U")
	return c.parsePoint()
}

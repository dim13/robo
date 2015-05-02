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

func NewCutter(io *bufio.ReadWriter, o Orientation) Cutter {
	c := Cutter{io}
	c.Initialize()
	if !c.Ready() {
		fmt.Println("not ready")
	}

	c.CR() // Home
	v, _ := c.Version()
	fmt.Println("Craft ROBO Ver.", v)

	pen := MediaID[113]
	c.MediaType(pen.ID)
	c.Speed(pen.Speed)
	c.Force(pen.Thickness)
	c.UnknownFC(pen.FC)

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

func (c Cutter) Send(a ...interface{}) {
	fmt.Fprint(c, a...)
	c.WriteByte(ETX)
	c.Flush()
}

func (c Cutter) TestCut() {
	c.Send("FH")
}

type StepDirection byte

const (
	StepStop StepDirection = 1 << iota >> 1
	StepDown
	StepUp
	StepRight
	StepLeft
)

func (c Cutter) Step(dir StepDirection) {
	c.WriteByte(ESC)
	c.WriteByte(NUL)
	c.WriteByte(byte(dir))
	c.Flush()
}

// CR returns carret to home on same line
func (c Cutter) CR() {
	c.Send("TT")
}

// Home retuns carret to home position
func (c Cutter) Home() {
	c.Send("H")
}

func (c Cutter) SetOrigin() {
	c.Send("FJ")
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

func (c Cutter) Factor(n int) {
	c.Send("&", n, ",", n, ",", n)
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

func (c Cutter) ReadUpperRight() (string, error) {
	c.Send("U")
	return c.readResponse()
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

func (c Cutter) UnknownFC(n int) {
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

func (c Cutter) RegistrationMarksLength(n int) {
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

// Emited after calibration
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

func (c Cutter) UnknownTB(n int) (string, error) {
	c.Send("TB", n)
	return c.readResponse()
}

func (c Cutter) UnknownFA() Point {
	c.Send("FA")
	return c.parsePoint()
}

// VersionUpgrade
func (c Cutter) BootVersion() (string, error) {
	c.WriteByte(ESC)
	c.WriteByte(1)
	c.Flush()
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
	c.WriteByte(ESC)
	c.WriteByte(4)
	c.Flush()
}

func (c Cutter) Ready() bool {
	c.WriteByte(ESC)
	c.WriteByte(5)
	c.Flush()
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

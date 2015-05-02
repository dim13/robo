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

	m := MediaID[113]
	c.MediaType(m.ID)
	c.Speed(m.Speed)
	c.Force(m.Thickness)
	c.UnknownFC(m.FC)
	c.TrackEnhancing(On)
	c.UnknownFE(0)

	p := c.GetCalibration()
	fmt.Println("Calibration", p)

	p = c.UnknownFA()
	fmt.Println("FA", p)
	c.Orientation(o)

	return Cutter{io}
}

const (
	NUL = 0x00
	ETX = 0x03 // End of Text
	ESC = 0x1b
)

func (c Cutter) Emit() {
	c.WriteByte(ETX)
	c.Flush()
}

func (c Cutter) TestCut() {
	c.WriteString("FH")
	c.Emit()
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
	c.WriteString("TT")
	c.Emit()
}

// Home retuns carret to home position
func (c Cutter) Home() {
	c.WriteString("H")
	c.Emit()
}

func (c Cutter) SetOrigin() {
	c.WriteString("FJ")
	c.Emit()
}

func (c Cutter) Draw(p Point) {
	fmt.Fprint(c, "D", p)
	c.Emit()
}

func (c Cutter) Move(p Point) {
	fmt.Fprint(c, "M", p)
	c.Emit()
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
	fmt.Fprint(c, "L", n)
	c.Emit()
}

func (c Cutter) LineScale(n int) {
	fmt.Fprint(c, "B", n)
	c.Emit()
}

func (c Cutter) Factor(n int) {
	fmt.Fprintf(c, "&%v,%v,%v", n, n, n)
	c.Emit()
}

func (c Cutter) Offset(p Point) {
	fmt.Fprint(c, "^", p)
	c.Emit()
}

func (c Cutter) WriteLowerLeft(p Point) {
	fmt.Fprint(c, "\\", p)
	c.Emit()
}

func (c Cutter) WriteUpperRight(p Point) {
	fmt.Fprint(c, "Z", p)
	c.Emit()
}

// CuttingArea ???
func (c Cutter) CuttingArea(p Point) {
	fmt.Fprint(c, "FU", p)
	c.Emit()
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
	c.WriteString("FG")
	c.Flush()
	return c.readResponse()
}

// MediaType (Meida ID)
func (c Cutter) MediaType(n int) {
	fmt.Fprint(c, "FW", n)
	c.Emit()
}

func (c Cutter) ReadUpperRight() (string, error) {
	c.WriteString("U")
	c.Flush()
	return c.readResponse()
}

// Speed 10..100 mm/s
func (c Cutter) Speed(n int) {
	if n >= 1 && n <= 10 {
		fmt.Fprint(c, "!", n)
		c.Emit()
	}
}

func (c Cutter) Thickness(n int) {
	if n >= 1 && n <= 30 {
		fmt.Fprint(c, "FX", n, ",0")
		c.Emit()
	}
}

func (c Cutter) Force(n int) {
	c.Thickness(n)
}

func (c Cutter) UnknownFC(n int) {
	fmt.Fprint(c, "FC", n)
	c.Emit()
}

func (c Cutter) UnknownFE(n int) {
	fmt.Fprint(c, "FE", n)
	c.Emit()
}

func parseDigit(s string) (n int) {
	fmt.Sscanf(s, "%v", &n)
	return
}

func parsePoint(s string) (p Point) {
	fmt.Sscanf(s, "%v,%v", &p.X, &p.Y)
	return
}

func (c Cutter) RegistrationMarksLength(n int) {
	fmt.Fprint(c, "TB51,", n)
	c.Emit()
}

func (c Cutter) Calibrate() {
	fmt.Fprint(c, "TB70")
	c.Emit()
}

// Sensor position
func (c Cutter) GetCalibration() Point {
	fmt.Fprint(c, "TB71")
	c.Emit()
	s, _ := c.readResponse()
	return parsePoint(s)
}

// Emited after calibration
func (c Cutter) UnknownFQ5() int {
	fmt.Fprint(c, "FQ5")
	c.Emit()
	s, _ := c.readResponse()
	return parseDigit(s)
}

func (c Cutter) SetCalibration(p Point) {
	if p.X > 40 || p.Y > 40 || p.X < -40 || p.Y < -40 {
		return
	}
	fmt.Fprint(c, "TB72,", p)
	c.Emit()
}

// Arg: percent +/- 2.00% -> +/- 200
func (c Cutter) DistanseCorrection(n int) {
	if n > 200 || n < -200 {
		return
	}
	fmt.Fprint(c, "FB", n, ",0")
	c.Emit()
}

func (c Cutter) UnknownTB(n int) (string, error) {
	fmt.Fprint(c, "TB", n)
	c.Emit()
	return c.readResponse()
}

func (c Cutter) UnknownFA() Point {
	fmt.Fprint(c, "FA")
	c.Emit()
	s, _ := c.readResponse()
	return parsePoint(s)
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
	c.WriteString("CC1VERUP")
	c.Flush()
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
	ans, _ := c.readResponse()
	return ans == "0"
}

func (c Cutter) Wait() {
	for !c.Ready() {
		time.Sleep(100 * time.Millisecond)
	}
}

func (c Cutter) Bezier(a int, p0, p1, p2, p3 Point) {
	fmt.Fprintf(c, "BZ%v,%v,%v,%v,%v", a, p0, p1, p2, p3)
	c.Emit()
}

type Orientation int

const (
	Portrait Orientation = iota
	Landscape
)

func (c Cutter) Orientation(l Orientation) {
	fmt.Fprint(c, "FN", l)
	c.Emit()
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

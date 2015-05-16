package robo

import (
	"bufio"
	"fmt"
)

const (
	NUL = 0x00
	ETX = 0x03 // End of Text
	ESC = 0x1B // Escape
	FS  = 0x1C // File Separator
)

func etx(c *bufio.Writer) {
	c.WriteByte(ETX)
	c.Flush()
}

func (p Point) send(c *bufio.Writer, a ...interface{}) {
	fmt.Fprint(c, a[0], p)
	for _, arg := range a[1:] {
		fmt.Fprint(c, arg)
	}
	etx(c)
}

func (p Point) Draw(c *bufio.Writer)         { p.send(c, "D") }
func (p Point) DrawRelative(c *bufio.Writer) { p.send(c, "E") }
func (p Point) Move(c *bufio.Writer)         { p.send(c, "M") }
func (p Point) MoveRelative(c *bufio.Writer) { p.send(c, "O") }
func (p Point) Offset(c *bufio.Writer)       { p.send(c, "^") }
func (p Point) LowerLeft(c *bufio.Writer)    { p.send(c, "\\") }
func (p Point) UpperRight(c *bufio.Writer)   { p.send(c, "Z") }
func (p Point) CuttingArea(c *bufio.Writer)  { p.send(c, "FU") }
func (p Point) Calibration(c *bufio.Writer)  { p.send(c, "TB72,") }

func (p Point) SearchMarks(c *bufio.ReadWriter) bool {
	send(c.Writer, "TB99")
	send(c.Writer, "TB55,1")
	send(c.Writer, "TB123,", p)
	return parseUnit(recv(c.Reader)) == 0
}

func (p Point) SearchMarksManual(c *bufio.ReadWriter) bool {
	send(c.Writer, "TB99")
	send(c.Writer, "TB55,1")
	send(c.Writer, "TB23,", p)
	return parseUnit(recv(c.Reader)) == 0
}

func (ph Path) send(c *bufio.Writer, a ...interface{}) {
	fmt.Fprint(c, a...)
	for _, p := range ph {
		fmt.Fprint(c, p, ",")
	}
	etx(c)
}

func (ph Path) Draw(c *bufio.Writer)                 { ph.send(c, "D") }
func (ph Path) DrawRelative(c *bufio.Writer)         { ph.send(c, "E") }
func (ph Path) Curve(c *bufio.Writer, a int)         { ph.send(c, "Y", a, ",") }
func (ph Path) CurveRelative(c *bufio.Writer, a int) { ph.send(c, "_", a, ",") }

func (ph Path) Bezier(c *bufio.Writer, a int)  { ph[0:4].send(c, "BZ", a, ",") }
func (ph Path) Circle(c *bufio.Writer)         { ph[0:3].send(c, "W") }
func (ph Path) Circle3P(c *bufio.Writer)       { ph[0:3].send(c, "WP") }
func (ph Path) CircleRelative(c *bufio.Writer) { ph[0:2].send(c, "]") }
func (ph Path) Ellipse(c *bufio.Writer)        { ph[0:4].send(c, ")") }

func (ph Path) Line(c *bufio.Writer) {
	ph[0].Move(c)
	ph[1:].Draw(c)
}

func (u Unit) send(c *bufio.Writer, a ...interface{}) {
	fmt.Fprint(c, a[0], u)
	for _, arg := range a[1:] {
		fmt.Fprint(c, arg)
	}
	etx(c)
}

func (u Unit) Origin(c *bufio.Writer)             { u.send(c, "SO") }
func (u Unit) LineScale(c *bufio.Writer)          { u.send(c, "B") }
func (u Unit) Media(c *bufio.Writer)              { u.send(c, "FW") }
func (u Unit) Speed(c *bufio.Writer)              { u.send(c, "!") }
func (u Unit) Force(c *bufio.Writer)              { u.send(c, "FX") }
func (u Unit) Overcut(c *bufio.Writer)            { u.send(c, "FC") }
func (u Unit) UnknownFE(c *bufio.Writer)          { u.send(c, "FE") }
func (u Unit) DistanceCorrection(c *bufio.Writer) { u.send(c, "FB", ",0") }
func (u Unit) TrackEnhancing(c *bufio.Writer)     { u.send(c, "FY") }
func (u Unit) RegMarkLen(c *bufio.Writer)         { u.send(c, "TB51,") }

func esc(c *bufio.Writer, bytes ...byte) {
	c.WriteByte(ESC)
	for _, b := range bytes {
		c.WriteByte(b)
	}
	c.Flush()
}

func Init(c *bufio.Writer) { esc(c, 4) }

func Ready(c *bufio.ReadWriter) bool {
	esc(c.Writer, 5)
	return parseUnit(recv(c.Reader)) == 0
}

func (u Unit) UnknownFQ(c *bufio.ReadWriter) Unit {
	u.send(c.Writer, "FQ", u)
	return parseUnit(recv(c.Reader))
}

func recv(c *bufio.Reader) string {
	ans, err := c.ReadString(ETX)
	if err != nil {
		panic(err)
	}
	return ans[:len(ans)-1]
}

func send(c *bufio.Writer, a ...interface{}) {
	fmt.Fprint(c, a...)
	etx(c)
}

func Raw(c *bufio.Writer, a ...interface{}) {
	for _, cmd := range a {
		send(c, cmd)
	}
}

func GoHome(c *bufio.Writer)    { send(c, "TT") }
func Home(c *bufio.Writer)      { send(c, "H") }
func Origin(c *bufio.Writer)    { send(c, "FJ") }
func Calibrate(c *bufio.Writer) { send(c, "TB70") }
func TestCut(c *bufio.Writer)   { send(c, "FH") }
func TestLoop(c *bufio.Writer)  { send(c, "FI") }

func point(c *bufio.ReadWriter, cmd string) Point {
	send(c.Writer, cmd)
	return parsePoint(recv(c.Reader))
}

func Calibration(c *bufio.ReadWriter) Point        { return point(c, "TB71") }
func Offset(c *bufio.ReadWriter) Point             { return point(c, "?") }
func LowerLeft(c *bufio.ReadWriter) Point          { return point(c, "[") }
func UpperRight(c *bufio.ReadWriter) Point         { return point(c, "U") }
func DistanceCorrection(c *bufio.ReadWriter) Point { return point(c, "FA") }

func str(c *bufio.ReadWriter, cmd string) string {
	send(c.Writer, cmd)
	return recv(c.Reader)
}

func Version(c *bufio.ReadWriter) string    { return str(c, "FG") }
func StatusWord(c *bufio.ReadWriter) string { return str(c, "@") }

type Orientation int

const (
	Portrait Orientation = iota
	Landscape
)

func (o Orientation) Orientation(c *bufio.Writer) { send(c, "FN", o) }

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

func (l LineStyle) LineStyle(c *bufio.Writer) { send(c, "L", l) }
func (p Point) LineStyle(c *bufio.Writer)     { p.send(c, "L100,1,") }

func triple(c *bufio.ReadWriter, cmd string) Triple {
	send(c.Writer, cmd)
	return parseTriple(recv(c.Reader))
}

func Gin(c *bufio.ReadWriter) Triple     { return triple(c, "G") }
func CallGin(c *bufio.ReadWriter) Triple { return triple(c, "C") }

func (t Triple) send(c *bufio.Writer, cmd string) {
	fmt.Fprint(c, cmd, t)
	etx(c)
}

func (t Triple) Factor(c *bufio.Writer) { t.send(c, "&") }

func Initialize(c *bufio.ReadWriter, mid int, o Orientation) {
	Init(c.Writer)
	if !Ready(c) {
		fmt.Println("not ready")
	}

	GoHome(c.Writer)
	fmt.Println(craftRobo, "Ver.", Version(c))

	if pen, ok := MediaID[mid]; ok {
		pen.Apply(c.Writer)
	}

	Unit(0).TrackEnhancing(c.Writer)
	Unit(0).UnknownFE(c.Writer)

	fmt.Println("Calibration", Calibration(c))
	fmt.Println("Correction ", DistanceCorrection(c))

	Unit(400).RegMarkLen(c.Writer)
	o.Orientation(c.Writer)
}

type Direction byte

const (
	Stop Direction = 1 << iota >> 1
	Down
	Up
	Right
	Left
)

func (d Direction) Step(c *bufio.Writer) { esc(c, NUL, byte(d)) }

// Untested
func BootUpgrade(c *bufio.ReadWriter) string {
	esc(c.Writer, 1)
	s, _ := c.ReadString(' ')
	return s
}
func UpdateFirmware(c *bufio.ReadWriter) bool {
	return str(c, "CC1VERUP") == string(NUL)
}
func EnableDebug(c *bufio.Writer) { send(c, "FP,GRFCC1") }

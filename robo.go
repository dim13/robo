package robo

import (
	"bufio"
	"fmt"
	"io"
)

const (
	NUL = 0x00
	ETX = 0x03 // End of Text
	ESC = 0x1B // Escape
	FS  = 0x1C // File Separator
)

func etx(c io.Writer) {
	c.Write([]byte{ETX})
}

func (p Point) send(c io.Writer, a ...any) {
	fmt.Fprint(c, a[0], p)
	for _, arg := range a[1:] {
		fmt.Fprint(c, arg)
	}
	etx(c)
}

func (p Point) Draw(c io.Writer)         { p.send(c, "D") }
func (p Point) DrawRelative(c io.Writer) { p.send(c, "E") }
func (p Point) Move(c io.Writer)         { p.send(c, "M") }
func (p Point) MoveRelative(c io.Writer) { p.send(c, "O") }
func (p Point) Offset(c io.Writer)       { p.send(c, "^") }
func (p Point) LowerLeft(c io.Writer)    { p.send(c, "\\") }
func (p Point) UpperRight(c io.Writer)   { p.send(c, "Z") }
func (p Point) CuttingArea(c io.Writer)  { p.send(c, "FU") }
func (p Point) Calibration(c io.Writer)  { p.send(c, "TB72,") }

func (p Point) SearchMarks(c io.ReadWriter, auto bool) bool {
	send(c, "TB99")
	send(c, "TB55,1")
	if auto {
		send(c, "TB123,", p)
	} else {
		send(c, "TB23,", p)
	}
	return parseUnit(recv(c)) == 0
}

func (p Point) Scale(f Unit) Point {
	return Point{p.X * f, p.Y * f}
}

func (p Point) Add(f Point) Point {
	return Point{p.X + f.X, p.Y + f.Y}
}

func (p Path) Scale(f Unit) (ret Path) {
	for _, pt := range p {
		ret = append(ret, pt.Scale(f))
	}
	return
}

func (p Path) Add(f Point) (ret Path) {
	for _, pt := range p {
		ret = append(ret, pt.Add(f))
	}
	return
}

func (p Path) send(c io.Writer, a ...any) {
	fmt.Fprint(c, a...)
	for _, pt := range p {
		fmt.Fprint(c, pt, ",")
	}
	etx(c)
}

func (p Path) Draw(c io.Writer)                 { p.send(c, "D") }
func (p Path) DrawRelative(c io.Writer)         { p.send(c, "E") }
func (p Path) Curve(c io.Writer, a int)         { p.send(c, "Y", a, ",") }
func (p Path) CurveRelative(c io.Writer, a int) { p.send(c, "_", a, ",") }

func (p Path) Bezier(c io.Writer, a int)  { p[0:4].send(c, "BZ", a, ",") }
func (p Path) Circle(c io.Writer)         { p[0:3].send(c, "W") }
func (p Path) Circle3P(c io.Writer)       { p[0:3].send(c, "WP") }
func (p Path) CircleRelative(c io.Writer) { p[0:2].send(c, "]") }
func (p Path) Ellipse(c io.Writer)        { p[0:4].send(c, ")") }

func (p Path) Line(c io.Writer) {
	p[0].Move(c)
	p[1:].Draw(c)
}

func (u Unit) send(c io.Writer, a ...any) {
	fmt.Fprint(c, a[0], u)
	for _, arg := range a[1:] {
		fmt.Fprint(c, arg)
	}
	etx(c)
}

func (u Unit) Origin(c io.Writer)             { u.send(c, "SO") }
func (u Unit) LineScale(c io.Writer)          { u.send(c, "B") }
func (u Unit) Media(c io.Writer)              { u.send(c, "FW") }
func (u Unit) Speed(c io.Writer)              { u.send(c, "!") }
func (u Unit) Force(c io.Writer)              { u.send(c, "FX") }
func (u Unit) Overcut(c io.Writer)            { u.send(c, "FC") }
func (u Unit) UnknownFE(c io.Writer)          { u.send(c, "FE") }
func (u Unit) DistanceCorrection(c io.Writer) { u.send(c, "FB", ",0") }
func (u Unit) TrackEnhancing(c io.Writer)     { u.send(c, "FY") }
func (u Unit) RegMarkLen(c io.Writer)         { u.send(c, "TB51,") }

func esc(c io.Writer, bytes ...byte) {
	c.Write([]byte{ESC})
	c.Write(bytes)
}

func Init(c io.Writer) { esc(c, 4) }

func Ready(c io.ReadWriter) bool {
	esc(c, 5)
	return parseUnit(recv(c)) == 0
}

func (u Unit) UnknownFQ(c io.ReadWriter) Unit {
	u.send(c, "FQ", u)
	return parseUnit(recv(c))
}

func readString(r io.Reader, delim byte) (string, error) {
	return bufio.NewReader(r).ReadString(delim)
}

func recv(c io.Reader) string {
	ans, err := readString(c, ETX)
	if err != nil {
		panic(err)
	}
	return ans[:len(ans)-1]
}

func Recv(c io.Reader) string {
	return recv(c)
}

func send(c io.Writer, a ...any) {
	fmt.Fprint(c, a...)
	etx(c)
}

func Send(c io.Writer, a any) {
	send(c, a)
}

func GoHome(c io.Writer)    { send(c, "TT") }
func Home(c io.Writer)      { send(c, "H") }
func Origin(c io.Writer)    { send(c, "FJ") }
func Calibrate(c io.Writer) { send(c, "TB70") }
func TestCut(c io.Writer)   { send(c, "FH") }
func TestLoop(c io.Writer)  { send(c, "FI") }

func point(c io.ReadWriter, cmd string) Point {
	send(c, cmd)
	return parsePoint(recv(c))
}

func Calibration(c io.ReadWriter) Point        { return point(c, "TB71") }
func Offset(c io.ReadWriter) Point             { return point(c, "?") }
func LowerLeft(c io.ReadWriter) Point          { return point(c, "[") }
func UpperRight(c io.ReadWriter) Point         { return point(c, "U") }
func DistanceCorrection(c io.ReadWriter) Point { return point(c, "FA") }

func str(c io.ReadWriter, cmd string) string {
	send(c, cmd)
	return recv(c)
}

func Version(c io.ReadWriter) string    { return str(c, "FG") }
func StatusWord(c io.ReadWriter) string { return str(c, "@") }

type Orientation int

const (
	Portrait Orientation = iota
	Landscape
)

func (o Orientation) Orientation(c io.Writer) {
	orientation = o
	send(c, "FN", o)
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

func (l LineStyle) LineStyle(c io.Writer) { send(c, "L", l) }
func (p Point) LineStyle(c io.Writer)     { p.send(c, "L100,1,") }

func triple(c io.ReadWriter, cmd string) Triple {
	send(c, cmd)
	return parseTriple(recv(c))
}

func Gin(c io.ReadWriter) Triple     { return triple(c, "G") }
func CallGin(c io.ReadWriter) Triple { return triple(c, "C") }

func (t Triple) send(c io.Writer, cmd string) {
	fmt.Fprint(c, cmd, t)
	etx(c)
}

func (t Triple) Factor(c io.Writer) { t.send(c, "&") }

func Initialize(c io.ReadWriter, mid int, o Orientation) {
	Init(c)
	if !Ready(c) {
		fmt.Println("not ready")
	}

	GoHome(c)
	fmt.Println("Craft ROBO Ver.", Version(c))

	if pen, ok := MediaID[mid]; ok {
		pen.Apply(c)
	}

	Unit(0).TrackEnhancing(c)
	Unit(0).UnknownFE(c)

	fmt.Println("Calibration", Calibration(c))
	fmt.Println("Correction ", DistanceCorrection(c))

	Unit(400).RegMarkLen(c)
	o.Orientation(c)
}

type Direction byte

const (
	Stop Direction = 1 << iota >> 1
	Down
	Up
	Right
	Left
)

func (d Direction) Step(c io.Writer) { esc(c, NUL, byte(d)) }

// Untested
func BootUpgrade(c io.ReadWriter) string {
	esc(c, 1)
	s, _ := readString(c, ' ')
	return s
}
func UpdateFirmware(c io.ReadWriter) bool {
	return str(c, "CC1VERUP") == string(rune(NUL))
}
func EnableDebug(c io.Writer) { send(c, "FP,GRFCC1") }

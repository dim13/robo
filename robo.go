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

// send writes one command with its terminator in a single write, so it
// travels as one USB transfer.
func send(c io.Writer, s string) {
	c.Write(append([]byte(s), ETX))
}

func (p Point) send(c io.Writer, a ...any) {
	s := fmt.Sprint(a[0], p)
	for _, arg := range a[1:] {
		s += fmt.Sprint(arg)
	}
	send(c, s)
}

func (p Point) Draw(c io.Writer)         { p.send(c, "D") }
func (p Point) DrawRelative(c io.Writer) { p.send(c, "E") }
func (p Point) Move(c io.Writer)         { p.send(c, "M") }
func (p Point) MoveRelative(c io.Writer) { p.send(c, "O") }
func (p Point) Origin(c io.Writer)       { p.send(c, "^") }
func (p Point) LowerLeft(c io.Writer)    { p.send(c, "\\") }
func (p Point) UpperRight(c io.Writer)   { p.send(c, "Z") }
func (p Point) CuttingArea(c io.Writer)  { p.send(c, "FU") }
func (p Point) Calibration(c io.Writer)  { p.send(c, "TB72,") }

func (p Point) SearchMarks(c io.ReadWriter, auto bool) (bool, error) {
	Send(c, "TB99")
	Send(c, "TB55,1")
	if auto {
		Send(c, "TB123,", p)
	} else {
		Send(c, "TB23,", p)
	}
	ans, err := Recv(c)
	if err != nil {
		return false, err
	}
	return parseUnit(ans) == 0, nil
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
	s := fmt.Sprint(a...)
	for _, pt := range p {
		s += fmt.Sprint(pt, ",")
	}
	send(c, s)
}

func (p Path) Draw(c io.Writer)                 { p.send(c, "D") }
func (p Path) DrawRelative(c io.Writer)         { p.send(c, "E") }
func (p Path) Curve(c io.Writer, a int)         { p.send(c, "Y", a, ",") }
func (p Path) CurveRelative(c io.Writer, a int) { p.send(c, "_", a, ",") }

func (p Path) Bezier(c io.Writer, a int)   { p[0:4].send(c, "BZ", a, ",") }
func (p Path) Circle(c io.Writer)          { p[0:3].send(c, "W") }
func (p Path) Circle3P(c io.Writer, a int) { p[0:3].send(c, "WP", a, ",") }
func (p Path) CircleRelative(c io.Writer)  { p[0:2].send(c, "]") }
func (p Path) Ellipse(c io.Writer)         { p[0:4].send(c, ")") }

func (p Path) Line(c io.Writer) {
	p[0].Move(c)
	p[1:].Draw(c)
}

func (u Unit) send(c io.Writer, a ...any) {
	s := fmt.Sprint(a[0], u)
	for _, arg := range a[1:] {
		s += fmt.Sprint(arg)
	}
	send(c, s)
}

func (u Unit) Origin(c io.Writer)         { u.send(c, "SO") }
func (u Unit) LineScale(c io.Writer)      { u.send(c, "B") }
func (u Unit) Media(c io.Writer)          { u.send(c, "FW") }
func (u Unit) Speed(c io.Writer)          { u.send(c, "!") }
func (u Unit) Force(c io.Writer)          { u.send(c, "FX") }
func (u Unit) CutterOffset(c io.Writer)   { u.send(c, "FC") } // 0.025 mm steps, 0..45, 0 for pens
func (u Unit) Overcut(c io.Writer)        { u.send(c, "FE") } // 1 on, 0 off
func (u Unit) TrackEnhancing(c io.Writer) { u.send(c, "FY") } // 0 on, 1 off
func (u Unit) RegMarkLen(c io.Writer)     { u.send(c, "TB51,") }

// DistanceCorrection sets both corrections, -200..200 each (±2 %).
func (p Point) DistanceCorrection(c io.Writer) { Send(c, "FB", p.X, ",", p.Y) }

// esc sends an escape sequence in a single write: the firmware only
// recognises it when ESC and its argument arrive in the same USB packet.
func esc(c io.Writer, bytes ...byte) {
	c.Write(append([]byte{ESC}, bytes...))
}

func Init(c io.Writer) { esc(c, 4) }

func Ready(c io.ReadWriter) (bool, error) {
	esc(c, 5)
	ans, err := Recv(c)
	if err != nil {
		return false, err
	}
	return parseUnit(ans) == 0, nil
}

func (u Unit) Query(c io.ReadWriter) (Unit, error) {
	u.send(c, "FQ")
	ans, err := Recv(c)
	if err != nil {
		return 0, err
	}
	return parseUnit(ans), nil
}

type stringReader interface {
	ReadString(delim byte) (string, error)
}

// Recv reads one reply up to ETX. Pass the same buffered reader for the
// whole session (USB has one); a fresh buffer would drop read-ahead bytes.
func Recv(c io.Reader) (string, error) {
	r, ok := c.(stringReader)
	if !ok {
		r = bufio.NewReader(c)
	}
	ans, err := r.ReadString(ETX)
	if err != nil {
		return "", err
	}
	return ans[:len(ans)-1], nil
}

func Send(c io.Writer, a ...any) {
	send(c, fmt.Sprint(a...))
}

func GoHome(c io.Writer)    { Send(c, "TT") }
func Home(c io.Writer)      { Send(c, "H") }
func Origin(c io.Writer)    { Send(c, "FJ") }
func Calibrate(c io.Writer) { Send(c, "TB70") }
func TestCut(c io.Writer)   { Send(c, "FH") }
func TestLoop(c io.Writer)  { Send(c, "FI") }

func point(c io.ReadWriter, cmd string) (Point, error) {
	Send(c, cmd)
	ans, err := Recv(c)
	if err != nil {
		return Point{}, err
	}
	return parsePoint(ans), nil
}

func Calibration(c io.ReadWriter) (Point, error)        { return point(c, "TB71") }
func Offset(c io.ReadWriter) (Point, error)             { return point(c, "?") }
func LowerLeft(c io.ReadWriter) (Point, error)          { return point(c, "[") }
func UpperRight(c io.ReadWriter) (Point, error)         { return point(c, "U") }
func DistanceCorrection(c io.ReadWriter) (Point, error) { return point(c, "FA") }

func str(c io.ReadWriter, cmd string) (string, error) {
	Send(c, cmd)
	return Recv(c)
}

func Version(c io.ReadWriter) (string, error)    { return str(c, "FG") }
func StatusWord(c io.ReadWriter) (string, error) { return str(c, "@") }

type Orientation int

const (
	Portrait Orientation = iota
	Landscape
)

func (o Orientation) Orientation(c io.Writer) {
	orientation = o
	Send(c, "FN", o)
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

func (l LineStyle) LineStyle(c io.Writer) { Send(c, "L", l) }
func (p Point) LineStyle(c io.Writer)     { p.send(c, "L100,1,") }

func triple(c io.ReadWriter, cmd string) (Triple, error) {
	Send(c, cmd)
	ans, err := Recv(c)
	if err != nil {
		return Triple{}, err
	}
	return parseTriple(ans), nil
}

func Gin(c io.ReadWriter) (Triple, error)     { return triple(c, "G") }
func CallGin(c io.ReadWriter) (Triple, error) { return triple(c, "C") }

func (t Triple) send(c io.Writer, cmd string) {
	Send(c, cmd, t)
}

func (t Triple) Factor(c io.Writer) { t.send(c, "&") }

func Initialize(c io.ReadWriter, mid int, o Orientation) error {
	Init(c)
	ready, err := Ready(c)
	if err != nil {
		return fmt.Errorf("ready: %w", err)
	}
	if !ready {
		fmt.Println("not ready")
	}

	GoHome(c)
	ver, err := Version(c)
	if err != nil {
		return fmt.Errorf("version: %w", err)
	}
	fmt.Println("Craft ROBO Ver.", ver)

	if pen, ok := MediaID[mid]; ok {
		pen.Apply(c)
	}

	Unit(0).TrackEnhancing(c)
	Unit(0).Overcut(c)

	cal, err := Calibration(c)
	if err != nil {
		return fmt.Errorf("calibration: %w", err)
	}
	fmt.Println("Calibration", cal)
	cor, err := DistanceCorrection(c)
	if err != nil {
		return fmt.Errorf("distance correction: %w", err)
	}
	fmt.Println("Correction ", cor)

	Unit(400).RegMarkLen(c)
	o.Orientation(c)
	return nil
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

func EnableDebug(c io.Writer) { Send(c, "FPGRFCC1") } // service mode until reset

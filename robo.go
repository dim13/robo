package main

import (
	"bufio"
	"fmt"
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

/*
func (p Point) SearchMarks(c *bufio.ReadWriter) bool {
	send(c, "TB99")
	send(c, "TB55,1")
	send(c, "TB123,", p)
	return true // == 0
}

func (ph Path) send(c *bufio.Writer, cmd string) {
	defer etx(c)
	fmt.Fprint(c, "D")
	for _, p := range ph {
		fmt.Fprint(c, p, ",")
	}
}
*/

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

func Version(c *bufio.ReadWriter) string {
	send(c.Writer, "FG")
	return recv(c.Reader)
}

func GoHome(c *bufio.Writer)    { send(c, "TT") }
func Home(c *bufio.Writer)      { send(c, "H") }
func Origin(c *bufio.Writer)    { send(c, "FJ") }
func Calibrate(c *bufio.Writer) { send(c, "TB70") }

func point(c *bufio.ReadWriter, cmd string) Point {
	send(c.Writer, cmd)
	return parsePoint(recv(c.Reader))
}

func Calibration(c *bufio.ReadWriter) Point        { return point(c, "TB71") }
func Offset(c *bufio.ReadWriter) Point             { return point(c, "?") }
func LowerLeft(c *bufio.ReadWriter) Point          { return point(c, "[") }
func UpperRight(c *bufio.ReadWriter) Point         { return point(c, "U") }
func StatusWord(c *bufio.ReadWriter) Point         { return point(c, "@") }
func DistanceCorrection(c *bufio.ReadWriter) Point { return point(c, "FA") }

func (o Orientation) Orientation(c *bufio.Writer) { send(c, "FN", o) }
func (l LineStyle) LineStyle(c *bufio.Writer)     { send(c, "L", l) }
func (p Point) LineStyle(c *bufio.Writer)         { p.send(c, "L100,1,") }

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

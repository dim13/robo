package robo

import (
	"fmt"
	"math"
	"strconv"
)

const (
	MM = Unit(20.0)
	CM = 10 * MM
	M  = 100 * CM
	IN = 25.4 * MM
	FT = 12 * IN
	PT = IN / 72
)

var (
	Zero   = Point{0, 0}
	Margin = Point{25 * MM, 10 * MM}
	A4     = Point{297 * MM, 210 * MM}
	Letter = Point{11 * IN, 8.5 * IN}
)

type Unit float64

// String formats u for the plotter: it keeps two decimals and drops the
// sign of -0.x, so values in (-1, 0) are rounded to an integer.
func (u Unit) String() string {
	f := math.Round(float64(u)*100) / 100
	if f > -1 && f < 0 {
		f = math.Round(f)
	}
	if f == 0 {
		f = 0 // no "-0"
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func parseUnit(s string) (u Unit) {
	fmt.Sscanf(s, "%v", &u)
	return
}

type Point struct {
	X, Y Unit
}

func (p Point) Sub(o Point) Point {
	return Point{
		X: p.X - o.X,
		Y: p.Y - o.Y,
	}
}

var orientation = Portrait

func (p Point) String() (s string) {
	switch orientation {
	case Portrait:
		s = fmt.Sprintf("%v,%v", p.X, p.Y)
	case Landscape:
		s = fmt.Sprintf("%v,%v", p.Y, p.X)
	}
	return
}

func parsePoint(s string) (p Point) {
	fmt.Sscanf(s, "%v,%v", &p.X, &p.Y)
	return
}

type Triple struct {
	U, V, W Unit
}

func (t Triple) String() string {
	return fmt.Sprintf("%v,%v,%v", t.U, t.V, t.W)
}

func parseTriple(s string) (t Triple) {
	fmt.Sscanf(s, "%v,%v,%v", &t.U, &t.V, &t.W)
	return
}

type Path []Point

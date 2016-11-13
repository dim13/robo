package robo

import (
	"fmt"
	"strings"
)

const (
	MM = Unit(20.0)
	CM = 10 * MM
	M  = 100 * CM
	IN = 25.4 * MM
	FT = 12 * IN
	PT = IN / 72
)

var A4 = Point{272 * MM, 200 * MM}

type Unit float64

func (u Unit) String() string {
	if u == Unit(int(u)) {
		return fmt.Sprint(int(u))
	} else {
		return fmt.Sprintf("%.3f", u)
	}
}

func parseUnit(s string) (u Unit) {
	fmt.Sscanf(s, "%v", &u)
	return
}

type Orientation int

const (
	Portrait Orientation = iota
	Landscape
)

var orientation = Portrait

type Point struct {
	X, Y Unit
}

func (p Point) Swap() Point { return Point{X: p.Y, Y: p.X} }

func (p Point) String() (s string) {
	switch orientation {
	case Portrait:
	case Landscape:
		p = p.Swap()
	}
	return fmt.Sprintf("%v,%v", p.X, p.Y)
}

func (p Point) Scale(f Unit) Point {
	return Point{p.X * f, p.Y * f}
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

func (p Path) String() string {
	pp := make([]string, len(p))
	for i, pt := range p {
		pp[i] = pt.String()
	}
	return strings.Join(pp, ",")
}

func (p Path) Scale(f Unit) Path {
	ret := make(Path, len(p))
	for i, pt := range p {
		ret[i] = pt.Scale(f)
	}
	return ret
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

type Direction byte

const (
	Stop Direction = 1 << iota >> 1
	Down
	Up
	Right
	Left
)

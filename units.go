package main

import "fmt"

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
		return fmt.Sprintf("%.2f", u)
	}
}

func parseUnit(s string) (u Unit) {
	fmt.Sscanf(s, "%v", &u)
	return
}

type Point struct {
	X, Y Unit
}

func (p Point) String() string {
	return fmt.Sprintf("%v,%v", p.X, p.Y)
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

type Polar struct {
	R, Theta Unit
}

type Path []Point

func (p Point) Add(u Point) Point {
	return Point{p.X + u.X, p.Y + u.Y}
}

func (p Point) Sub(u Point) Point {
	return Point{p.X - u.X, p.Y - u.Y}
}

func (p Point) AddX(u Unit) Point {
	return Point{p.X + u, p.Y}
}

func (p Point) AddY(u Unit) Point {
	return Point{p.X, p.Y + u}
}

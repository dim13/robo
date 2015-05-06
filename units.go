package main

import "fmt"

const (
	MM = Unit(20.0)
	CM = 10 * MM
	DM = 10 * CM
	IN = 25.4 * MM
	PT = IN / 72
)

type Unit float64

func (u Unit) String() string {
	return fmt.Sprintf("%.2f", u)
}

type Point struct {
	X, Y Unit
}

func (p Point) String() string {
	return fmt.Sprintf("%v,%v", p.X, p.Y)
}

type Triple struct {
	U, V, W Unit
}

func (t Triple) String() string {
	return fmt.Sprintf("%v,%v,%v", t.U, t.V, t.W)
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

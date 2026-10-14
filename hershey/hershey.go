// Package hershey provides a stroke font derived from the Hershey fonts.
package hershey

// Point is a glyph coordinate: X grows downwards from the top of the line,
// Y runs in the advance direction.
type Point struct {
	X, Y int
}

type Path []Point

type Set []Path

type Glyph struct {
	S Set
	W int // advance width
}

type Font map[rune]Glyph

const (
	Height   = 72 // line pitch
	Baseline = 50
)

func Lookup(r rune) (Glyph, bool) {
	g, ok := font[r]
	return g, ok
}

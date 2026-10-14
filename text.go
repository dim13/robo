package robo

import (
	"bufio"
	"io"
	"log"
)

type Font map[rune]Glyph

type Glyph struct {
	S Set
	W Unit
}

type Set []Path

// Lookup returns the Hershey glyph for r. Glyph X grows downwards from the top
// of the line with the baseline at 50, Y is the advance direction.
func Lookup(r rune) (Glyph, bool) {
	g, ok := font[r]
	return g, ok
}

func Print(c io.Writer, in io.Reader, scale Unit) {
	var off Point

	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		font.putchar(c, scanner.Text(), scale, &off)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func (f Font) putchar(c io.Writer, s string, scale Unit, off *Point) {
	for _, ch := range s {
		gl, ok := f[ch]
		if ok {
			if off.Y+gl.W*scale >= 4000 {
				off.X += height * scale
				off.Y = 0
			}
			off.Origin(c)
			for _, p := range gl.S {
				p.Scale(scale).Line(c)
				//p.Scale(scale).Curve(c, 0)
			}
			off.Y += gl.W * scale
		} else if ch == '\t' {
			tab := 8 * f['W'].W * scale // widest char
			pos := int(off.Y / tab)
			off.Y = Unit(pos+1) * tab
		}
	}
	off.X += height * scale
	off.Y = 0
}

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
			off.Offset(c)
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

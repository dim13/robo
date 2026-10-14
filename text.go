package robo

import (
	"bufio"
	"io"
	"log"

	"github.com/dim13/robo/hershey"
)

func Print(c io.Writer, in io.Reader, scale Unit) {
	var off Point

	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		putchar(c, scanner.Text(), scale, &off)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func putchar(c io.Writer, s string, scale Unit, off *Point) {
	height := hershey.Height * scale
	for _, ch := range s {
		gl, ok := hershey.Lookup(ch)
		w := Unit(gl.W) * scale
		if ok {
			if off.Y+w >= 4000 {
				off.X += height
				off.Y = 0
			}
			off.Origin(c)
			for _, p := range gl.S {
				path(p).Scale(scale).Line(c)
			}
			off.Y += w
		} else if ch == '\t' {
			wide, _ := hershey.Lookup('W')
			tab := 8 * Unit(wide.W) * scale // widest char
			pos := int(off.Y / tab)
			off.Y = Unit(pos+1) * tab
		}
	}
	off.X += height
	off.Y = 0
}

func path(p hershey.Path) Path {
	r := make(Path, len(p))
	for i, q := range p {
		r[i] = Point{X: Unit(q.X), Y: Unit(q.Y)}
	}
	return r
}

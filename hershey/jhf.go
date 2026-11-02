package hershey

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// font.jhf is in Hershey JHF format: one glyph per line, numbered by its
// Unicode code point, coordinates as letters relative to 'R', " R" lifts the
// pen. gen.go adds the non-ASCII Hershey glyphs, joins strokes (see
// font_test.go) and orders them for the shortest pen-up travel.
//
//go:generate go run gen.go
//go:embed font.jhf
var jhf []byte

var font = mustParse(jhf)

const (
	scale = 2  // font.jhf holds plain Hershey coordinates
	top   = 32 // X of JHF y 0 (the middle of the line)
)

func parse(r io.Reader) (Font, error) {
	f := make(Font)
	var code rune
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		if len(line) < 10 {
			return nil, fmt.Errorf("glyph %q: short line", code)
		}
		c, err := strconv.Atoi(strings.TrimSpace(line[:5]))
		if err != nil {
			return nil, fmt.Errorf("after %q: %w", code, err)
		}
		code = rune(c)
		n, err := strconv.Atoi(strings.TrimSpace(line[5:8]))
		if err != nil {
			return nil, fmt.Errorf("glyph %q: %w", code, err)
		}
		data := line[8:]
		for len(data) < 2*n && sc.Scan() { // long glyphs wrap
			data += sc.Text()
		}
		if len(data) < 2*n {
			return nil, fmt.Errorf("glyph %q: truncated", code)
		}
		f[code] = glyph(data[:2*n])
	}
	return f, sc.Err()
}

func glyph(d string) Glyph {
	left, right := int(d[0])-'R', int(d[1])-'R'
	g := Glyph{W: scale * (right - left)}
	var p Path
	for i := 2; i+1 < len(d); i += 2 {
		if d[i:i+2] == " R" {
			if len(p) > 0 {
				g.S = append(g.S, p)
			}
			p = nil
			continue
		}
		p = append(p, Point{
			X: scale*(int(d[i+1])-'R') + top,
			Y: scale * (int(d[i]) - 'R' - left),
		})
	}
	if len(p) > 0 {
		g.S = append(g.S, p)
	}
	return g
}

func mustParse(b []byte) Font {
	f, err := parse(bytes.NewReader(b))
	if err != nil {
		panic(err) // embedded data is broken
	}
	return f
}

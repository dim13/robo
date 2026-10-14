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

// font.jhf is in Hershey JHF format: one glyph per line, numbered from ' ',
// coordinates as letters relative to 'R', " R" lifts the pen. Strokes are
// joined and ordered for short pen travel (see font_test.go).
//
//go:embed font.jhf
var jhf []byte

var font = mustParse(jhf)

// top maps JHF y (0 = middle of the line) to X (0 = top of the line).
const top = 32

func parse(r io.Reader) (Font, error) {
	f := make(Font)
	code := ' '
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		if len(line) < 10 {
			return nil, fmt.Errorf("glyph %q: short line", code)
		}
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
		code++
	}
	return f, sc.Err()
}

func glyph(d string) Glyph {
	left, right := int(d[0])-'R', int(d[1])-'R'
	g := Glyph{W: right - left}
	var p Path
	for i := 2; i+1 < len(d); i += 2 {
		if d[i:i+2] == " R" {
			if len(p) > 0 {
				g.S = append(g.S, p)
			}
			p = nil
			continue
		}
		p = append(p, Point{X: int(d[i+1]) - 'R' + top, Y: int(d[i]) - 'R' - left})
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

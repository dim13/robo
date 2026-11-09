// Plot an HP-GL file.
package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/dim13/robo"
	"github.com/dim13/robo/hershey"
	"github.com/dim13/robo/usb"
)

const (
	plu = robo.MM / 40 // HP-GL plotter unit, 0.025 mm
	cm  = 400          // plotter units per cm
	cap = 42           // Hershey cap height
)

type point struct{ x, y float64 }

var siA4 = point{0.187 * cm, 0.269 * cm} // default SI for A4

func (q point) robo() robo.Point {
	return robo.Point{X: robo.Unit(q.x) * plu, Y: robo.Unit(q.y) * plu}
}

type plotter struct {
	w       io.Writer
	pos     point
	down    bool
	rel     bool
	skipped map[string]int
	missing int   // characters without a glyph
	size    point // character width and cap height (SI)
	dir     point // label direction, unit vector
}

func newPlotter(w io.Writer) *plotter {
	p := &plotter{w: w, skipped: make(map[string]int)}
	p.reset()
	return p
}

// reset sets the IN defaults; SI defaults to the A4 size.
func (p *plotter) reset() {
	p.down, p.rel = false, false
	p.size = siA4
	p.dir = point{1, 0}
}

func (p *plotter) to(q point) {
	if p.down {
		q.robo().Draw(p.w)
	} else {
		q.robo().Move(p.w)
	}
	p.pos = q
}

func (p *plotter) points(a []float64) {
	for i := 0; i+1 < len(a); i += 2 {
		q := point{a[i], a[i+1]}
		if p.rel {
			q = point{p.pos.x + q.x, p.pos.y + q.y}
		}
		p.to(q)
	}
}

// arc draws from the current position around c by sweep degrees
// (counterclockwise if positive) in chords of at most chord degrees.
func (p *plotter) arc(c point, sweep, chord float64) {
	if chord <= 0 {
		chord = 5
	}
	r := math.Hypot(p.pos.x-c.x, p.pos.y-c.y)
	a0 := math.Atan2(p.pos.y-c.y, p.pos.x-c.x)
	n := max(int(math.Ceil(math.Abs(sweep)/chord)), 1)
	for i := 1; i <= n; i++ {
		a := a0 + sweep*float64(i)/float64(n)*math.Pi/180
		p.to(point{c.x + r*math.Cos(a), c.y + r*math.Sin(a)})
	}
}

func (p *plotter) circle(r, chord float64) {
	c, down := p.pos, p.down
	p.down = false
	p.to(point{c.x + r, c.y})
	p.down = true
	p.arc(c, 360, chord)
	p.down = false
	p.to(c)
	p.down = down
}

// label draws s with the Hershey font. The pen ends after the last
// character in its previous up/down state; CR returns to the start of the
// line, LF moves one line (twice the cap height) down.
func (p *plotter) label(s string) {
	sy := p.size.y / cap
	sx := sy * p.size.x / p.size.y * siA4.y / siA4.x // default SI keeps the font aspect
	run, up := p.dir, point{-p.dir.y, p.dir.x}
	at := func(o point, u, v float64) point {
		return point{o.x + u*run.x + v*up.x, o.y + u*run.y + v*up.y}
	}
	down, start := p.down, p.pos
	defer func() {
		p.down = false
		p.to(p.pos)
		p.down = down
	}()
	for _, r := range s {
		switch r {
		case '\r':
			p.pos = start
			continue
		case '\n':
			start = at(start, 0, -2*p.size.y)
			p.pos = at(p.pos, 0, -2*p.size.y)
			continue
		}
		g, ok := hershey.Lookup(r)
		if !ok {
			p.missing++
			continue
		}
		o := p.pos
		for _, stroke := range g.S {
			for i, q := range stroke {
				p.down = i > 0
				p.to(at(o, float64(q.Y)*sx, float64(hershey.Baseline-q.X)*sy))
			}
		}
		p.pos = at(o, float64(g.W)*sx, 0)
	}
}

func (p *plotter) op(cmd, s string) {
	a := numbers(s)
	arg := func(i int) float64 {
		if i < len(a) {
			return a[i]
		}
		return 0
	}
	switch cmd {
	case "IN", "DF":
		p.reset()
	case "LB":
		p.label(s)
	case "SI":
		if len(a) > 1 {
			p.size = point{a[0] * cm, a[1] * cm}
		} else {
			p.size = siA4
		}
	case "DI":
		if n := math.Hypot(arg(0), arg(1)); n > 0 {
			p.dir = point{a[0] / n, a[1] / n}
		} else {
			p.dir = point{1, 0}
		}
	case "PU":
		p.down = false
		p.points(a)
	case "PD":
		p.down = true
		p.points(a)
	case "PA":
		p.rel = false
		p.points(a)
	case "PR":
		p.rel = true
		p.points(a)
	case "CI":
		if len(a) > 0 {
			p.circle(a[0], arg(1))
		}
	case "AA":
		if len(a) > 2 {
			p.arc(point{a[0], a[1]}, a[2], arg(3))
		}
	case "AR":
		if len(a) > 2 {
			p.arc(point{p.pos.x + a[0], p.pos.y + a[1]}, a[2], arg(3))
		}
	case "VS":
		// cm/s, the same scale as GP-GL "!"
		if len(a) > 0 {
			robo.Unit(min(max(math.Round(a[0]), 1), 10)).Speed(p.w)
		}
	case "SP", "LT", "PS", "RO", "IP", "SC":
	default:
		p.skipped[cmd]++
	}
}

func isLetter(c byte) bool {
	return c|0x20 >= 'a' && c|0x20 <= 'z'
}

func numbers(s string) []float64 {
	var a []float64
	for _, f := range strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || unicode.IsSpace(r)
	}) {
		if v, err := strconv.ParseFloat(f, 64); err == nil {
			a = append(a, v)
		}
	}
	return a
}

// parse splits HP-GL into two-letter mnemonics and their parameters.
// A label (LB) runs up to its ETX terminator.
func parse(b []byte, f func(cmd, s string)) {
	for i := 0; i+1 < len(b); {
		if !isLetter(b[i]) || !isLetter(b[i+1]) {
			i++
			continue
		}
		cmd := strings.ToUpper(string(b[i : i+2]))
		i += 2
		if cmd == "LB" {
			n := strings.IndexByte(string(b[i:]), robo.ETX)
			if n < 0 {
				n = len(b) - i
			}
			f(cmd, string(b[i:i+n]))
			i += n
			continue
		}
		j := i
		for j < len(b) && !isLetter(b[j]) && b[j] != ';' {
			j++
		}
		f(cmd, string(b[i:j]))
		i = j
	}
}

func plot(w io.Writer, b []byte) *plotter {
	p := newPlotter(w)
	parse(b, p.op)
	p.down = false
	p.to(point{})
	return p
}

func main() {
	var (
		media = flag.Int("media", 113, "media preset (113 = pen)")
		dry   = flag.Bool("n", false, "write GP-GL to stdout instead of plotting")
	)
	flag.Parse()

	in := io.Reader(os.Stdin)
	if flag.NArg() > 0 {
		f, err := os.Open(flag.Arg(0))
		if err != nil {
			log.Fatal("open: ", err)
		}
		defer f.Close()
		in = f
	}
	b, err := io.ReadAll(in)
	if err != nil {
		log.Fatal("read: ", err)
	}

	var p *plotter
	if *dry {
		w := bufio.NewWriter(os.Stdout)
		p = plot(w, b)
		w.Flush()
	} else {
		dev, err := usb.Open()
		if err != nil {
			log.Fatal("open: ", err)
		}
		defer dev.Close()
		if err := robo.Initialize(dev, *media, robo.Portrait); err != nil {
			log.Fatal("init: ", err)
		}
		p = plot(dev, b)
	}
	if p.missing > 0 {
		log.Printf("%d characters have no Hershey glyph", p.missing)
	}
	for cmd, n := range p.skipped {
		log.Printf("skipped %d %q", n, cmd)
	}
}

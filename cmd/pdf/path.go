package main

// matrix is a PDF transformation [a b c d e f].
type matrix [6]float64

var identity = matrix{1, 0, 0, 1, 0, 0}

// mul returns m followed by n.
func (m matrix) mul(n matrix) matrix {
	return matrix{
		m[0]*n[0] + m[1]*n[2],
		m[0]*n[1] + m[1]*n[3],
		m[2]*n[0] + m[3]*n[2],
		m[2]*n[1] + m[3]*n[3],
		m[4]*n[0] + m[5]*n[2] + n[4],
		m[4]*n[1] + m[5]*n[3] + n[5],
	}
}

func (m matrix) apply(x, y float64) point {
	return point{m[0]*x + m[2]*y + m[4], m[1]*x + m[3]*y + m[5]}
}

type point struct{ x, y float64 }

// segment is a line to p[0], or a cubic Bézier through p[0], p[1] to p[2].
type segment []point

type subpath struct {
	start point
	segs  []segment
}

// saved is the graphics state kept by q and restored by Q.
type saved struct {
	ctm matrix
	textState
}

// painter collects the stroked and filled outlines of a page.
type painter struct {
	ctm     matrix
	stack   []saved
	cur     []subpath
	pos     point
	paths   []subpath
	skipped map[string]int
	missing int // characters without a glyph
	fonts   map[string]*font
	textState
}

func newPainter(ctm matrix) *painter {
	return &painter{
		ctm:       ctm,
		skipped:   make(map[string]int),
		fonts:     make(map[string]*font),
		textState: textState{tz: 100},
	}
}

func (p *painter) moveTo(x, y float64) {
	p.pos = p.ctm.apply(x, y)
	p.cur = append(p.cur, subpath{start: p.pos})
}

func (p *painter) add(s segment) {
	if len(p.cur) == 0 {
		p.cur = append(p.cur, subpath{start: p.pos})
	}
	last := &p.cur[len(p.cur)-1]
	last.segs = append(last.segs, s)
	p.pos = s[len(s)-1]
}

func (p *painter) closePath() {
	if len(p.cur) == 0 {
		return
	}
	start := p.cur[len(p.cur)-1].start
	if p.pos != start {
		p.add(segment{start})
	}
	p.pos = start
}

// op executes one content stream operator with its numeric operands.
func (p *painter) op(name string, a []float64) {
	switch name {
	case "q":
		p.stack = append(p.stack, saved{p.ctm, p.textState})
	case "Q":
		if n := len(p.stack); n > 0 {
			p.ctm, p.textState = p.stack[n-1].ctm, p.stack[n-1].textState
			p.stack = p.stack[:n-1]
		}
	case "cm":
		if len(a) == 6 {
			p.ctm = matrix(a).mul(p.ctm)
		}
	case "m":
		if len(a) == 2 {
			p.moveTo(a[0], a[1])
		}
	case "l":
		if len(a) == 2 {
			p.add(segment{p.ctm.apply(a[0], a[1])})
		}
	case "c":
		if len(a) == 6 {
			p.add(segment{p.ctm.apply(a[0], a[1]), p.ctm.apply(a[2], a[3]), p.ctm.apply(a[4], a[5])})
		}
	case "v":
		if len(a) == 4 {
			p.add(segment{p.pos, p.ctm.apply(a[0], a[1]), p.ctm.apply(a[2], a[3])})
		}
	case "y":
		if len(a) == 4 {
			end := p.ctm.apply(a[2], a[3])
			p.add(segment{p.ctm.apply(a[0], a[1]), end, end})
		}
	case "h":
		p.closePath()
	case "re":
		if len(a) == 4 {
			x, y, w, h := a[0], a[1], a[2], a[3]
			p.moveTo(x, y)
			p.add(segment{p.ctm.apply(x+w, y)})
			p.add(segment{p.ctm.apply(x+w, y+h)})
			p.add(segment{p.ctm.apply(x, y+h)})
			p.closePath()
		}
	case "s", "b", "b*":
		p.closePath()
		p.paint()
	case "S", "f", "F", "f*", "B", "B*":
		p.paint()
	case "n":
		p.cur = nil
	case "Do", "BI", "sh":
		p.skipped[name]++
	}
}

func (p *painter) paint() {
	for _, sp := range p.cur {
		if len(sp.segs) > 0 {
			p.paths = append(p.paths, sp)
		}
	}
	p.cur = nil
}

//go:build ignore

// Gen extends font.jhf with the non-ASCII glyphs of the Hershey occidental
// set, using the Unicode mapping of
// github.com/dim13/hershey/docs/unicodemap.csv.
//
//	go run gen.go -src ../../hershey
package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"math"
	"math/bits"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

type pt struct{ x, y int }

// rank prefers simplex glyphs, then plain, then complex.
func rank(hid int) int {
	switch {
	case hid >= 500 && hid < 1000:
		return 0
	case hid < 500:
		return 1
	case hid >= 2000 && hid < 3000:
		return 2
	default:
		return 3
	}
}

func decode(d string) (left, right int, s [][]pt) {
	left, right = int(d[0])-'R', int(d[1])-'R'
	var p []pt
	for i := 2; i+1 < len(d); i += 2 {
		if d[i:i+2] == " R" {
			if len(p) > 0 {
				s = append(s, p)
			}
			p = nil
			continue
		}
		p = append(p, pt{int(d[i]) - 'R', int(d[i+1]) - 'R'})
	}
	if len(p) > 0 {
		s = append(s, p)
	}
	return left, right, s
}

func encode(code, left, right int, s [][]pt) string {
	var b strings.Builder
	n := 1
	for i, p := range s {
		if i > 0 {
			b.WriteString(" R")
			n++
		}
		for _, q := range p {
			b.WriteByte(byte(q.x + 'R'))
			b.WriteByte(byte(q.y + 'R'))
			n++
		}
	}
	return fmt.Sprintf("%5d%3d%c%c%s", code, n, left+'R', right+'R', b.String())
}

func closed(p []pt) bool {
	return p[0] == p[len(p)-1]
}

func reversed(p []pt) []pt {
	r := slices.Clone(p)
	slices.Reverse(r)
	return r
}

// merge joins b onto a if they share an end, or an end of one lies on the
// other's closed loop.
func merge(a, b []pt) ([]pt, bool) {
	cat := func(x, y []pt) []pt {
		return append(slices.Clone(x), y[1:]...)
	}
	switch {
	case a[len(a)-1] == b[0]:
		return cat(a, b), true
	case a[len(a)-1] == b[len(b)-1]:
		return cat(a, reversed(b)), true
	case a[0] == b[len(b)-1]:
		return cat(b, a), true
	case a[0] == b[0]:
		return cat(reversed(b), a), true
	}
	if closed(b) {
		for k := range b {
			rot := append(slices.Clone(b[k:]), b[1:k+1]...)
			switch b[k] {
			case a[len(a)-1]:
				return cat(a, rot), true
			case a[0]:
				return cat(rot, a), true
			}
		}
	}
	return nil, false
}

func join(s [][]pt) [][]pt {
	for i := 0; i < len(s); i++ {
		for j := i + 1; j < len(s); j++ {
			m, ok := merge(s[i], s[j])
			if !ok {
				m, ok = merge(s[j], s[i])
			}
			if ok {
				s[i] = m
				s = slices.Delete(s, j, j+1)
				i, j = 0, 0
			}
		}
	}
	return s
}

// base is the baseline in JHF y; the pen enters a glyph at (left, base)
// and leaves it at (right, base).
const base = 9

func dist(a, b pt) float64 {
	return math.Hypot(float64(a.x-b.x), float64(a.y-b.y))
}

// travel is the pen-up distance to plot s from in to out.
func travel(s [][]pt, in, out pt) float64 {
	var d float64
	for _, p := range s {
		d += dist(in, p[0])
		in = p[len(p)-1]
	}
	return d + dist(in, out)
}

// order finds the stroke order, directions and loop start points with the
// shortest pen-up travel from in to out (Held-Karp, fine for a few strokes).
func order(s [][]pt, in, out pt) [][]pt {
	type opt struct {
		stroke     int
		start, end pt
		path       []pt
	}
	var opts []opt
	for i, p := range s {
		switch {
		case closed(p) && len(p) > 2:
			for k := range len(p) - 1 {
				rot := append(slices.Clone(p[k:]), p[1:k+1]...)
				opts = append(opts, opt{i, p[k], p[k], rot})
			}
		default:
			opts = append(opts, opt{i, p[0], p[len(p)-1], p})
			opts = append(opts, opt{i, p[len(p)-1], p[0], reversed(p)})
		}
	}
	n, full := len(opts), 1<<len(s)-1
	cost := make([][]float64, full+1)
	prev := make([][]int, full+1)
	for m := range cost {
		cost[m] = make([]float64, n)
		prev[m] = make([]int, n)
		for j := range cost[m] {
			cost[m][j] = math.Inf(1)
		}
	}
	for j, o := range opts {
		cost[1<<o.stroke][j] = dist(in, o.start)
		prev[1<<o.stroke][j] = -1
	}
	for m := 1; m <= full; m++ {
		for j, o := range opts {
			c := cost[m][j]
			if math.IsInf(c, 1) || m&(1<<o.stroke) == 0 {
				continue
			}
			for k, q := range opts {
				if m&(1<<q.stroke) != 0 {
					continue
				}
				mm := m | 1<<q.stroke
				if d := c + dist(o.end, q.start); d < cost[mm][k] {
					cost[mm][k] = d
					prev[mm][k] = j
				}
			}
		}
	}
	best, bc := -1, math.Inf(1)
	for j, o := range opts {
		if d := cost[full][j] + dist(o.end, out); d < bc {
			best, bc = j, d
		}
	}
	r := make([][]pt, len(s))
	for m, j := full, best; j >= 0; {
		o := opts[j]
		r[bits.OnesCount(uint(m))-1] = o.path
		m, j = m&^(1<<o.stroke), prev[m][j]
	}
	return r
}

func main() {
	src := flag.String("src", "../../hershey", "github.com/dim13/hershey checkout")
	flag.Parse()

	var out []string
	var before, after float64
	emit := func(code, left, right int, s [][]pt) {
		in, end := pt{left, base}, pt{right, base}
		s = join(s)
		before += travel(s, in, end)
		s = order(s, in, end)
		after += travel(s, in, end)
		out = append(out, encode(code, left, right, s))
	}

	// keep the ASCII glyphs
	old, err := os.ReadFile("font.jhf")
	if err != nil {
		log.Fatal(err)
	}
	for _, line := range strings.Split(strings.TrimSuffix(string(old), "\n"), "\n") {
		if code, _ := strconv.Atoi(strings.TrimSpace(line[:5])); code < 128 {
			left, right, s := decode(line[8:])
			emit(code, left, right, s)
		}
	}

	f, err := os.Open(filepath.Join(*src, "docs", "unicodemap.csv"))
	if err != nil {
		log.Fatal(err)
	}
	rows, err := csv.NewReader(f).ReadAll()
	f.Close()
	if err != nil {
		log.Fatal(err)
	}
	best := make(map[int]int) // code point to Hershey id
	for _, r := range rows[1:] {
		hid, _ := strconv.Atoi(r[0])
		code, _ := strconv.Atoi(r[1])
		if code < 128 || unicode.Is(unicode.Mn, rune(code)) {
			continue // combining marks need placement over the base glyph
		}
		if code > 99999 {
			continue // JHF has five digits for the glyph number
		}
		if h, ok := best[code]; !ok || rank(hid) < rank(h) || rank(hid) == rank(h) && hid < h {
			best[code] = hid
		}
	}

	data := make(map[int]string)
	f, err = os.Open(filepath.Join(*src, "data", "hershey"))
	if err != nil {
		log.Fatal(err)
	}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		hid, _ := strconv.Atoi(strings.TrimSpace(line[:5]))
		data[hid] = line[8:]
	}
	f.Close()
	if err := sc.Err(); err != nil {
		log.Fatal(err)
	}

	codes := make([]int, 0, len(best))
	for code := range best {
		codes = append(codes, code)
	}
	slices.Sort(codes)
	for _, code := range codes {
		d, ok := data[best[code]]
		if !ok {
			log.Fatalf("U+%04X: no glyph %d", code, best[code])
		}
		left, right, s := decode(d)
		emit(code, left, right, s)
	}
	log.Printf("pen-up travel %.0f → %.0f", before, after)
	if err := os.WriteFile("font.jhf", []byte(strings.Join(out, "\n")+"\n"), 0644); err != nil {
		log.Fatal(err)
	}
}

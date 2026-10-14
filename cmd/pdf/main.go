// Plot vector outlines of a PDF page.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/dim13/robo"
	"rsc.io/pdf"
)

// inherited looks up a page attribute, following Parent links.
func inherited(v pdf.Value, key string) pdf.Value {
	for ; !v.IsNull(); v = v.Key("Parent") {
		if r := v.Key(key); !r.IsNull() {
			return r
		}
	}
	return pdf.Value{}
}

func readPage(file string, n int, text bool) (p *painter, box [4]float64, err error) {
	// rsc.io/pdf panics on streams it cannot decode
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()
	r, err := pdf.Open(file)
	if err != nil {
		return nil, box, err
	}
	if n < 1 || n > r.NumPage() {
		return nil, box, fmt.Errorf("page %d of %d", n, r.NumPage())
	}
	page := r.Page(n)
	mb := inherited(page.V, "MediaBox")
	for i := range box {
		box[i] = mb.Index(i).Float64()
	}
	box, ctm := rotate(box, int(inherited(page.V, "Rotate").Int64()))
	p = newPainter(ctm)
	fonts := inherited(page.V, "Resources").Key("Font")
	do := func(stk *pdf.Stack, op string) {
		v := make([]pdf.Value, stk.Len())
		a := make([]float64, len(v))
		for i := len(v) - 1; i >= 0; i-- {
			v[i] = stk.Pop()
			a[i] = v[i].Float64()
		}
		p.op(op, a)
		if text {
			p.textOp(op, v, fonts)
		}
	}
	contents := page.V.Key("Contents")
	if contents.Kind() == pdf.Array {
		for i := range contents.Len() {
			pdf.Interpret(contents.Index(i), do)
		}
	} else {
		pdf.Interpret(contents, do)
	}
	return p, box, nil
}

// toRobo maps PDF points to plotter units. The long side of the page runs
// along X (media feed): a portrait page feeds in top edge first, a
// landscape page left edge first.
// ponytail: fixed rotation, add a flag if output comes out mirrored.
func toRobo(pt point, box [4]float64) robo.Point {
	if landscape(box) {
		return robo.Point{
			X: robo.Unit(pt.x-box[0]) * robo.PT,
			Y: robo.Unit(pt.y-box[1]) * robo.PT,
		}
	}
	return robo.Point{
		X: robo.Unit(box[3]-pt.y) * robo.PT,
		Y: robo.Unit(pt.x-box[0]) * robo.PT,
	}
}

// rotate applies the page /Rotate (clockwise degrees) and returns the
// rotated box with the matrix that maps content into it.
func rotate(box [4]float64, deg int) ([4]float64, matrix) {
	x0, y0, x1, y1 := box[0], box[1], box[2], box[3]
	switch (deg%360 + 360) % 360 {
	case 90:
		return [4]float64{0, 0, y1 - y0, x1 - x0}, matrix{0, -1, 1, 0, -y0, x1}
	case 180:
		return [4]float64{0, 0, x1 - x0, y1 - y0}, matrix{-1, 0, 0, -1, x1, y1}
	case 270:
		return [4]float64{0, 0, y1 - y0, x1 - x0}, matrix{0, 1, -1, 0, y1, -x0}
	default:
		return box, identity
	}
}

func landscape(box [4]float64) bool {
	return box[2]-box[0] > box[3]-box[1]
}

func plot(w io.Writer, paths []subpath, box [4]float64) {
	for _, sp := range paths {
		cur := toRobo(sp.start, box)
		cur.Move(w)
		for _, s := range sp.segs {
			switch len(s) {
			case 1:
				cur = toRobo(s[0], box)
				cur.Draw(w)
			case 3:
				bz := robo.Path{cur, toRobo(s[0], box), toRobo(s[1], box), toRobo(s[2], box)}
				bz.Bezier(w, 1)
				cur = bz[3]
			}
		}
	}
}

func main() {
	var (
		page  = flag.Int("page", 1, "page number")
		media = flag.Int("media", 113, "media preset (113 = pen)")
		dry   = flag.Bool("n", false, "write GP-GL to stdout instead of plotting")
		text  = flag.Bool("text", true, "plot text in a Hershey font")
	)
	flag.Parse()
	if flag.NArg() != 1 {
		log.Fatal("usage: pdf [-page n] [-media id] [-text=false] [-n] file.pdf")
	}

	p, box, err := readPage(flag.Arg(0), *page, *text)
	if err != nil {
		log.Fatal("read: ", err)
	}
	if p.missing > 0 {
		log.Printf("%d characters have no Hershey glyph", p.missing)
	}
	for op, n := range p.skipped {
		log.Printf("skipped %d %q (images, shadings and forms are not plotted)", n, op)
	}
	size := toRobo(point{box[2], box[1]}, box)
	if landscape(box) {
		size = toRobo(point{box[2], box[3]}, box)
	}
	if size.Y > 200*robo.MM {
		log.Printf("page is %v mm wide, clipped to 200 mm", size.Y/robo.MM)
		size.Y = 200 * robo.MM
	}

	if *dry {
		w := bufio.NewWriter(os.Stdout)
		defer w.Flush()
		size.UpperRight(w)
		plot(w, p.paths, box)
		return
	}

	dev, err := robo.Open()
	if err != nil {
		log.Fatal("open: ", err)
	}
	defer dev.Close()
	defer robo.Home(dev)
	if err := robo.Initialize(dev, *media, robo.Portrait); err != nil {
		log.Fatal("init: ", err)
	}
	size.UpperRight(dev)
	plot(dev, p.paths, box)
}

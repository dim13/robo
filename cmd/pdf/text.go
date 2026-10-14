package main

import (
	"encoding/hex"
	"io"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf16"

	"github.com/dim13/robo/hershey"
	"rsc.io/pdf"
)

// em is the Hershey em: cap height 42 of about 60 units.
const em = 60

// font maps character codes to text and advance widths (1/1000 em).
type font struct {
	twoByte bool
	unicode map[int]string
	widths  map[int]float64
	dflt    float64
}

func (f *font) codes(s string) []int {
	var c []int
	if f.twoByte {
		for i := 0; i+1 < len(s); i += 2 {
			c = append(c, int(s[i])<<8|int(s[i+1]))
		}
		return c
	}
	for i := range len(s) {
		c = append(c, int(s[i]))
	}
	return c
}

func (f *font) text(code int) string {
	if s, ok := f.unicode[code]; ok {
		return s
	}
	if f.twoByte {
		return ""
	}
	return string(rune(code)) // ponytail: Latin-1 for simple fonts, ignores /Differences
}

func (f *font) width(code int) float64 {
	if w, ok := f.widths[code]; ok {
		return w
	}
	return f.dflt
}

// textState is the PDF text state of a painter.
type textState struct {
	tm, tlm                   matrix
	font                      *font
	size, leading, cs, ws, tz float64
	rise                      float64
}

func (p *painter) beginText() {
	p.tm, p.tlm = identity, identity
}

func (p *painter) moveText(tx, ty float64) {
	p.tlm = matrix{1, 0, 0, 1, tx, ty}.mul(p.tlm)
	p.tm = p.tlm
}

func (p *painter) setTextMatrix(m matrix) {
	p.tm, p.tlm = m, m
}

func (p *painter) nextLine() {
	p.moveText(0, -p.leading)
}

// show draws the glyphs of a string and advances the text matrix.
func (p *painter) show(s string) {
	f := p.font
	if f == nil {
		return
	}
	th := p.tz / 100
	for _, code := range f.codes(s) {
		trm := matrix{p.size * th, 0, 0, p.size, 0, p.rise}.mul(p.tm).mul(p.ctm)
		for _, r := range f.text(code) {
			p.glyph(r, trm)
		}
		tx := f.width(code)/1000*p.size + p.cs
		if code == ' ' && !f.twoByte {
			tx += p.ws
		}
		p.tm = matrix{1, 0, 0, 1, tx * th, 0}.mul(p.tm)
	}
}

// adjust applies a TJ position adjustment in 1/1000 em.
func (p *painter) adjust(n float64) {
	p.tm = matrix{1, 0, 0, 1, -n / 1000 * p.size * p.tz / 100, 0}.mul(p.tm)
}

func (p *painter) glyph(r rune, trm matrix) {
	g, ok := hershey.Lookup(r)
	if !ok {
		if r != ' ' {
			p.missing++
		}
		return
	}
	for _, stroke := range g.S {
		if len(stroke) < 2 {
			continue
		}
		sp := subpath{start: glyphPoint(stroke[0], trm)}
		for _, q := range stroke[1:] {
			sp.segs = append(sp.segs, segment{glyphPoint(q, trm)})
		}
		p.paths = append(p.paths, sp)
	}
}

func glyphPoint(q hershey.Point, trm matrix) point {
	return trm.apply(float64(q.Y)/em, float64(hershey.Baseline-q.X)/em)
}

// textOp executes a text operator; operands are raw PDF values.
func (p *painter) textOp(name string, a []pdf.Value, fonts pdf.Value) {
	num := func(i int) float64 {
		return a[i].Float64()
	}
	switch name {
	case "BT":
		p.beginText()
	case "Tf":
		if len(a) == 2 {
			p.font = p.loadFont(fonts.Key(a[0].Name()))
			p.size = num(1)
		}
	case "Tc":
		if len(a) == 1 {
			p.cs = num(0)
		}
	case "Tw":
		if len(a) == 1 {
			p.ws = num(0)
		}
	case "Tz":
		if len(a) == 1 {
			p.tz = num(0)
		}
	case "TL":
		if len(a) == 1 {
			p.leading = num(0)
		}
	case "Ts":
		if len(a) == 1 {
			p.rise = num(0)
		}
	case "Td":
		if len(a) == 2 {
			p.moveText(num(0), num(1))
		}
	case "TD":
		if len(a) == 2 {
			p.leading = -num(1)
			p.moveText(num(0), num(1))
		}
	case "Tm":
		if len(a) == 6 {
			p.setTextMatrix(matrix{num(0), num(1), num(2), num(3), num(4), num(5)})
		}
	case "T*":
		p.nextLine()
	case "Tj":
		if len(a) == 1 {
			p.show(a[0].RawString())
		}
	case "'":
		if len(a) == 1 {
			p.nextLine()
			p.show(a[0].RawString())
		}
	case "\"":
		if len(a) == 3 {
			p.ws, p.cs = num(0), num(1)
			p.nextLine()
			p.show(a[2].RawString())
		}
	case "TJ":
		if len(a) == 1 {
			for i := range a[0].Len() {
				v := a[0].Index(i)
				if v.Kind() == pdf.String {
					p.show(v.RawString())
				} else {
					p.adjust(v.Float64())
				}
			}
		}
	}
}

func (p *painter) loadFont(v pdf.Value) *font {
	key := v.String()
	if f, ok := p.fonts[key]; ok {
		return f
	}
	f := &font{unicode: toUnicode(v.Key("ToUnicode")), widths: make(map[int]float64)}
	if v.Key("Subtype").Name() == "Type0" {
		f.twoByte = true
		d := v.Key("DescendantFonts").Index(0)
		f.dflt = 1000
		if dw := d.Key("DW"); !dw.IsNull() {
			f.dflt = dw.Float64()
		}
		w := d.Key("W")
		for i := 0; i+1 < w.Len(); {
			c := int(w.Index(i).Int64())
			if next := w.Index(i + 1); next.Kind() == pdf.Array {
				for j := range next.Len() {
					f.widths[c+j] = next.Index(j).Float64()
				}
				i += 2
			} else {
				for k := c; k <= int(next.Int64()); k++ {
					f.widths[k] = w.Index(i + 2).Float64()
				}
				i += 3
			}
		}
	} else {
		first := int(v.Key("FirstChar").Int64())
		ws := v.Key("Widths")
		for i := range ws.Len() {
			f.widths[first+i] = ws.Index(i).Float64()
		}
		f.dflt = v.Key("FontDescriptor").Key("MissingWidth").Float64()
	}
	p.fonts[key] = f
	return f
}

// toUnicode reads the bfchar and bfrange mappings of a ToUnicode CMap.
func toUnicode(cmap pdf.Value) map[int]string {
	if cmap.IsNull() {
		return map[int]string{}
	}
	rc := cmap.Reader()
	defer rc.Close()
	b, err := io.ReadAll(rc)
	if err != nil {
		return map[int]string{}
	}
	return parseCMap(string(b))
}

var (
	cmapSection = regexp.MustCompile(`(?s)begin(bfchar|bfrange)(.*?)endbf`)
	cmapToken   = regexp.MustCompile(`<[0-9A-Fa-f\s]*>|\[|\]`)
)

func parseCMap(s string) map[int]string {
	m := make(map[int]string)
	for _, sec := range cmapSection.FindAllStringSubmatch(s, -1) {
		tok := cmapToken.FindAllString(sec[2], -1)
		switch sec[1] {
		case "bfchar":
			for i := 0; i+1 < len(tok); i += 2 {
				m[hexCode(tok[i])] = utf16be(hexBytes(tok[i+1]))
			}
		case "bfrange":
			for i := 0; i+2 < len(tok); {
				lo, hi := hexCode(tok[i]), hexCode(tok[i+1])
				if tok[i+2] == "[" {
					j := i + 3
					for c := lo; j < len(tok) && tok[j] != "]"; c, j = c+1, j+1 {
						m[c] = utf16be(hexBytes(tok[j]))
					}
					i = j + 1
					continue
				}
				u := []rune(utf16be(hexBytes(tok[i+2])))
				for c := lo; c <= hi && len(u) > 0; c++ {
					m[c] = string(u)
					u[len(u)-1]++
				}
				i += 3
			}
		}
	}
	return m
}

func hexBytes(t string) string {
	h := strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) || r == '<' || r == '>' {
			return -1
		}
		return r
	}, t)
	if len(h)%2 == 1 {
		h += "0"
	}
	b, _ := hex.DecodeString(h)
	return string(b)
}

func hexCode(t string) int {
	var c int
	for _, b := range []byte(hexBytes(t)) {
		c = c<<8 | int(b)
	}
	return c
}

func utf16be(s string) string {
	u := make([]uint16, len(s)/2)
	for i := range u {
		u[i] = uint16(s[2*i])<<8 | uint16(s[2*i+1])
	}
	return string(utf16.Decode(u))
}

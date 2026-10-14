package main

import "testing"

func TestShow(t *testing.T) {
	p := newPainter(identity)
	p.font = &font{widths: map[int]float64{'H': 700}, dflt: 500}
	p.size = em
	p.beginText()
	p.moveText(100, 200)
	p.show("H\x80")

	if got, want := p.missing, 1; got != want {
		t.Errorf("missing: got %d, want %d", got, want)
	}
	if len(p.paths) == 0 {
		t.Fatal("no strokes for H")
	}
	for _, sp := range p.paths {
		for _, q := range append([]point{sp.start}, sp.segs[0]...) {
			if q.x < 100 || q.x > 140 || q.y < 200 || q.y > 242 {
				t.Errorf("point %v outside the H cell", q)
			}
		}
	}
	// advance: 700/1000 em for H, 500 for the missing glyph
	if got, want := p.tm[4], 100+0.7*em+0.5*em; got != want {
		t.Errorf("advance: got %v, want %v", got, want)
	}
}

func TestRotate(t *testing.T) {
	box, m := rotate([4]float64{0, 0, 100, 200}, 90)
	if got, want := box, [4]float64{0, 0, 200, 100}; got != want {
		t.Errorf("box: got %v, want %v", got, want)
	}
	// top left of the unrotated page ends up top right
	if got, want := m.apply(0, 200), (point{200, 100}); got != want {
		t.Errorf("corner: got %v, want %v", got, want)
	}
}

func TestUTF16(t *testing.T) {
	if got, want := utf16be("\x00A\x30\xa2"), "Aア"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestParseCMap(t *testing.T) {
	m := parseCMap(`/CIDInit /ProcSet findresource begin
2 beginbfchar
<0003> <0020>
<0024> <0041>
endbfchar
2 beginbfrange
<0044> <0046> <0061>
<0050> <0051> [<30A2> <30A4>]
endbfrange
endcmap`)
	for code, want := range map[int]string{3: " ", 0x24: "A", 0x44: "a", 0x46: "c", 0x50: "ア", 0x51: "イ"} {
		if got := m[code]; got != want {
			t.Errorf("%#x: got %q, want %q", code, got, want)
		}
	}
}

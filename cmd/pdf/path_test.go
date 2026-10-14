package main

import "testing"

func TestPainter(t *testing.T) {
	p := newPainter(identity)
	p.op("q", nil)
	p.op("cm", []float64{2, 0, 0, 2, 10, 20})
	p.op("re", []float64{0, 0, 5, 5})
	p.op("S", nil)
	p.op("Q", nil)
	p.op("m", []float64{1, 1})
	p.op("c", []float64{2, 2, 3, 3, 4, 4})
	p.op("n", nil)
	p.op("m", []float64{0, 0})
	p.op("v", []float64{1, 0, 1, 1})
	p.op("f", nil)
	p.op("Do", nil)

	if got, want := len(p.paths), 2; got != want {
		t.Fatalf("got %d paths, want %d", got, want)
	}
	rect := p.paths[0]
	if got, want := rect.start, (point{10, 20}); got != want {
		t.Errorf("rect start: got %v, want %v", got, want)
	}
	if got, want := len(rect.segs), 4; got != want {
		t.Errorf("rect segments: got %d, want %d", got, want)
	}
	if got, want := rect.segs[1][0], (point{20, 30}); got != want {
		t.Errorf("rect corner: got %v, want %v", got, want)
	}
	if got, want := rect.segs[3][0], rect.start; got != want {
		t.Errorf("rect not closed: got %v, want %v", got, want)
	}
	curve := p.paths[1].segs[0]
	if got, want := curve[0], (point{0, 0}); got != want {
		t.Errorf("v first control: got %v, want %v", got, want)
	}
	if got, want := p.skipped["Do"], 1; got != want {
		t.Errorf("skipped Do: got %d, want %d", got, want)
	}
}

func TestMatrix(t *testing.T) {
	scale := matrix{2, 0, 0, 2, 0, 0}
	move := matrix{1, 0, 0, 1, 5, 7}
	if got, want := scale.mul(move).apply(1, 1), (point{7, 9}); got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

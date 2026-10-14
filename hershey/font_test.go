package hershey

import "testing"

// TestFontContinuity checks that no two strokes of a glyph could be plotted
// as one: open strokes must not meet end to end or touch a closed loop.
func TestFontContinuity(t *testing.T) {
	for r, g := range font {
		for i, a := range g.S {
			for _, b := range g.S[i+1:] {
				if joinable(a, b) || joinable(b, a) {
					t.Errorf("%q: strokes %v and %v can be joined", r, a[0], b[0])
				}
			}
		}
	}
}

func joinable(a, b Path) bool {
	ends := []Point{b[0], b[len(b)-1]}
	if a[0] == a[len(a)-1] { // closed loop: any vertex is a start point
		for _, p := range a {
			for _, e := range ends {
				if p == e {
					return true
				}
			}
		}
		return false
	}
	for _, e := range ends {
		if a[0] == e || a[len(a)-1] == e {
			return true
		}
	}
	return false
}

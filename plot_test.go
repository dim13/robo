package robo

import "testing"

func TestDraw(t *testing.T) {
	p := Point{1, 3}
	t.Log(Draw(p))
	t.Log(Draw(p, p))
	t.Log(Curve(10, p, p))
	t.Log(Move(p))
}

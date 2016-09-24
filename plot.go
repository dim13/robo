package robo

type Plotter interface {
	Plot() []byte
}

type Line []Point
type Bezier [4]Point
type Circle [3]Point

func (v Line) Draw() []byte   { return nil } // MDDDD...
func (v Bezier) Draw() []byte { return nil } // BZ...
func (v Circle) Draw() []byte { return nil } // W...

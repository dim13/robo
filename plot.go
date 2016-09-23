package robo

type Plotter interface {
	Plot() []byte
}

type Line []Point
type Berzier [4]Point

func (v Line) Draw() []byte   { return nil } // MDDDD...
func (v Bezier) Draw() []byte { return nil } // BZ...

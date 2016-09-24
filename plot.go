package robo

import (
	"fmt"
	"strings"
)

type Plotter interface {
	Plot() []byte
}

type Line []Point
type Bezier [4]Point
type Circle [3]Point

func (v Line) Plot() []byte   { return nil } // MDDDD...
func (v Bezier) Plot() []byte { return nil } // BZ...
func (v Circle) Plot() []byte { return nil } // W...

func join(p ...Point) string {
	path := make([]string, len(p))
	for i, v := range p {
		path[i] = v.String()
	}
	return strings.Join(path, ",")
}

// ???
func Draw(p ...Point) string          { return fmt.Sprintf("D%v", Path(p)) }
func DrawRelateive(p ...Point) string { return fmt.Sprintf("E%v", Path(p)) }
func Move(p Point) string             { return fmt.Sprintf("M%v", p) }
func MoveRelative(p Point) string     { return fmt.Sprintf("O%v", p) }

//func Offset(p Point) string          { return "^" + p.String() }
//func LowerLeft(p Point) string   { return "\\" + p.String() }
//func UpperRight(p Point) string  { return "Z" + p.String() }
func CuttingArea(p Point) string { return "FU" + p.String() }

//func Calibration(p Point) string     { return "TB72" + p.String() }

func Curve(a int, p ...Point) string { return fmt.Sprintf("Y%d,%v", a, Path(p)) }

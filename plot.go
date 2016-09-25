package robo

import (
	"fmt"
	"log"
	"strings"
)

type Robo struct {
	dev Device
}

func NewRobo() (Robo, error) {
	dev, err := Open()
	if err != nil {
		return Robo{}, err
	}
	return Robo{dev}, nil
}

func (r Robo) Close() error {
	return r.dev.Close()
}

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

func (r Robo) Version() string {
	r.dev.WriteString("FG")
	resp, _ := r.dev.ReadString()
	return resp
}

func (r Robo) Init() {
	r.dev.Command([]byte{4})
}

func (r Robo) Ready() bool {
	r.dev.Command([]byte{5})
	resp, _ := r.dev.ReadString()
	log.Printf("ready %q", resp)
	return resp == "0"
}

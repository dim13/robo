package robo

import (
	"fmt"
	"strings"
	"time"
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
	return strings.TrimSpace(resp)
}

func (r Robo) Wait4Ready() {
	t := time.NewTicker(time.Second)
	defer t.Stop()
	for range t.C {
		if r.Ready() {
			return
		}
	}
}

func (r Robo) Step(d Direction) { r.dev.Command([]byte{0, byte(d)}) }
func (r Robo) BootUpgrade()     { r.dev.Command([]byte{1}) }
func (r Robo) Init()            { r.dev.Command([]byte{4}) }

func (r Robo) Ready() bool {
	r.dev.Command([]byte{5})
	resp, _ := r.dev.ReadString()
	return resp == "0"
}

func (r Robo) Printf(f string, a ...interface{}) {
	s := fmt.Sprintf(f, a...)
	r.dev.WriteString(s)
}

func (r Robo) GoHome()                 { r.Printf("TT") }
func (r Robo) Home()                   { r.Printf("H") }
func (r Robo) Origin()                 { r.Printf("FJ") }
func (r Robo) Calibrate()              { r.Printf("TB70") }
func (r Robo) TestCut()                { r.Printf("FH") }
func (r Robo) TestLoop()               { r.Printf("FI") }
func (r Robo) Factor(x, y, z Unit)     { r.Printf("&%v", Triple{x, y, z}) }
func (r Robo) Offset(p Point)          { r.Printf("^%v", p) }
func (r Robo) LowerLeft(p Point)       { r.Printf("\\%v", p) }
func (r Robo) UpperRight(p Point)      { r.Printf("Z%v", p) }
func (r Robo) CuttingArea(p Point)     { r.Printf("FU%v", p) }
func (r Robo) Calibration(p Point)     { r.Printf("TB72,%v", p) }
func (r Robo) Move(p Point)            { r.Printf("M%v", p) }
func (r Robo) MoveRelative(p Point)    { r.Printf("O%v", p) }
func (r Robo) Draw(p ...Point)         { r.Printf("D%v", Path(p)) }
func (r Robo) DrawRelative(p Point)    { r.Printf("E%v", p) }
func (r Robo) Bezier(a, b, c, d Point) { r.Printf("BZ1,%v", Path{a, b, c, d}) }
func (r Robo) Line(p ...Point) {
	r.Move(p[0])
	r.Draw(p[1:]...)
}

package gc

import (
	"image"
	"log"

	"github.com/llgcode/draw2d"
)

type GraphicContext struct {
	*draw2d.StackGraphicContext
}

func NewGraphicContext() *GraphicContext {
	return &GraphicContext{draw2d.NewStackGraphicContext()}
}

func (gc *GraphicContext) CreateStringPath(s string, x, y float64) float64 {
	panic("not implemented")
}

func (gc *GraphicContext) FillStringAt(text string, x, y float64) (cursor float64) {
	panic("not implemented")
}

func (gc *GraphicContext) GetStringBounds(s string) (left, top, right, bottom float64) {
	panic("not implemented")
}

func (gc *GraphicContext) StrokeString(text string) (cursor float64) {
	return gc.StrokeStringAt(text, 0, 0)
}

func (gc *GraphicContext) StrokeStringAt(text string, x, y float64) (cursor float64) {
	width := gc.CreateStringPath(text, x, y)
	gc.Stroke()
	return width
}

func (gc *GraphicContext) SetDPI(dpi int) {
}

func (gc *GraphicContext) GetDPI() int {
	return -1
}

func (gc *GraphicContext) Clear() {
	log.Println("clear")
}

func (gc *GraphicContext) ClearRect(x1, y1, x2, y2 int) {
}

func (gc *GraphicContext) DrawImage(img image.Image) {
}

func (gc *GraphicContext) FillString(text string) (cursor float64) {
	log.Println("fillstring", text)
	return 0
}

func (gc *GraphicContext) Stroke(paths ...*draw2d.PathStorage) {
	log.Println("stroke", gc.Current)
	for _, p := range paths {
		log.Println("stroke", p)
	}
}

func (gc *GraphicContext) Fill(paths ...*draw2d.PathStorage) {
	log.Println("fill", gc.Current)
	for _, p := range paths {
		log.Println("fill", p)
	}
}

func (gc *GraphicContext) FillStroke(paths ...*draw2d.PathStorage) {
	log.Println("fillstroke", gc.Current)
	for _, p := range paths {
		log.Println("fillstroke", p)
	}
}

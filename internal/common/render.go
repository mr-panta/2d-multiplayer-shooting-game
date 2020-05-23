package common

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

type Render func(win *pixelgl.Window, pos pixel.Vec)

type RenderObject interface {
	GetZ() int
	GetShape() pixel.Rect
	Render(win *pixelgl.Window, pos pixel.Vec)
}

type renderObject struct {
	z      int
	shape  pixel.Rect
	render Render
}

func NewRenderObject(z int, shape pixel.Rect, render Render) RenderObject {
	return &renderObject{
		z:      z,
		shape:  shape,
		render: render,
	}
}

func (obj *renderObject) GetZ() int {
	return obj.z
}

func (obj *renderObject) GetShape() pixel.Rect {
	return obj.shape
}

func (obj *renderObject) Render(win *pixelgl.Window, pos pixel.Vec) {
	if obj.render != nil {
		obj.render(win, pos)
	}
}

package entity

import (
	"github.com/faiface/pixel"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/animation"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/common"
)

const (
	waterZ = -2
)

var (
	waterShape = pixel.R(0, 0, 64, 64)
)

type Water struct {
	world   common.World
	ready   bool
	objects []common.RenderObject
}

func NewWater(world common.World) *Water {
	return &Water{
		world: world,
	}
}

func (w *Water) GetRenderObjects() []common.RenderObject {
	if !w.ready {
		w.setup()
	}
	return w.objects
}

func (w *Water) setup() {
	w.setupSouth()
	w.setupWest()
	w.setupEast()
	w.setupCorners()
	w.ready = true
}

func (w *Water) setupSouth() {
	width, _ := w.world.GetSize()
	for i := 0; i < width; i++ {
		shape := pixel.R(
			float64(i)*waterShape.W(),
			-waterShape.H(),
			float64(i+1)*waterShape.W(),
			0,
		)
		w.objects = append(
			w.objects,
			common.NewRenderObject(waterZ, shape, w.getRenderSouth(i, shape)),
		)
	}
}

func (w *Water) getRenderSouth(i int, shape pixel.Rect) common.Render {
	return func(target pixel.Target, viewPos pixel.Vec) {
		pos := shape.Min
		anim := animation.NewWaterSouth()
		anim.Pos = pos.Sub(viewPos)
		anim.WaterSouthType = i
		anim.Draw(target)
	}
}

func (w *Water) setupWest() {
	_, height := w.world.GetSize()
	for i := 0; i < height-1; i++ {
		shape := pixel.R(
			-waterShape.W(),
			float64(i)*waterShape.H(),
			0,
			float64(i+1)*waterShape.H(),
		)
		w.objects = append(
			w.objects,
			common.NewRenderObject(waterZ, shape, w.getRenderSide(i, shape, false)),
		)
	}
}

func (w *Water) setupEast() {
	width, height := w.world.GetSize()
	for i := 0; i < height-1; i++ {
		shape := pixel.R(
			float64(width)*waterShape.W(),
			float64(i)*waterShape.H(),
			float64(width+1)*waterShape.W(),
			float64(i+1)*waterShape.H(),
		)
		w.objects = append(
			w.objects,
			common.NewRenderObject(waterZ, shape, w.getRenderSide(i, shape, true)),
		)
	}
}

func (w *Water) getRenderSide(i int, shape pixel.Rect, right bool) common.Render {
	return func(target pixel.Target, viewPos pixel.Vec) {
		pos := shape.Min
		anim := animation.NewWaterSide()
		anim.Pos = pos.Sub(viewPos)
		anim.WaterSideType = i
		anim.Right = right
		anim.Draw(target)
	}
}

func (w *Water) setupCorners() {
	width, height := w.world.GetSize()
	{
		shape := pixel.R(
			-waterShape.W(),
			-waterShape.H(),
			0,
			0,
		)
		w.objects = append(
			w.objects,
			common.NewRenderObject(waterZ, shape, w.getRenderCorner(shape, false, false)),
		)
	}
	{
		shape := pixel.R(
			float64(width)*waterShape.W(),
			-waterShape.H(),
			float64(width+1)*waterShape.W(),
			0,
		)
		w.objects = append(
			w.objects,
			common.NewRenderObject(waterZ, shape, w.getRenderCorner(shape, true, false)),
		)
	}
	{
		shape := pixel.R(
			-waterShape.W(),
			float64(height-1)*waterShape.H(),
			0,
			float64(height)*waterShape.H(),
		)
		w.objects = append(
			w.objects,
			common.NewRenderObject(waterZ, shape, w.getRenderCorner(shape, false, true)),
		)
	}
	{
		shape := pixel.R(
			float64(width)*waterShape.W(),
			float64(height-1)*waterShape.H(),
			float64(width+1)*waterShape.W(),
			float64(height)*waterShape.H(),
		)
		w.objects = append(
			w.objects,
			common.NewRenderObject(waterZ, shape, w.getRenderCorner(shape, true, true)),
		)
	}
}

func (w *Water) getRenderCorner(shape pixel.Rect, right bool, north bool) common.Render {
	return func(target pixel.Target, viewPos pixel.Vec) {
		pos := shape.Min
		anim := animation.NewWaterCorner()
		anim.Pos = pos.Sub(viewPos)
		anim.Right = right
		anim.North = north
		anim.Draw(target)
	}
}

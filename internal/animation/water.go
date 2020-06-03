package animation

import (
	"github.com/faiface/pixel"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/ticktime"
)

const (
	waterFrameMS = 300
)

var (
	waterFrameOffset = pixel.V(0, 768)
	waterSouthSet    = [][]pixel.Rect{
		{
			pixel.R(0*64, 0, 1*64, 64).Moved(waterFrameOffset),
			pixel.R(1*64, 0, 2*64, 64).Moved(waterFrameOffset),
			pixel.R(2*64, 0, 3*64, 64).Moved(waterFrameOffset),
		},
		{
			pixel.R(3*64, 0, 4*64, 64).Moved(waterFrameOffset),
			pixel.R(4*64, 0, 5*64, 64).Moved(waterFrameOffset),
			pixel.R(5*64, 0, 6*64, 64).Moved(waterFrameOffset),
		},
		{
			pixel.R(6*64, 0, 7*64, 64).Moved(waterFrameOffset),
			pixel.R(7*64, 0, 8*64, 64).Moved(waterFrameOffset),
			pixel.R(8*64, 0, 9*64, 64).Moved(waterFrameOffset),
		},
	}
	waterSideFrames = []pixel.Rect{
		pixel.R(6*64, 64, 7*64, 128).Moved(waterFrameOffset),
		pixel.R(7*64, 64, 8*64, 128).Moved(waterFrameOffset),
		pixel.R(8*64, 64, 9*64, 128).Moved(waterFrameOffset),
	}
	waterSouthCornerFrames = []pixel.Rect{
		pixel.R(0*64, 64, 1*64, 128).Moved(waterFrameOffset),
		pixel.R(1*64, 64, 2*64, 128).Moved(waterFrameOffset),
		pixel.R(2*64, 64, 3*64, 128).Moved(waterFrameOffset),
	}
	waterNorthCornerFrames = []pixel.Rect{
		pixel.R(3*64, 64, 4*64, 128).Moved(waterFrameOffset),
		pixel.R(4*64, 64, 5*64, 128).Moved(waterFrameOffset),
		pixel.R(5*64, 64, 6*64, 128).Moved(waterFrameOffset),
	}
)

type WaterSouth struct {
	Pos            pixel.Vec
	WaterSouthType int
}

func NewWaterSouth() *WaterSouth {
	return &WaterSouth{}
}

func (w *WaterSouth) Draw(target pixel.Target) {
	frames := waterSouthSet[w.WaterSouthType%len(waterSouthSet)]
	index := int(ticktime.GetServerTimeMS()/waterFrameMS) % len(frames)
	frame := frames[index]
	x := frame.W() / 2
	y := frame.H() / 2
	matrix := pixel.IM.Moved(w.Pos.Add(pixel.V(x, y)))
	sprite := pixel.NewSprite(objectSheet, frame)
	sprite.Draw(target, matrix)
}

type WaterSide struct {
	Pos           pixel.Vec
	WaterSideType int
	Right         bool
}

func NewWaterSide() *WaterSide {
	return &WaterSide{}
}

func (w *WaterSide) Draw(target pixel.Target) {
	index := w.WaterSideType
	index += int(ticktime.GetServerTimeMS() / waterFrameMS)
	index %= len(waterSideFrames)
	matrix := pixel.IM
	if w.Right {
		matrix = matrix.ScaledXY(pixel.ZV, pixel.V(-1, 1))
		index = len(waterSideFrames) - index - 1
	}
	frame := waterSideFrames[index]
	x := frame.W() / 2
	y := frame.H() / 2
	matrix = matrix.Moved(w.Pos.Add(pixel.V(x, y)))
	sprite := pixel.NewSprite(objectSheet, frame)
	sprite.Draw(target, matrix)
}

type WaterCorner struct {
	Pos   pixel.Vec
	North bool
	Right bool
}

func NewWaterCorner() *WaterCorner {
	return &WaterCorner{}
}

func (w *WaterCorner) Draw(target pixel.Target) {
	frames := waterSouthCornerFrames
	if w.North {
		frames = waterNorthCornerFrames
	}
	index := int(ticktime.GetServerTimeMS()/waterFrameMS) % len(frames)
	matrix := pixel.IM
	if w.Right {
		matrix = matrix.ScaledXY(pixel.ZV, pixel.V(-1, 1))
	}
	frame := frames[index]
	x := frame.W() / 2
	y := frame.H() / 2
	matrix = matrix.Moved(w.Pos.Add(pixel.V(x, y)))
	sprite := pixel.NewSprite(objectSheet, frame)
	sprite.Draw(target, matrix)
}

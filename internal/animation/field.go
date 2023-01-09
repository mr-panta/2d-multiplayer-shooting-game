package animation

import (
	"github.com/faiface/pixel"
)

var (
	fieldFrameOffset = pixel.V(0, 672)
	fieldFrames      = []pixel.Rect{
		pixel.R(0*96, 0, 1*96, 96).Moved(fieldFrameOffset),
		pixel.R(1*96, 0, 2*96, 96).Moved(fieldFrameOffset),
		pixel.R(2*96, 0, 3*96, 96).Moved(fieldFrameOffset),
		pixel.R(3*96, 0, 4*96, 96).Moved(fieldFrameOffset),
		pixel.R(4*96, 0, 5*96, 96).Moved(fieldFrameOffset),
	}
)

const (
	fieldSize = 64.0
)

type Field struct {
	Pos         pixel.Vec
	TerrainType int
}

func NewField() *Field {
	return &Field{}
}

func (f *Field) Draw(t pixel.Target) {
	frame := fieldFrames[f.TerrainType%len(fieldFrames)]
	sprite := pixel.NewSprite(objectSheet, frame)
	x := fieldSize / 2
	y := fieldSize / 2
	matrix := pixel.IM.Moved(f.Pos.Add(pixel.V(x, y)))
	sprite.Draw(t, matrix)
}

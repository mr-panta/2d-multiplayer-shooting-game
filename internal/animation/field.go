package animation

import (
	"image/color"

	"github.com/faiface/pixel"
)

var (
	fieldFrame = pixel.R(0, 0, 60, 60)
	fieldColor = color.RGBA{0xb0, 0xff, 0x8d, 0xff}
)

type Field struct {
	Pos pixel.Vec
}

func NewField() *Field {
	return &Field{}
}

func (f *Field) Draw(t pixel.Target) {
	sprite := pixel.NewSprite(FieldSheet, fieldFrame)
	matrix := pixel.IM.Scaled(pixel.ZV, 4).Moved(f.Pos)
	sprite.DrawColorMask(t, matrix, fieldColor)
}

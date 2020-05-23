package animation

import (
	"image/color"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

var itemAmmoFrame = pixel.R(0, 0, 45, 47)

type ItemAmmo struct {
	Pos   pixel.Vec
	Color color.Color
}

func NewItemAmmo() *ItemAmmo {
	return &ItemAmmo{}
}

func (i *ItemAmmo) Draw(win *pixelgl.Window) {
	sprite := pixel.NewSprite(itemSheet, itemAmmoFrame)
	matrix := pixel.IM.Moved(i.Pos)
	sprite.DrawColorMask(win, matrix, i.Color)
}

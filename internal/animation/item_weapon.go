package animation

import (
	"image/color"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

var itemWeaponFrame = pixel.R(50, 0, 50+45, 47)

type ItemWeapon struct {
	Pos   pixel.Vec
	Color color.Color
}

func NewItemWeapon() *ItemWeapon {
	return &ItemWeapon{}
}

func (w *ItemWeapon) Draw(win *pixelgl.Window) {
	sprite := pixel.NewSprite(itemSheet, itemWeaponFrame)
	matrix := pixel.IM.Moved(w.Pos)
	sprite.DrawColorMask(win, matrix, w.Color)
}

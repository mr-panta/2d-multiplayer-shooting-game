package animation

import (
	"image/color"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

var itemAmmoSMFrame = pixel.R(150, 0, 150+20, 26)

type ItemAmmoSM struct {
	Pos   pixel.Vec
	Color color.Color
}

func NewItemAmmoSM() *ItemAmmoSM {
	return &ItemAmmoSM{}
}

func (i *ItemAmmoSM) Draw(win *pixelgl.Window) {
	sprite := pixel.NewSprite(itemSheet, itemAmmoSMFrame)
	matrix := pixel.IM.Moved(i.Pos)
	sprite.DrawColorMask(win, matrix, i.Color)
}

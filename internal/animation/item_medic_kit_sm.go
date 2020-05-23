package animation

import (
	"image/color"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

var itemMedicKitSMFrame = pixel.R(175, 0, 175+20, 26)

type ItemMedicKitSM struct {
	Pos   pixel.Vec
	Color color.Color
}

func NewItemMedicKitSM() *ItemMedicKitSM {
	return &ItemMedicKitSM{}
}

func (i *ItemMedicKitSM) Draw(win *pixelgl.Window) {
	sprite := pixel.NewSprite(itemSheet, itemMedicKitSMFrame)
	matrix := pixel.IM.Moved(i.Pos)
	sprite.DrawColorMask(win, matrix, i.Color)
}

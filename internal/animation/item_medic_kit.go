package animation

import (
	"image/color"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

var itemMedicKitFrame = pixel.R(100, 0, 100+49, 35)

type ItemMedicKit struct {
	Pos   pixel.Vec
	Color color.Color
}

func NewItemMedicKit() *ItemMedicKit {
	return &ItemMedicKit{}
}

func (i *ItemMedicKit) Draw(win *pixelgl.Window) {
	sprite := pixel.NewSprite(itemSheet, itemMedicKitFrame)
	matrix := pixel.IM.Moved(i.Pos)
	sprite.DrawColorMask(win, matrix, i.Color)
}

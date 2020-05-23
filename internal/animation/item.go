package animation

import (
	"image/color"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

var (
	itemAmmoFrame       = pixel.R(0, 0, 45, 47)
	itemWeaponFrame     = pixel.R(50, 0, 50+45, 47)
	itemMedicKitFrame   = pixel.R(100, 0, 100+49, 35)
	itemAmmoSMFrame     = pixel.R(150, 0, 150+20, 26)
	itemMedicKitSMFrame = pixel.R(175, 0, 175+20, 26)
)

func NewItemAmmo() *Item {
	return &Item{frame: itemAmmoFrame}
}

func NewItemWeapon() *Item {
	return &Item{frame: itemWeaponFrame}
}

func NewItemMedicKit() *Item {
	return &Item{frame: itemMedicKitFrame}
}

func NewItemAmmoSM() *Item {
	return &Item{frame: itemAmmoSMFrame}
}

func NewItemMedicKitSM() *Item {
	return &Item{frame: itemMedicKitSMFrame}
}

type Item struct {
	frame pixel.Rect
	Pos   pixel.Vec
	Color color.Color
}

func (i *Item) Draw(win *pixelgl.Window) {
	sprite := pixel.NewSprite(itemSheet, i.frame)
	matrix := pixel.IM.Moved(i.Pos)
	sprite.DrawColorMask(win, matrix, i.Color)
}

package animation

import (
	"image/color"

	"github.com/faiface/pixel"
)

var (
	itemFrameOffset       = pixel.V(0, 3*64)
	itemAmmoFrame         = pixel.R(0*64, 0, 1*64, 64).Moved(itemFrameOffset)
	itemWeaponFrame       = pixel.R(1*64, 0, 2*64, 64).Moved(itemFrameOffset)
	itemMedicKitFrame     = pixel.R(2*64, 0, 3*64, 64).Moved(itemFrameOffset)
	itemAmmoSMFrame       = pixel.R(3*64, 0, 3*64+32, 32).Moved(itemFrameOffset)
	itemMedicKitSMFrame   = pixel.R(3*64, 32, 3*64+32, 64).Moved(itemFrameOffset)
	itemWeaponSniperFrame = pixel.R(7*32, 0, 9*32, 64).Moved(itemFrameOffset)
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

func NewItemWeaponSniper() *Item {
	return &Item{frame: itemWeaponSniperFrame}
}

type Item struct {
	frame pixel.Rect
	Pos   pixel.Vec
	Color color.Color
}

func (i *Item) Draw(target pixel.Target) {
	sprite := pixel.NewSprite(objectSheet, i.frame)
	matrix := pixel.IM.Moved(i.Pos.Add(pixel.V(0, i.frame.H()/2)))
	sprite.DrawColorMask(target, matrix, i.Color)
}

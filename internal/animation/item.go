package animation

import (
	"image/color"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
)

var (
	itemFrameOffset       = pixel.V(0, 320)
	itemAmmoFrame         = pixel.R(0*64, 0, 1*64, 64).Moved(itemFrameOffset)
	itemWeaponFrame       = pixel.R(1*64, 0, 2*64, 64).Moved(itemFrameOffset)
	itemMedicKitFrame     = pixel.R(2*64, 0, 3*64, 64).Moved(itemFrameOffset)
	itemAmmoSMFrame       = pixel.R(3*64, 0, 3*64+32, 32).Moved(itemFrameOffset)
	itemMedicKitSMFrame   = pixel.R(3*64, 32, 3*64+32, 64).Moved(itemFrameOffset)
	itemWeaponSniperFrame = pixel.R(7*32, 0, 9*32, 64).Moved(itemFrameOffset)
	itemArmorFrame        = pixel.R(9*32, 0, 11*32, 64).Moved(itemFrameOffset)
	itemArmorBlueFrame    = pixel.R(11*32, 0, 13*32, 64).Moved(itemFrameOffset)
	itemSkullFrame        = pixel.R(13*32, 1, 15*32, 63).Moved(itemFrameOffset)
	itemMysteryFrame      = pixel.R(15*32, 1, 17*32, 63).Moved(itemFrameOffset)
	itemLandMineFrame     = pixel.R(17*32, 1, 19*32, 63).Moved(itemFrameOffset)
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

func NewItemArmor() *Item {
	return &Item{frame: itemArmorFrame}
}

func NewItemArmorBlue() *Item {
	return &Item{frame: itemArmorBlueFrame}
}

func NewItemSkull() *Item {
	return &Item{
		frame:     itemSkullFrame,
		shadowImd: imdraw.New(nil),
	}
}

func NewItemMystery() *Item {
	return &Item{frame: itemMysteryFrame}
}

func NewItemLandMine() *Item {
	return &Item{frame: itemLandMineFrame}
}

type Item struct {
	frame     pixel.Rect
	shadowImd *imdraw.IMDraw
	Pos       pixel.Vec
	Color     color.Color
}

func (i *Item) Draw(target pixel.Target) {
	sprite := pixel.NewSprite(objectSheet, i.frame)
	matrix := pixel.IM.Moved(i.Pos.Add(pixel.V(0, i.frame.H()/2)))
	if i.shadowImd != nil {
		i.shadowImd.Clear()
		i.shadowImd.Color = shadowColor
		i.shadowImd.Push(pixel.V(0, -i.frame.H()/2))
		i.shadowImd.SetMatrix(matrix)
		i.shadowImd.Ellipse(pixel.V(i.frame.W()/4, i.frame.H()/12), 0)
		i.shadowImd.Draw(target)
	}
	sprite.DrawColorMask(target, matrix, i.Color)
}

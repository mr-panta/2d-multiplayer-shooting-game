package animation

import (
	"image/color"

	"github.com/faiface/pixel"
)

var (
	iconFrameOffset    = pixel.V(0, 896)
	iconSkullFrame     = pixel.R(0, 1, 64, 63).Moved(iconFrameOffset)
	iconHeartFrame     = pixel.R(64, 1, 2*64, 63).Moved(iconFrameOffset)
	iconShieldFrame    = pixel.R(2*64, 1, 3*64, 63).Moved(iconFrameOffset)
	iconInventoryFrame = pixel.R(3*64, 1, 5*64, 63).Moved(iconFrameOffset)
	iconLandMineFrame  = pixel.R(5*64, 1, 6*64, 63).Moved(iconFrameOffset)
)

type Icon struct {
	frame pixel.Rect
	Color color.Color
	Pos   pixel.Vec
	Size  float64
}

func NewIconSkull() *Icon {
	return &Icon{
		frame: iconSkullFrame,
	}
}

func NewIconHeart() *Icon {
	return &Icon{
		frame: iconHeartFrame,
	}
}

func NewIconShield() *Icon {
	return &Icon{
		frame: iconShieldFrame,
	}
}

func NewIconInventory() *Icon {
	return &Icon{
		frame: iconInventoryFrame,
	}
}

func NewIconLandMine() *Icon {
	return &Icon{
		frame: iconLandMineFrame,
	}
}

func (i *Icon) Draw(target pixel.Target) {
	sprite := pixel.NewSprite(objectSheet, i.frame)
	matrix := pixel.IM
	if i.Size > 0 {
		matrix = matrix.Scaled(pixel.ZV, i.Size)
	}
	matrix = matrix.Moved(i.Pos)
	sprite.DrawColorMask(target, matrix, i.Color)
}

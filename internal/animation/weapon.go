package animation

import (
	"image/color"
	"math"

	_ "image/png"

	"github.com/faiface/pixel"
)

var (
	weaponM4Frame  = pixel.R(0, 2*64+1, 3*32, 3*64-1)
	weaponM4Offset = pixel.V(-16, 0)
)

const (
	// state
	WeaponM4IdleState   = 0
	WeaponM4ReloadState = 1
)

type WeaponM4 struct {
	Pos   pixel.Vec
	Dir   pixel.Vec
	Color color.Color
	State int
}

func NewWeaponM4() *WeaponM4 {
	return &WeaponM4{}
}

func (m *WeaponM4) Draw(target pixel.Target) {
	sprite := pixel.NewSprite(objectSheet, weaponM4Frame)
	var dir pixel.Vec
	switch m.State {
	case WeaponM4IdleState:
		dir = m.Dir
	case WeaponM4ReloadState:
		dir = pixel.V(m.Dir.X, -math.Abs(m.Dir.X))
	}
	matrix := pixel.IM.Moved(weaponM4Offset)
	if m.Dir.X > 0 {
		matrix = matrix.ScaledXY(pixel.ZV, pixel.V(-1, 1))
		matrix = matrix.Rotated(pixel.ZV, dir.Angle())
	} else {
		matrix = matrix.Rotated(pixel.ZV, pixel.ZV.Sub(dir).Angle())
	}
	matrix = matrix.Moved(m.Pos)
	sprite.DrawColorMask(target, matrix, m.Color)
}

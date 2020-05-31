package animation

import (
	"image/color"
	"math"

	_ "image/png"

	"github.com/faiface/pixel"
)

var (
	// m4
	weaponM4Frame  = pixel.R(0, 2*64+1, 3*32, 3*64-1)
	weaponM4Offset = pixel.V(-16, 0)
	// shotgun
	weaponShotgunFrame  = pixel.R(8*32, 2*64+1, 11*32, 3*64-1)
	weaponShotgunOffset = pixel.V(-16, 0)
	// sniper
	weaponSniperFrame  = pixel.R(3*32, 2*64+1, 8*32, 3*64-1)
	weaponSniperOffset = pixel.V(-24, 0)
	// pistol
	weaponPistolFrame  = pixel.R(11*32, 2*64+1, 12*32, 3*64-1)
	weaponPistolOffset = pixel.V(-24, 0)
	// pistol
	weaponSMGFrame  = pixel.R(12*32, 2*64+1, 14*32, 3*64-1)
	weaponSMGOffset = pixel.V(-12, 0)
)

const (
	// state
	WeaponIdleState   = 0
	WeaponReloadState = 1
)

func NewWeaponM4() *Weapon {
	return &Weapon{
		frame:  weaponM4Frame,
		offset: weaponM4Offset,
	}
}

func NewWeaponShotgun() *Weapon {
	return &Weapon{
		frame:  weaponShotgunFrame,
		offset: weaponShotgunOffset,
	}
}

func NewWeaponSniper() *Weapon {
	return &Weapon{
		frame:  weaponSniperFrame,
		offset: weaponSniperOffset,
	}
}

func NewWeaponPistol() *Weapon {
	return &Weapon{
		frame:  weaponPistolFrame,
		offset: weaponPistolOffset,
	}
}

func NewWeaponSMG() *Weapon {
	return &Weapon{
		frame:  weaponSMGFrame,
		offset: weaponSMGOffset,
	}
}

type Weapon struct {
	frame  pixel.Rect
	offset pixel.Vec
	Pos    pixel.Vec
	Dir    pixel.Vec
	Color  color.Color
	State  int
}

func (m *Weapon) Draw(target pixel.Target) {
	sprite := pixel.NewSprite(objectSheet, m.frame)
	var dir pixel.Vec
	switch m.State {
	case WeaponIdleState:
		dir = m.Dir
	case WeaponReloadState:
		dir = pixel.V(m.Dir.X, -math.Abs(m.Dir.X))
	}
	matrix := pixel.IM.Moved(m.offset)
	if m.Dir.X > 0 {
		matrix = matrix.ScaledXY(pixel.ZV, pixel.V(-1, 1))
		matrix = matrix.Rotated(pixel.ZV, dir.Angle())
	} else {
		matrix = matrix.Rotated(pixel.ZV, pixel.ZV.Sub(dir).Angle())
	}
	matrix = matrix.Moved(m.Pos)
	sprite.DrawColorMask(target, matrix, m.Color)
}

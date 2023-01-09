package animation

import (
	"image/color"
	"math"
	"time"

	_ "image/png"

	"github.com/faiface/pixel"
)

var (
	// offset
	weaponOffset = pixel.V(0, 256)
	// m4
	weaponM4Frame  = pixel.R(0, 1, 3*32, 63).Moved(weaponOffset)
	weaponM4Offset = pixel.V(-16, 0)
	// shotgun
	weaponShotgunFrame  = pixel.R(8*32, 1, 11*32, 63).Moved(weaponOffset)
	weaponShotgunOffset = pixel.V(-16, 0)
	// sniper
	weaponSniperFrame  = pixel.R(3*32, 1, 8*32, 63).Moved(weaponOffset)
	weaponSniperOffset = pixel.V(-24, 0)
	// pistol
	weaponPistolFrame  = pixel.R(11*32, 1, 13*32, 63).Moved(weaponOffset)
	weaponPistolOffset = pixel.V(-20, 0)
	// smg
	weaponSMGFrame  = pixel.R(13*32, 1, 15*32, 63).Moved(weaponOffset)
	weaponSMGOffset = pixel.V(-12, 0)
	// knife
	weaponKnifeFrame  = pixel.R(15*32, 1, 17*32, 63).Moved(weaponOffset)
	weaponKnifeOffset = pixel.V(-19, 0)
)

const (
	// Recoil
	recoilAngle   = 6
	recoilPosDiff = 6
	// state
	WeaponIdleState    = 0
	WeaponReloadState  = 1
	WeaponTriggerState = 2
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
	frame           pixel.Rect
	offset          pixel.Vec
	Pos             pixel.Vec
	Dir             pixel.Vec
	Color           color.Color
	TriggerDuration time.Duration
	TriggerCooldown time.Duration
	State           int
}

func (m *Weapon) Draw(target pixel.Target) {
	sprite := pixel.NewSprite(objectSheet, m.frame)
	var dir pixel.Vec
	var posDiff pixel.Vec
	var angleDiff float64
	switch m.State {
	case WeaponIdleState:
		dir = m.Dir
	case WeaponReloadState:
		dir = pixel.V(m.Dir.X, -math.Abs(m.Dir.X))
	case WeaponTriggerState:
		fac := pixel.Clamp(1.0-(float64(m.TriggerDuration)/float64(m.TriggerCooldown)), 0, 1)
		posDiff = pixel.V(recoilPosDiff*fac, 0)
		angleDiff = -math.Pi / 180 * (recoilAngle * fac)
		dir = m.Dir
	}
	matrix := pixel.IM.Moved(m.offset.Add(posDiff)).Rotated(pixel.ZV, angleDiff)
	if m.Dir.X > 0 {
		matrix = matrix.ScaledXY(pixel.ZV, pixel.V(-1, 1))
		matrix = matrix.Rotated(pixel.ZV, dir.Angle())
	} else {
		matrix = matrix.Rotated(pixel.ZV, pixel.ZV.Sub(dir).Angle())
	}
	matrix = matrix.Moved(m.Pos)
	sprite.DrawColorMask(target, matrix, m.Color)
}

func NewWeaponKnife() *WeaponKnife {
	return &WeaponKnife{}
}

type WeaponKnife struct {
	Pos    pixel.Vec
	Dir    pixel.Vec
	Radius float64
	Color  color.Color
}

func (m *WeaponKnife) Draw(target pixel.Target) {
	sprite := pixel.NewSprite(objectSheet, weaponKnifeFrame)
	matrix := pixel.IM.Moved(pixel.V(-m.Radius, 0).Add(weaponKnifeOffset))
	if m.Dir.X > 0 {
		matrix = matrix.ScaledXY(pixel.ZV, pixel.V(-1, 1))
		matrix = matrix.Rotated(pixel.ZV, m.Dir.Angle())
	} else {
		matrix = matrix.Rotated(pixel.ZV, pixel.ZV.Sub(m.Dir).Angle())
	}
	matrix = matrix.Moved(m.Pos)
	sprite.DrawColorMask(target, matrix, m.Color)
}

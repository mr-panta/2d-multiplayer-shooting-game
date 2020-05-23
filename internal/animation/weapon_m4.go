package animation

import (
	"image/color"
	"math"

	_ "image/png"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

const (
	// info
	weaponM4Width  = 114
	weaponM4Height = 32
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

func (m *WeaponM4) Draw(win *pixelgl.Window) {
	sprite := pixel.NewSprite(
		weaponSheet,
		pixel.R(0, 0, float64(weaponM4Width), float64(weaponM4Height)),
	)
	var dir pixel.Vec
	switch m.State {
	case WeaponM4IdleState:
		dir = m.Dir
	case WeaponM4ReloadState:
		dir = pixel.V(m.Dir.X, -math.Abs(m.Dir.X))
	}
	matrix := pixel.IM
	if m.Dir.X > 0 {
		matrix = matrix.ScaledXY(pixel.ZV, pixel.V(-1, 1))
		matrix = matrix.Rotated(pixel.ZV, dir.Angle())
	} else {
		matrix = matrix.Rotated(pixel.ZV, pixel.ZV.Sub(dir).Angle())
	}
	matrix = matrix.Moved(m.Pos)
	sprite.DrawColorMask(win, matrix, m.Color)
}

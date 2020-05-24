package entity

import (
	"image/color"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/common"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/config"
)

const (
	scopeZ = 5
)

var scopeColor = color.RGBA{0x00, 0x00, 0x00, 0xa0}

type Scope struct {
	world   common.World
	imd     *imdraw.IMDraw
	pos     pixel.Vec
	radius  float64
	visible bool
}

func NewScope(world common.World) common.Scope {
	return &Scope{
		world:   world,
		imd:     imdraw.New(nil),
		visible: false,
	}
}

func (s *Scope) Update() {
	if p := s.getPlayer(); p != nil && p.Exists() {
		pos := s.world.GetWindow().MousePosition()
		dist := p.GetPivot().Sub(s.world.GetCameraViewPos()).Sub(pos).Len()
		s.radius = p.GetScopeRadius(dist)
		s.pos = pos
		s.visible = true
	} else {
		s.visible = false
	}
}

func (s *Scope) GetRenderObject() common.RenderObject {
	if s.visible {
		return common.NewRenderObject(scopeZ, pixel.ZR, s.render)
	}
	return nil
}

func (s *Scope) render(target pixel.Target, viewPos pixel.Vec) {
	win := s.world.GetWindow()
	s.imd.Clear()
	s.imd.Color = scopeColor
	if win.MouseInsideWindow() {
		s.imd.Push(s.pos)
		s.imd.Circle(s.radius+win.Bounds().W(), 2*win.Bounds().W())
	} else {
		s.imd.Push(win.Bounds().Min, win.Bounds().Max)
		s.imd.Rectangle(0)
	}
	s.imd.Draw(target)
}

func (s *Scope) Intersects(shape pixel.Rect) bool {
	if !s.visible || config.EnvDebug() {
		return true
	}
	if s.radius == 0 {
		return false
	}
	shape = shape.Moved(pixel.ZV.Sub(s.world.GetCameraViewPos()))
	circle := pixel.C(s.pos, s.radius)
	v := circle.IntersectRect(shape)
	return !v.Eq(pixel.ZV)
}

func (s *Scope) getPlayer() common.Player {
	if playerID := s.world.GetMainPlayerID(); playerID != "" {
		if o, exists := s.world.GetObjectDB().SelectOne(playerID); exists {
			return o.(common.Player)
		}
	}
	return nil
}

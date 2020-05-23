package entity

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/common"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/config"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/protocol"
)

type Boundary struct {
	world    common.World
	id       string
	collider pixel.Rect
}

func NewBoundary(world common.World, id string, collider pixel.Rect) *Boundary {
	return &Boundary{
		world:    world,
		id:       id,
		collider: collider,
	}
}

func (o *Boundary) GetID() string {
	return o.id
}
func (o *Boundary) GetType() int {
	return 0
}
func (o *Boundary) Destroy() {
	// NOOP
}
func (o *Boundary) Exists() bool {
	return true
}
func (o *Boundary) GetShape() pixel.Rect {
	return o.collider
}
func (o *Boundary) GetCollider() (pixel.Rect, bool) {
	return o.collider, true

}
func (o *Boundary) GetRenderObjects() []common.RenderObject {
	return []common.RenderObject{
		common.NewRenderObject(1, o.GetShape(), o.render),
	}
}
func (o *Boundary) GetSnapshot(tick int64) *protocol.ObjectSnapshot {
	return &protocol.ObjectSnapshot{
		ID:   o.GetID(),
		Type: o.GetType(),
	}
}
func (o *Boundary) SetSnapshot(tick int64, snapshot *protocol.ObjectSnapshot) {
	// NOOP

}
func (o *Boundary) ServerUpdate(tick int64) {
	// NOOP

}
func (o *Boundary) ClientUpdate() {
	// NOOP
}

func (o *Boundary) render(win *pixelgl.Window, viewPos pixel.Vec) {
	imd := imdraw.New(nil)
	imd.Color = config.ColliderColor
	imd.Push(o.GetShape().Min, o.GetShape().Max)
	imd.SetMatrix(pixel.IM.Moved(pixel.ZV.Sub(viewPos)))
	imd.Rectangle(1)
	imd.Draw(win)
}

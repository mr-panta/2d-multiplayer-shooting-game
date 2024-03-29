package entity

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/animation"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/common"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/config"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/protocol"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/util"
)

const (
	treeZ = 10
)

var (
	// shape
	treeAShape = pixel.R(0, 0, 100, 250)
	treeBShape = pixel.R(0, 0, 183, 260)
	treeCShape = pixel.R(0, 0, 240, 280)
	treeDShape = pixel.R(0, 0, 132, 90)
	treeEShape = pixel.R(0, 0, 138, 90)
	// collider
	treeACollider = pixel.R(0, 0, 40, 40)
	treeBCollider = pixel.R(0, 0, 40, 40)
	treeCCollider = pixel.R(0, 12, 30, 52)
)

type Tree struct {
	world       common.World
	shapeImd    *imdraw.IMDraw
	colliderImd *imdraw.IMDraw
	id          string
	pos         pixel.Vec
	treeType    string
	right       bool
	transparent bool
}

func NewTree(world common.World, id string) *Tree {
	return &Tree{
		world:       world,
		id:          id,
		shapeImd:    imdraw.New(nil),
		colliderImd: imdraw.New(nil),
	}
}

func (o *Tree) GetID() string {
	return o.id
}

func (o *Tree) GetType() int {
	return config.TreeObject
}

func (o *Tree) Destroy() {
	// NOOP
}

func (o *Tree) Exists() bool {
	return true
}

func (o *Tree) GetShape() pixel.Rect {
	var shape pixel.Rect
	switch o.treeType {
	case config.TreeTypeA:
		shape = treeAShape
	case config.TreeTypeB:
		shape = treeBShape
	case config.TreeTypeC:
		shape = treeCShape
	case config.TreeTypeD:
		shape = treeDShape
	case config.TreeTypeE:
		shape = treeEShape
	}
	return shape.Moved(o.pos.Sub(pixel.V(shape.W()/2, 0)))
}

func (o *Tree) GetCollider() (pixel.Rect, bool) {
	if o.treeType == config.TreeTypeD || o.treeType == config.TreeTypeE {
		return pixel.ZR, false
	}
	var collider pixel.Rect
	switch o.treeType {
	case config.TreeTypeA:
		collider = treeACollider
	case config.TreeTypeB:
		collider = treeBCollider
	case config.TreeTypeC:
		collider = treeCCollider
	default:
		return pixel.ZR, false
	}
	return collider.Moved(o.pos.Sub(pixel.V(collider.W()/2, 0))), true
}

func (o *Tree) GetRenderObjects() []common.RenderObject {
	return []common.RenderObject{
		common.NewRenderObject(treeZ, o.GetShape(), o.render),
	}
}

func (o *Tree) GetSnapshot(tick int64) *protocol.ObjectSnapshot {
	return &protocol.ObjectSnapshot{
		ID:   o.GetID(),
		Type: o.GetType(),
		Tree: &protocol.TreeSnapshot{
			Pos:      util.ConvertVec(o.pos),
			TreeType: o.treeType,
			Right:    o.right,
		},
	}
}

func (o *Tree) SetSnapshot(tick int64, snapshot *protocol.ObjectSnapshot) {
	ss := snapshot.Tree
	o.SetState(ss.Pos.Convert(), ss.TreeType, ss.Right)
}

func (o *Tree) ServerUpdate(tick int64) {
	// NOOP
}

func (o *Tree) ClientUpdate() {
	// transparent := o.world.GetScope().Intersects(o.GetShape()) // Show when on scope
	transparent := false
	if !transparent {
		for _, obj := range o.world.GetObjectDB().SelectAll() {
			if obj.Exists() &&
				obj.GetType() == config.PlayerObject &&
				obj.GetShape().Min.Y > o.GetShape().Min.Y &&
				obj.GetShape().Intersects(o.GetShape()) {
				player := obj.(common.Player)
				if player.IsVisible() || player.GetID() == o.world.GetMainPlayerID() {
					transparent = true
					break
				}
			}
		}
	}
	o.transparent = transparent
}

func (o *Tree) SetState(pos pixel.Vec, treeType string, right bool) {
	o.pos = pos
	o.treeType = treeType
	o.right = right
}

func (o *Tree) render(target pixel.Target, viewPos pixel.Vec) {
	var anim *animation.Tree
	switch o.treeType {
	case config.TreeTypeA:
		anim = animation.NewTreeA()
	case config.TreeTypeB:
		anim = animation.NewTreeB()
	case config.TreeTypeC:
		anim = animation.NewTreeC()
	case config.TreeTypeD:
		anim = animation.NewTreeD()
	case config.TreeTypeE:
		anim = animation.NewTreeE()
	default:
		return
	}
	anim.Pos = o.pos.Sub(viewPos)
	anim.Right = o.right
	anim.Transparent = o.transparent
	anim.Draw(target)
	// debug
	if config.EnvDebug() {
		o.renderShape(target, viewPos)
		o.renderCollider(target, viewPos)
	}
}

func (o *Tree) renderCollider(target pixel.Target, viewPos pixel.Vec) {
	if collider, exists := o.GetCollider(); exists {
		o.colliderImd.Clear()
		o.colliderImd.Color = config.ColliderColor
		o.colliderImd.Push(collider.Min, collider.Max)
		o.colliderImd.Rectangle(1)
		o.colliderImd.SetMatrix(pixel.IM.Moved(pixel.ZV.Sub(viewPos)))
		o.colliderImd.Draw(target)
	}
}

func (o *Tree) renderShape(target pixel.Target, viewPos pixel.Vec) {
	o.shapeImd.Clear()
	o.shapeImd.Color = config.ShapeColor
	o.shapeImd.Push(o.GetShape().Min, o.GetShape().Max)
	o.shapeImd.Rectangle(1)
	o.shapeImd.SetMatrix(pixel.IM.Moved(pixel.ZV.Sub(viewPos)))
	o.shapeImd.Draw(target)
}

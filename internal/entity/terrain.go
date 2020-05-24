package entity

import (
	"github.com/faiface/pixel"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/common"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/config"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/protocol"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/util"
)

const (
	terrainZ = -1
)

type Terrain struct {
	world       common.World
	id          string
	pos         pixel.Vec
	terrainType int
	ready       bool
	fields      []common.Field
}

func NewTerrain(world common.World, id string) *Terrain {
	return &Terrain{
		world: world,
		id:    id,
	}
}

func (o *Terrain) GetID() string {
	return o.id
}

func (o *Terrain) GetType() int {
	return config.TerrainObject
}

func (o *Terrain) Destroy() {
	// NOOP
}

func (o *Terrain) Exists() bool {
	return true
}

func (o *Terrain) GetShape() pixel.Rect {
	return pixel.Rect{Min: o.pos, Max: o.pos}
}

func (o *Terrain) GetCollider() (pixel.Rect, bool) {
	return pixel.ZR, false
}

func (o *Terrain) GetRenderObjects() (objs []common.RenderObject) {
	for _, f := range o.fields {
		objs = append(objs, common.NewRenderObject(terrainZ, f.GetShape(), f.Render))
	}
	return objs
}

func (o *Terrain) GetSnapshot(tick int64) *protocol.ObjectSnapshot {
	return &protocol.ObjectSnapshot{
		ID:   o.GetID(),
		Type: o.GetType(),
		Terrain: &protocol.TerrainSnapshot{
			Pos:         util.ConvertVec(o.pos),
			TerrainType: o.terrainType,
		},
	}
}

func (o *Terrain) SetSnapshot(tick int64, snapshot *protocol.ObjectSnapshot) {
	ss := snapshot.Terrain
	o.SetState(ss.Pos.Convert(), ss.TerrainType)
}

func (o *Terrain) SetState(pos pixel.Vec, terrainType int) {
	o.pos = pos
	o.terrainType = terrainType
}

func (o *Terrain) GetTerrainType() int {
	return o.terrainType
}

func (o *Terrain) ServerUpdate(tick int64) {
	if !o.ready {
		o.setupFields()
	}
}

func (o *Terrain) ClientUpdate() {
	if !o.ready {
		o.setupFields()
	}
}

func (o *Terrain) setupFields() {
	otherTerrains := []common.Terrain{}
	for _, obj := range o.world.GetObjectDB().SelectAll() {
		if obj.GetType() == config.TerrainObject && obj.GetID() != o.id {
			otherTerrains = append(otherTerrains, obj.(common.Terrain))
		}
	}
	w, h := o.world.GetSize()
	for i := 0; i < h; i++ {
		for j := 0; j < w; j++ {
			pos := pixel.V(
				float64(j)*fieldShape.W(),
				float64(i)*fieldShape.H(),
			)
			ok := true
			diff := pos.Sub(o.pos).Len()
			for _, terrain := range otherTerrains {
				if diff > pos.Sub(terrain.GetShape().Center()).Len() {
					ok = false
					break
				}
			}
			if ok {
				o.fields = append(o.fields, NewField(pos, o.terrainType))
			}
		}
	}
	o.ready = true
}

package entity

import (
	"github.com/faiface/pixel"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/animation"
)

var fieldShape = pixel.R(0, 0, 64, 64)

type Field struct {
	pos         pixel.Vec
	terrainType int
}

func NewField(pos pixel.Vec, terrainType int) *Field {
	return &Field{
		pos:         pos,
		terrainType: terrainType,
	}
}

func (o *Field) GetShape() pixel.Rect {
	return fieldShape.Moved(o.pos)
}

func (o *Field) Render(t pixel.Target, viewPos pixel.Vec) {
	anim := animation.NewField()
	anim.Pos = o.pos.Sub(viewPos)
	anim.TerrainType = o.terrainType
	anim.Draw(t)
}

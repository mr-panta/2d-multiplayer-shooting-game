package entity

import (
	"github.com/faiface/pixel"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/animation"
)

var fieldShape = pixel.R(0, 0, 240, 240)

type Field struct {
	pos pixel.Vec
}

func NewField(pos pixel.Vec) *Field {
	return &Field{
		pos: pos,
	}
}

func (o *Field) GetShape() pixel.Rect {
	return fieldShape.Moved(o.pos)
}

func (o *Field) Render(t pixel.Target, viewPos pixel.Vec) {
	anim := animation.NewField()
	anim.Pos = o.pos.Add(fieldShape.Center()).Sub(viewPos)
	anim.Draw(t)
}

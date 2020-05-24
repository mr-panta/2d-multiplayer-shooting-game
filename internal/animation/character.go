package animation

import (
	"image/color"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"golang.org/x/image/colornames"
)

var (
	characterIdleFrames = []pixel.Rect{
		pixel.R(0*128, 0, 1*128, 128),
		pixel.R(1*128, 0, 2*128, 128),
		pixel.R(2*128, 0, 3*128, 128),
		pixel.R(3*128, 0, 4*128, 128),
	}
	characterMoveFrames = []pixel.Rect{
		pixel.R(4*128, 0, 5*128, 128),
		pixel.R(5*128, 0, 6*128, 128),
		pixel.R(6*128, 0, 7*128, 128),
		pixel.R(7*128, 0, 8*128, 128),
	}
	characterHitColor    = colornames.Red
	characterShadowColor = color.RGBA{0, 0, 0, 88}
)

const (
	// info
	characterWidth  = 128
	characterHeight = 128
	// States
	CharacterIdleState = 0
	CharacterRunState  = 1
)

type Character struct {
	shadowImd *imdraw.IMDraw
	Pos       pixel.Vec
	Right     bool
	State     int
	FrameTime int // in milliseconds
	Color     color.Color
	Hit       bool
	Shadow    bool
}

func NewCharacter() *Character {
	return &Character{
		shadowImd: imdraw.New(nil),
		FrameTime: 1,
	}
}

func (c *Character) Draw(target pixel.Target) {
	if c.Shadow {
		c.drawShadow(target)
	}
	c.draw(target)
}

func (c *Character) draw(target pixel.Target) {
	var frames []pixel.Rect
	switch c.State {
	case CharacterIdleState:
		frames = characterIdleFrames
	case CharacterRunState:
		frames = characterMoveFrames
	default:
		return
	}
	frame := frames[int((timeMS()/int64(c.FrameTime))%int64(len(frames)))]
	sprite := pixel.NewSprite(objectSheet, frame)
	matrix := pixel.IM.Moved(c.Pos)
	if c.Right {
		matrix = matrix.ScaledXY(c.Pos, pixel.V(-1, 1))
	}
	color := c.Color
	if c.Hit {
		color = characterHitColor
	}
	sprite.DrawColorMask(target, matrix, color)
}

func (c *Character) drawShadow(target pixel.Target) {
	matrix := pixel.IM.Moved(c.Pos)
	c.shadowImd.Clear()
	c.shadowImd.Color = characterShadowColor
	c.shadowImd.Push(pixel.V(0, -characterHeight/2))
	c.shadowImd.SetMatrix(matrix)
	c.shadowImd.Ellipse(pixel.V(characterWidth/5, characterHeight/12), 0)
	c.shadowImd.Draw(target)
}

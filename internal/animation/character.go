package animation

import (
	"image/color"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

const (
	// info
	characterWidth  = 128
	characterHeight = 128
	// States
	CharacterIdleState = 0
	CharacterRunState  = 1
)

var (
	characterSetFrames   = []int{4, 4}
	characterHitColor    = colornames.Red
	characterShadowColor = color.RGBA{0, 0, 0, 88}
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

func (c *Character) Draw(win *pixelgl.Window) {
	if c.Shadow {
		c.drawShadow(win)
	}
	c.draw(win)
}

func (c *Character) draw(win *pixelgl.Window) {
	spriteSet := len(characterSetFrames) - c.State - 1
	spriteFrame := int((timeMS() / int64(c.FrameTime)) % int64(characterSetFrames[spriteSet]))
	sprite := pixel.NewSprite(
		characterSheet,
		pixel.R(
			float64(characterWidth*spriteFrame),
			float64(characterHeight*spriteSet),
			float64(characterWidth*(spriteFrame+1)),
			float64(characterHeight*(spriteSet+1)),
		),
	)
	matrix := pixel.IM.Moved(c.Pos)
	if c.Right {
		matrix = matrix.ScaledXY(c.Pos, pixel.V(-1, 1))
	}
	color := c.Color
	if c.Hit {
		color = characterHitColor
	}
	sprite.DrawColorMask(win, matrix, color)
}

func (c *Character) drawShadow(win *pixelgl.Window) {
	// win.SetSmooth(false)
	matrix := pixel.IM.Moved(c.Pos)
	c.shadowImd.Clear()
	c.shadowImd.Color = characterShadowColor
	c.shadowImd.Push(pixel.V(0, -characterHeight/2))
	c.shadowImd.SetMatrix(matrix)
	c.shadowImd.Ellipse(pixel.V(characterWidth/5, characterHeight/12), 0)
	c.shadowImd.Draw(win)
}

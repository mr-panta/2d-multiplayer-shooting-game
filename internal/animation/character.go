package animation

import (
	"image/color"
	"math"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/ticktime"
	"golang.org/x/image/colornames"
)

var (
	characterIdleFrames = []pixel.Rect{
		pixel.R(0*128, 128, 1*128, 256),
		pixel.R(1*128, 128, 2*128, 256),
		pixel.R(2*128, 128, 3*128, 256),
		pixel.R(3*128, 128, 4*128, 256),
	}
	characterMoveFrames = []pixel.Rect{
		pixel.R(4*128, 128, 5*128, 256),
		pixel.R(5*128, 128, 6*128, 256),
		pixel.R(6*128, 128, 7*128, 256),
		pixel.R(7*128, 128, 8*128, 256),
	}
	characterDieFrames = []pixel.Rect{
		pixel.R(0*128, 0, 1*128, 128),
		pixel.R(1*128, 0, 2*128, 128),
		pixel.R(2*128, 0, 3*128, 128),
		pixel.R(3*128, 0, 4*128, 128),
		pixel.R(4*128, 0, 5*128, 128),
		pixel.R(5*128, 0, 6*128, 128),
		pixel.R(6*128, 0, 7*128, 128),
		pixel.R(7*128, 0, 8*128, 128),
	}
	characterHitColor          = colornames.Red
	characterArmorHitColor     = color.RGBA{0x1f, 0xa7, 0xff, 0xff}
	characterInvulnerableColor = color.RGBA{160, 160, 160, 160}
)

const (
	// info
	characterWidth  = 128
	characterHeight = 128
	// States
	CharacterIdleState = 0
	CharacterRunState  = 1
	CharacterDieState  = 2
	// dead duration
	characterDeadDuration = 800
)

type Character struct {
	shadowImd    *imdraw.IMDraw
	Pos          pixel.Vec
	Right        bool
	State        int
	FrameTime    int // in milliseconds
	Color        color.Color
	Hit          bool
	ArmorHit     bool
	Invulnerable bool
	Shadow       bool
	DieTime      time.Time
}

func NewCharacter() *Character {
	return &Character{
		shadowImd: imdraw.New(nil),
		FrameTime: 1,
	}
}

func (c *Character) Draw(target pixel.Target) {
	if c.Shadow && c.State != CharacterDieState {
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
	case CharacterDieState:
		frames = characterDieFrames
	default:
		return
	}
	keyFrame := int((ticktime.GetServerTimeMS() / int64(c.FrameTime)) % int64(len(frames)))
	if c.State == CharacterDieState {
		deadMS := ticktime.GetServerTime().Sub(c.DieTime).Seconds() * 1000
		keyFrame = int(math.Floor(deadMS * float64(len(frames)) / float64(characterDeadDuration)))
		if keyFrame >= len(frames) {
			// Render nothing
			return
		}
	}
	frame := frames[keyFrame]
	sprite := pixel.NewSprite(objectSheet, frame)
	matrix := pixel.IM.Moved(c.Pos)
	if c.Right {
		matrix = matrix.ScaledXY(c.Pos, pixel.V(-1, 1))
	}
	color := c.Color
	if c.Hit {
		color = characterHitColor
	} else if c.ArmorHit {
		color = characterArmorHitColor
	} else if c.Invulnerable {
		color = characterInvulnerableColor
	}
	sprite.DrawColorMask(target, matrix, color)
}

func (c *Character) drawShadow(target pixel.Target) {
	matrix := pixel.IM.Moved(c.Pos)
	c.shadowImd.Clear()
	c.shadowImd.Color = shadowColor
	c.shadowImd.Push(pixel.V(0, -characterHeight/2))
	c.shadowImd.SetMatrix(matrix)
	c.shadowImd.Ellipse(pixel.V(characterWidth/5, characterHeight/12), 0)
	c.shadowImd.Draw(target)
}

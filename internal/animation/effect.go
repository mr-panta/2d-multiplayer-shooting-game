package animation

import (
	"image/color"
	"time"

	"github.com/faiface/pixel"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/ticktime"
)

const (
	effectExplosionFrameTime = 100 * time.Millisecond
)

var (
	effectExplosionOffset = pixel.V(0, 960)
	effectExplosionFrames = []pixel.Rect{
		pixel.R(0, 0, 3*32, 2*32).Moved(effectExplosionOffset),
		pixel.R(3*32, 0, 9*32, 4*32).Moved(effectExplosionOffset),
		pixel.R(0, 4*32, 7*32, 8*32).Moved(effectExplosionOffset),
		pixel.R(9*32, 0, 17*32, 6*32).Moved(effectExplosionOffset),
		pixel.R(17*32, 0, 26*32, 7*32).Moved(effectExplosionOffset),
	}
)

type Effect struct {
	Pos       pixel.Vec
	Color     color.Color
	started   bool
	frames    []pixel.Rect
	frameTime time.Duration
	startTime time.Time
}

func NewEffectExplosion() *Effect {
	return &Effect{
		frames:    effectExplosionFrames,
		frameTime: effectExplosionFrameTime,
	}
}

func (e *Effect) Start() {
	e.started = true
	e.startTime = ticktime.GetServerTime()
}

func (e *Effect) Draw(target pixel.Target) {
	if !e.started {
		return
	}
	now := ticktime.GetServerTime()
	duration := now.Sub(e.startTime)
	index := int(duration / e.frameTime)
	if len(e.frames) <= index {
		e.started = false
		return
	}
	frame := e.frames[index]
	sprite := pixel.NewSprite(objectSheet, frame)
	matrix := pixel.IM.Moved(e.Pos.Add(pixel.V(0, frame.H()/2)))
	sprite.DrawColorMask(target, matrix, e.Color)
}

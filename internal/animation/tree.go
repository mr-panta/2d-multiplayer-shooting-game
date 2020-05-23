package animation

import (
	"image/color"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
)

var (
	// frame
	treeAFrame = pixel.R(0, 0, 100, 250)
	treeBFrame = pixel.R(100, 0, 300, 340)
	treeCFrame = pixel.R(300, 0, 600, 268)
	treeDFrame = pixel.R(600, 0, 730, 75)
	treeEFrame = pixel.R(600, 100, 730, 175)
	// shadow
	treeBShadow = pixel.V(80, 12)
	treeCShadow = pixel.V(100, 16)
)

func NewTreeA() *Tree {
	return &Tree{
		frame: treeAFrame,
	}
}

func NewTreeB() *Tree {
	return &Tree{
		frame:     treeBFrame,
		shadow:    treeBShadow,
		shadowImd: imdraw.New(nil),
	}
}

func NewTreeC() *Tree {
	return &Tree{
		frame:     treeCFrame,
		shadow:    treeCShadow,
		shadowImd: imdraw.New(nil),
	}
}

func NewTreeD() *Tree {
	return &Tree{
		frame: treeDFrame,
	}
}

func NewTreeE() *Tree {
	return &Tree{
		frame: treeEFrame,
	}
}

type Tree struct {
	frame     pixel.Rect
	shadow    pixel.Vec
	shadowImd *imdraw.IMDraw
	Pos       pixel.Vec
	Color     color.Color
	Right     bool
}

func (t *Tree) Draw(win *pixelgl.Window) {
	t.drawShadow(win)
	t.draw(win)
}

func (t *Tree) draw(win *pixelgl.Window) {
	sprite := pixel.NewSprite(treeSheet, t.frame)
	matrix := pixel.IM.Moved(t.Pos.Add(pixel.V(0, t.frame.H()/2)))
	if t.Right {
		matrix = matrix.ScaledXY(t.Pos, pixel.V(-1, 1))
	}
	sprite.DrawColorMask(win, matrix, t.Color)
}

func (t *Tree) drawShadow(win *pixelgl.Window) {
	if t.shadowImd != nil {
		matrix := pixel.IM.Moved(t.Pos)
		if t.Right {
			matrix = matrix.ScaledXY(t.Pos, pixel.V(-1, 1))
		}
		t.shadowImd.Clear()
		t.shadowImd.Color = characterShadowColor
		t.shadowImd.Push(pixel.V(0, 4))
		t.shadowImd.SetMatrix(matrix)
		t.shadowImd.Ellipse(t.shadow, 0)
		t.shadowImd.Draw(win)
	}
}

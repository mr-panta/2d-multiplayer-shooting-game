package animation

import (
	"image/color"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
)

var (
	// frame
	treeFrameOffset = pixel.V(0, 384)
	treeAFrame      = pixel.R(0, 0, 2*64, 4*64).Moved(treeFrameOffset)
	treeBFrame      = pixel.R(2*64, 0, 5*64, 4*64+32).Moved(treeFrameOffset)
	treeCFrame      = pixel.R(5*64+1, 0, 9*64-1, 4*64+32-1).Moved(treeFrameOffset)
	treeDFrame      = pixel.R(9*64, 0, 12*64, 2*64).Moved(treeFrameOffset)
	treeEFrame      = pixel.R(9*64, 2*64, 12*64, 4*64).Moved(treeFrameOffset)
	// shadow
	treeBShadow = pixel.V(80, 12)
	treeCShadow = pixel.V(100, 16)
	// shadow offset
	treeBShadowOffset = pixel.V(0, 4)
	treeCShadowOffset = pixel.V(0, 16)
	// transparent
	treeTransparentColor = color.RGBA{127, 127, 127, 127}
)

func NewTreeA() *Tree {
	return &Tree{
		frame: treeAFrame,
	}
}

func NewTreeB() *Tree {
	return &Tree{
		frame:        treeBFrame,
		shadow:       treeBShadow,
		shadowOffset: treeBShadowOffset,
		shadowImd:    imdraw.New(nil),
	}
}

func NewTreeC() *Tree {
	return &Tree{
		frame:        treeCFrame,
		shadow:       treeCShadow,
		shadowOffset: treeCShadowOffset,
		shadowImd:    imdraw.New(nil),
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
	frame        pixel.Rect
	shadow       pixel.Vec
	shadowOffset pixel.Vec
	shadowImd    *imdraw.IMDraw
	Pos          pixel.Vec
	Color        color.Color
	Right        bool
	Transparent  bool
}

func (t *Tree) Draw(target pixel.Target) {
	t.drawShadow(target)
	t.draw(target)
}

func (t *Tree) draw(target pixel.Target) {
	sprite := pixel.NewSprite(objectSheet, t.frame)
	matrix := pixel.IM.Moved(t.Pos.Add(pixel.V(0, t.frame.H()/2)))
	if t.Right {
		matrix = matrix.ScaledXY(t.Pos, pixel.V(-1, 1))
	}
	color := t.Color
	if t.Transparent {
		color = treeTransparentColor
	}
	sprite.DrawColorMask(target, matrix, color)
}

func (t *Tree) drawShadow(target pixel.Target) {
	if t.shadowImd != nil {
		matrix := pixel.IM.Moved(t.Pos)
		if t.Right {
			matrix = matrix.ScaledXY(t.Pos, pixel.V(-1, 1))
		}
		t.shadowImd.Clear()
		t.shadowImd.Color = shadowColor
		t.shadowImd.Push(t.shadowOffset)
		t.shadowImd.SetMatrix(matrix)
		t.shadowImd.Ellipse(t.shadow, 0)
		t.shadowImd.Draw(target)
	}
}

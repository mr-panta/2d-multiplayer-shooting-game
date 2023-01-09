package menu

import (
	"fmt"
	"image/color"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

type Button struct {
	win        *pixelgl.Window
	leftImd    *imdraw.IMDraw
	rightImd   *imdraw.IMDraw
	actived    bool
	hovered    bool
	focused    bool
	Color      color.Color
	FocusColor color.Color
	HoverColor color.Color
	Pos        pixel.Vec
	Size       float64
	Width      float64
	Height     float64
	Thickness  float64
	Label      string
}

func NewButton(win *pixelgl.Window) *Button {
	return &Button{
		win:        win,
		leftImd:    imdraw.New(nil),
		rightImd:   imdraw.New(nil),
		Color:      colornames.White,
		FocusColor: colornames.White,
		HoverColor: colornames.White,
	}
}

func (b *Button) Update() {
	r := pixel.R(0, 0, b.Width, b.Height).Moved(b.Pos)
	b.hovered = r.Contains(b.win.MousePosition())
	b.actived = b.hovered && b.win.JustReleased(pixelgl.MouseButton1)
	b.focused = b.hovered && b.win.Pressed(pixelgl.MouseButton1)
}

func (b *Button) Render() {
	b.renderLeft()
	b.renderRight()
	b.renderLabel()
}

func (b *Button) renderLabel() {
	r := pixel.R(0, 0, b.Width, b.Height).Moved(b.Pos)
	atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	txt := text.New(pixel.ZV, atlas)
	txt.Clear()
	txt.LineHeight = atlas.LineHeight()
	if b.focused {
		txt.Color = b.FocusColor
	} else if b.hovered {
		txt.Color = b.HoverColor
	} else {
		txt.Color = b.Color
	}
	fmt.Fprintf(txt, b.Label)
	txt.Draw(b.win, pixel.IM.
		Moved(r.Center().Sub(pixel.V(txt.Bounds().W()/2, txt.Bounds().H()/2))).
		Scaled(r.Center(), b.Size),
	)
}

func (b *Button) renderLeft() {
	r := pixel.R(0, 0, b.Width, b.Height).Moved(b.Pos)
	imd := b.leftImd
	imd.Clear()
	if b.focused {
		imd.Color = b.FocusColor
	} else if b.hovered {
		imd.Color = b.HoverColor
	} else {
		imd.Color = b.Color
	}
	pos := r.Vertices()[1]
	imd.Push(
		pos.Add(pixel.V(b.Height/3, 0)),
		pos,
		pos.Add(pixel.V(0, -b.Height/3)),
	)
	imd.Line(b.Thickness)
	imd.Draw(b.win)
}

func (b *Button) renderRight() {
	r := pixel.R(0, 0, b.Width, b.Height).Moved(b.Pos)
	imd := b.rightImd
	imd.Clear()
	if b.focused {
		imd.Color = b.FocusColor
	} else if b.hovered {
		imd.Color = b.HoverColor
	} else {
		imd.Color = b.Color
	}
	pos := r.Vertices()[3]
	imd.Push(
		pos.Add(pixel.V(-b.Height/3, 0)),
		pos,
		pos.Add(pixel.V(0, b.Height/3)),
	)
	imd.Line(b.Thickness)
	imd.Draw(b.win)
}

func (b *Button) Actived() bool {
	return b.actived
}

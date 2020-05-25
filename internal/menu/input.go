package menu

import (
	"fmt"
	"image/color"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/ticktime"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

const (
	inputCursorPaddingLeft  = 8
	inputValuePaddingBottom = 12
	inputCursorBlinkTimeMS  = 300
)

var inputLabelColor = color.RGBA{63, 63, 63, 255}

type Input struct {
	win        *pixelgl.Window
	lineImd    *imdraw.IMDraw
	cursorImd  *imdraw.IMDraw
	value      string
	focused    bool
	hovered    bool
	Color      color.Color
	HoverColor color.Color
	FocusColor color.Color
	Pos        pixel.Vec
	Size       float64
	Width      float64
	Height     float64
	Thickness  float64
	Label      string
}

func NewInput(win *pixelgl.Window) *Input {
	return &Input{
		win:        win,
		lineImd:    imdraw.New(nil),
		cursorImd:  imdraw.New(nil),
		Color:      colornames.White,
		FocusColor: colornames.White,
		HoverColor: colornames.White,
	}
}

func (i *Input) contains(pos pixel.Vec) bool {
	r := pixel.R(0, 0, i.Width, i.Height).Moved(i.Pos)
	return r.Contains(pos)
}

func (i *Input) Update() {
	pos := i.win.MousePosition()
	i.hovered = i.contains(pos)
	if i.win.Pressed(pixelgl.MouseButton1) {
		i.focused = i.hovered
	}
	if i.focused {
		text := i.win.Typed()
		if len(text) > 0 {
			i.value += text
		}
		if i.win.JustPressed(pixelgl.KeyBackspace) || i.win.Repeated(pixelgl.KeyBackspace) {
			if len(i.value) > 0 {
				i.value = i.value[:len(i.value)-1]
			}
		}
	}
}

func (i *Input) Render() {
	if len(i.value) == 0 {
		i.renderLabel()
	}
	i.renderLine()
	i.renderValue()
	if i.focused {
		i.renderCursor()
	}
}

func (i *Input) GetValue() string {
	return i.value
}

func (i *Input) getValueText() *text.Text {
	atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	txt := text.New(pixel.ZV, atlas)
	txt.Clear()
	txt.LineHeight = atlas.LineHeight()
	txt.Color = i.Color
	fmt.Fprintf(txt, i.value)
	return txt
}

func (i *Input) getCenterMatrix(txt *text.Text) pixel.Matrix {
	return pixel.IM.
		Moved(pixel.V(-txt.Bounds().W()/2, 0)).
		Scaled(pixel.ZV, i.Size).
		Moved(i.Pos.Add(pixel.V(i.Width/2, inputValuePaddingBottom)))
}

func (i *Input) renderValue() {
	txt := i.getValueText()
	txt.Draw(i.win, i.getCenterMatrix(txt))
}

func (i *Input) renderLabel() {
	atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	txt := text.New(pixel.ZV, atlas)
	txt.Clear()
	txt.LineHeight = atlas.LineHeight()
	txt.Color = inputLabelColor
	fmt.Fprintf(txt, i.Label)
	txt.Draw(i.win, i.getCenterMatrix(txt))
}

func (i *Input) renderLine() {
	i.lineImd.Clear()
	if i.focused {
		i.lineImd.Color = i.FocusColor
	} else if i.hovered {
		i.lineImd.Color = i.HoverColor
	} else {
		i.lineImd.Color = i.Color
	}
	i.lineImd.Push(pixel.ZV, pixel.V(i.Width, 0))
	i.lineImd.SetMatrix(pixel.IM.Moved(i.Pos))
	i.lineImd.Line(i.Thickness)
	i.lineImd.Draw(i.win)
}

func (i *Input) renderCursor() {
	timeMS := ticktime.GetServerTimeMS()
	show := (timeMS / inputCursorBlinkTimeMS) % 2
	if show == 0 {
		return
	}
	txt := i.getValueText()
	pos := txt.Bounds().Max.Add(i.Pos).Add(pixel.V(i.Width/2, -txt.Bounds().H()))
	if len(i.value) > 0 {
		pos = pos.Add(pixel.V(inputCursorPaddingLeft, 0))
	}
	i.cursorImd.Clear()
	i.cursorImd.Color = i.Color
	i.cursorImd.Push(pixel.V(0, inputValuePaddingBottom), pixel.V(0, i.Height))
	i.cursorImd.SetMatrix(pixel.IM.Moved(pos))
	i.cursorImd.Line(i.Thickness)
	i.cursorImd.Draw(i.win)
}

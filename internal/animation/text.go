package animation

import (
	"fmt"
	"image/color"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

var (
	atlas      = text.NewAtlas(basicfont.Face7x13, text.ASCII)
	defaultTxt = text.New(pixel.ZV, atlas)
)

func NewText() *text.Text {
	return text.New(pixel.ZV, atlas)
}

func GetTextCenterBounds(pos pixel.Vec, value string, size float64) pixel.Rect {
	txt := defaultTxt
	txt.Clear()
	txt.LineHeight = atlas.LineHeight()
	fmt.Fprintf(txt, value)
	rect := txt.Bounds().
		Moved(pixel.ZV.Sub(pixel.V(txt.Bounds().W()/2, 0))).
		Moved(pos)
	return pixel.Rect{
		Min: rect.Min,
		Max: rect.Min.Add(pixel.V(rect.W()*size, rect.H()*size)),
	}
}

func DrawStrokeTextCenter(txt *text.Text, target pixel.Target, pos pixel.Vec, value string, size float64,
	color, strokeColor color.Color) {
	txt.Clear()
	txt.LineHeight = atlas.LineHeight()
	txt.Color = colornames.Black
	if strokeColor != nil {
		txt.Color = strokeColor
	}
	fmt.Fprintf(txt, value)
	m := pixel.IM.
		Moved(pixel.ZV.Sub(pixel.V(txt.Bounds().W()/2, 0))).
		Scaled(pixel.ZV, size).
		Moved(pos)
	txt.Draw(target, m.Moved(pixel.V(0, 0.5).Scaled(size)))
	txt.Draw(target, m.Moved(pixel.V(0, -0.5).Scaled(size)))
	txt.Draw(target, m.Moved(pixel.V(0.5, 0).Scaled(size)))
	txt.Draw(target, m.Moved(pixel.V(-0.5, 0).Scaled(size)))
	txt.Color = colornames.White
	if color != nil {
		txt.Color = color
	}
	fmt.Fprintf(txt, "\r%s", value)
	txt.Draw(target, m)

}

func GetTextLeftBounds(pos pixel.Vec, value string, size float64) pixel.Rect {
	txt := defaultTxt
	txt.Clear()
	txt.LineHeight = atlas.LineHeight()
	fmt.Fprintf(txt, value)
	rect := txt.Bounds().Moved(pos)
	return pixel.Rect{
		Min: rect.Min,
		Max: rect.Min.Add(pixel.V(rect.W()*size, rect.H()*size)),
	}
}

func DrawShadowTextLeft(txt *text.Text, target pixel.Target, pos pixel.Vec, value string, size float64) {
	txt.Clear()
	txt.LineHeight = atlas.LineHeight()
	txt.Color = colornames.Black
	fmt.Fprintf(txt, value)
	m := pixel.IM.Scaled(pixel.ZV, size).Moved(pos)
	txt.Draw(target, m.Moved(pixel.V(0.5, -0.5).Scaled(size)))
	txt.Color = colornames.White
	fmt.Fprintf(txt, "\r%s", value)
	txt.Draw(target, m)
}

func GetTextRightBounds(pos pixel.Vec, value string, size float64) pixel.Rect {
	txt := defaultTxt
	txt.Clear()
	txt.LineHeight = atlas.LineHeight()
	fmt.Fprintf(txt, value)
	rect := txt.Bounds().
		Moved(pixel.ZV.Sub(pixel.V(txt.Bounds().W(), 0))).
		Moved(pos)
	return pixel.Rect{
		Min: rect.Min,
		Max: rect.Min.Add(pixel.V(rect.W()*size, rect.H()*size)),
	}
}

func DrawShadowTextRight(txt *text.Text, target pixel.Target, pos pixel.Vec, value string, size float64) {
	txt.Clear()
	txt.LineHeight = atlas.LineHeight()
	txt.Color = colornames.Black
	fmt.Fprintf(txt, value)
	m := pixel.IM.
		Moved(pixel.ZV.Sub(pixel.V(txt.Bounds().W(), 0))).
		Scaled(pixel.ZV, size).
		Moved(pos)
	txt.Draw(target, m.Moved(pixel.V(0.5, -0.5).Scaled(size)))
	txt.Color = colornames.White
	fmt.Fprintf(txt, "\r%s", value)
	txt.Draw(target, m)
}

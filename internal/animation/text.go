package animation

import (
	"fmt"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

func DrawStrokeTextCenter(target pixel.Target, pos pixel.Vec, value string, size float64) {
	atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	txt := text.New(pixel.ZV, atlas)
	txt.Clear()
	txt.LineHeight = atlas.LineHeight()
	txt.Color = colornames.Black
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
	fmt.Fprintf(txt, "\r%s", value)
	txt.Draw(target, m)
}

func DrawShadowTextLeft(target pixel.Target, pos pixel.Vec, value string, size float64) {
	atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	txt := text.New(pixel.ZV, atlas)
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

func DrawShadowTextRight(target pixel.Target, pos pixel.Vec, value string, size float64) {
	atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	txt := text.New(pixel.ZV, atlas)
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

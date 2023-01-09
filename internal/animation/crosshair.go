package animation

import (
	"image/color"
	"math"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"golang.org/x/image/colornames"
)

var (
	crosshairLine  = pixel.L(pixel.V(4, 0), pixel.V(16, 0))
	crosshairColor = colornames.Green
)

const crosshairThickness = 4

type Crosshair struct {
	batch      *pixel.Batch
	lines      [4]*imdraw.IMDraw
	shadowList [4]*imdraw.IMDraw
	Pos        pixel.Vec
	Color      color.Color
}

func NewCrosshair() *Crosshair {
	return &Crosshair{
		batch: pixel.NewBatch(&pixel.TrianglesData{}, objectSheet),
		lines: [4]*imdraw.IMDraw{
			imdraw.New(nil),
			imdraw.New(nil),
			imdraw.New(nil),
			imdraw.New(nil),
		},
		shadowList: [4]*imdraw.IMDraw{
			imdraw.New(nil),
			imdraw.New(nil),
			imdraw.New(nil),
			imdraw.New(nil),
		},
	}
}

func (c *Crosshair) Draw(target pixel.Target) {
	c.batch.Clear()
	for i := range c.lines {
		c.drawLineShadow(c.batch, i)
	}
	for i := range c.lines {
		c.drawLine(c.batch, i)
	}
	c.batch.Draw(target)
}

func (c *Crosshair) drawLineShadow(target pixel.Target, i int) {
	m := pixel.IM.Rotated(pixel.ZV, float64(i)*math.Pi/2.0).Moved(c.Pos)
	c.shadowList[i].Clear()
	c.shadowList[i].Color = colornames.Black
	c.shadowList[i].Push(crosshairLine.A, crosshairLine.B)
	c.shadowList[i].Line(crosshairThickness)
	c.shadowList[i].SetMatrix(m.Moved(pixel.V(1, -1)))
	c.shadowList[i].Draw(target)
}

func (c *Crosshair) drawLine(target pixel.Target, i int) {
	m := pixel.IM.Rotated(pixel.ZV, float64(i)*math.Pi/2.0).Moved(c.Pos)
	c.lines[i].Clear()
	if c.Color == nil {
		c.lines[i].Color = crosshairColor
	} else {
		c.lines[i].Color = c.Color
	}
	c.lines[i].Push(crosshairLine.A, crosshairLine.B)
	c.lines[i].Line(crosshairThickness)
	c.lines[i].SetMatrix(m)
	c.lines[i].Draw(target)
}

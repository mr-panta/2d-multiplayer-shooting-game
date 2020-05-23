package util

import (
	"math/rand"

	"github.com/faiface/pixel"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/protocol"
)

func V(x, y float64) *protocol.Vec {
	return &protocol.Vec{
		X: x,
		Y: y,
	}
}

func ConvertVec(v pixel.Vec) *protocol.Vec {
	return &protocol.Vec{
		X: v.X,
		Y: v.Y,
	}
}

func RandomVec(r pixel.Rect) pixel.Vec {
	x := r.Min.X + rand.Float64()*r.W()
	y := r.Min.Y + rand.Float64()*r.H()
	return pixel.V(x, y)
}

func GetHighVec() pixel.Vec {
	return pixel.V(1000000, 1000000)
}

func LerpScalar(a, b, t float64) float64 {
	return (b-a)*t + a
}

func CheckCollision(fixedObject, prevCollider, nextCollider pixel.Rect) (staticAdjust, dynamicAdjust pixel.Vec) {
	if fixedObject.Intersect(nextCollider).Area() == 0 {
		return pixel.ZV, pixel.ZV
	}
	prevVertices := prevCollider.Vertices()
	nextVertices := nextCollider.Vertices()
	edges := fixedObject.Edges()
	staticAdjust = pixel.ZV
	for i := 0; i < 4; i++ {
		prev := prevVertices[i]
		next := nextVertices[i]
		line := pixel.L(prev, next)
		for j := 0; j < 4; j++ {
			edge := edges[j]
			vec, exists := edge.Intersect(line)
			if exists && vec.Len() > 0 {
				if len := next.Sub(vec).Len(); len > staticAdjust.Len() {
					staticAdjust = next.Sub(vec)
				}
			}
		}
	}
	dynamicAdjust = staticAdjust.Project(pixel.V(1, 0))
	if r := nextCollider.Moved(pixel.ZV.Sub(dynamicAdjust)); fixedObject.Intersect(r).Area() == 0 {
		return staticAdjust, dynamicAdjust
	}
	dynamicAdjust = staticAdjust.Project(pixel.V(0, 1))
	if r := nextCollider.Moved(pixel.ZV.Sub(dynamicAdjust)); fixedObject.Intersect(r).Area() == 0 {
		return staticAdjust, dynamicAdjust
	}
	staticAdjust = nextCollider.Center().Sub(prevCollider.Center())
	return staticAdjust, staticAdjust
}

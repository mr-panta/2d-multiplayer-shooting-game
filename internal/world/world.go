package world

import (
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/animation"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/common"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/entity"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/ticktime"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/util"
)

var (
	worldFieldSize = pixel.R(0, 0, 64, 64)
)

const (
	minNextItemPerd       = 10
	maxNextItemPerd       = 20
	worldFieldWidth       = 32
	worldFieldHeight      = 16
	worldTreeAmount       = 64
	worldTerrainAmount    = 16
	worldMinSpawnDist     = 48
	worldMinWindowRenderZ = 1000
)

type world struct {
	// common
	objectDB common.ObjectDB
	// client
	win           *pixelgl.Window
	batch         *pixel.Batch
	currRawInput  *common.RawInput
	prevRawInput  *common.RawInput
	mainPlayerID  string
	cameraPos     pixel.Vec
	hud           common.Hud
	scope         common.Scope
	water         common.Water
	fps           int
	frameCount    int
	fpsUpdateTime time.Time
	// server
	tick         int64
	nextItemTime time.Time
}

func New(clientProcessor common.ClientProcessor) common.World {
	world := &world{
		// common
		objectDB: common.NewObjectDB(),
		// client
		currRawInput: &common.RawInput{},
		prevRawInput: &common.RawInput{},
		// server
		nextItemTime: ticktime.GetServerTime(),
	}
	// common
	world.setupBoundaries()
	if clientProcessor != nil {
		// client
		world.win = clientProcessor.GetWindow()
		world.batch = pixel.NewBatch(&pixel.TrianglesData{}, animation.GetObjectSheet())
		world.hud = entity.NewHud(world)
		world.scope = entity.NewScope(world)
		world.water = entity.NewWater(world)
		world.fpsUpdateTime = ticktime.GetServerTime()
	} else {
		// server
		world.createTrees()
		world.createTerrains()
	}
	return world
}

func (w *world) GetObjectDB() common.ObjectDB {
	return w.objectDB
}

func (w *world) CheckCollision(id string, prevCollider, nextCollider pixel.Rect) (
	obj common.Object, staticAdjust, dynamicAdjust pixel.Vec) {
	count := 0
	for _, o := range w.objectDB.SelectAll() {
		if !o.Exists() || o.GetID() == id {
			continue
		}
		if collider, exists := o.GetCollider(); exists {
			static, dynamic := util.CheckCollision(
				collider,
				prevCollider,
				nextCollider,
			)
			if static.Len() > 0 {
				obj = o
				staticAdjust = static
				dynamicAdjust = dynamic
				count++
			}
			if count > 1 {
				staticAdjust = nextCollider.Center().Sub(prevCollider.Center())
				dynamicAdjust = nextCollider.Center().Sub(prevCollider.Center())
				break
			}
		}
	}
	return obj, staticAdjust, dynamicAdjust
}

func (w *world) GetSize() (width, height int) {
	return worldFieldWidth, worldFieldHeight
}

func (w *world) getSizeRect() pixel.Rect {
	return pixel.R(
		0, 0,
		float64(worldFieldWidth)*worldFieldSize.W(),
		float64(worldFieldHeight)*worldFieldSize.H(),
	)
}

func (w *world) setupBoundaries() {
	size := w.getSizeRect()
	w.objectDB.Set(entity.NewBoundary(w, w.objectDB.GetAvailableID(), pixel.Rect{
		Min: pixel.V(size.Min.X-200, size.Min.Y-200),
		Max: pixel.V(size.Max.X+200, size.Min.Y),
	}))
	w.objectDB.Set(entity.NewBoundary(w, w.objectDB.GetAvailableID(), pixel.Rect{
		Min: pixel.V(size.Min.X-200, size.Max.Y),
		Max: pixel.V(size.Max.X+200, size.Max.Y+200),
	}))
	w.objectDB.Set(entity.NewBoundary(w, w.objectDB.GetAvailableID(), pixel.Rect{
		Min: pixel.V(size.Min.X-200, size.Min.Y-200),
		Max: pixel.V(size.Min.X, size.Max.Y+200),
	}))
	w.objectDB.Set(entity.NewBoundary(w, w.objectDB.GetAvailableID(), pixel.Rect{
		Min: pixel.V(size.Max.X, size.Min.Y-200),
		Max: pixel.V(size.Max.X+200, size.Max.Y+200),
	}))
}

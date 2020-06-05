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
	minNextItemPerd  = 10
	maxNextItemPerd  = 20
	worldFieldWidth  = 30
	worldFieldHeight = 30
	worldTreeAmount  = 36
	// worldFieldWidth       = 8
	// worldFieldHeight      = 8
	// worldTreeAmount       = 0
	worldTerrainAmount    = 12
	worldMinSpawnDist     = 48
	worldMinWindowRenderZ = 1000
	worldBoundarySize     = 200
)

type world struct {
	// common
	objectDB    common.ObjectDB
	hud         common.Hud
	fieldWidth  int
	fieldHeight int
	// client
	win              *pixelgl.Window
	toggleFPSLimit   func()
	batch            *pixel.Batch
	currRawInput     *common.RawInput
	prevRawInput     *common.RawInput
	currSettingInput *common.RawInput
	prevSettingInput *common.RawInput
	mainPlayerID     string
	cameraPos        pixel.Vec
	scope            common.Scope
	water            common.Water
	fps              int
	frameCount       int
	fpsUpdateTime    time.Time
	// server
	tick         int64
	nextItemTime time.Time
}

func New(clientProcessor common.ClientProcessor) common.World {
	world := &world{
		// common
		objectDB:    common.NewObjectDB(),
		fieldWidth:  worldFieldWidth,
		fieldHeight: worldFieldHeight,
		// client
		currRawInput:     &common.RawInput{},
		prevRawInput:     &common.RawInput{},
		currSettingInput: &common.RawInput{},
		prevSettingInput: &common.RawInput{},
		// server
		nextItemTime: ticktime.GetServerTime(),
	}
	// common
	world.hud = entity.NewHud(world)
	if clientProcessor != nil {
		// client
		world.win = clientProcessor.GetWindow()
		world.toggleFPSLimit = clientProcessor.ToggleFPSLimit
		world.batch = pixel.NewBatch(&pixel.TrianglesData{}, animation.GetObjectSheet())
		world.scope = entity.NewScope(world)
		world.water = entity.NewWater(world)
		world.fpsUpdateTime = ticktime.GetServerTime()
	} else {
		// server
		world.createTrees()
		world.createTerrains()
		world.createBoundaries()
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

func (w *world) GetHud() common.Hud {
	return w.hud
}

func (w *world) GetSize() (width, height int) {
	return w.fieldWidth, w.fieldHeight
}

func (w *world) getSizeRect() pixel.Rect {
	return pixel.R(
		0, 0,
		float64(w.fieldWidth)*worldFieldSize.W(),
		float64(w.fieldHeight)*worldFieldSize.H(),
	)
}

package world

import (
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/common"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/entity"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/ticktime"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/util"
)

type world struct {
	// common
	field    pixel.Rect
	objectDB common.ObjectDB
	// client
	core         common.Core
	win          *pixelgl.Window
	origin       *imdraw.IMDraw
	currRawInput *common.RawInput
	prevRawInput *common.RawInput
	mainPlayerID string
	cameraPos    pixel.Vec
	hud          common.Hud
	scope        common.Scope
	// server
	tick         int64
	nextItemTime time.Time
}

func New(core common.Core) common.World {
	world := &world{
		// common
		field:    pixel.R(-240, -240, 240, 240),
		objectDB: common.NewObjectDB(),
		// client
		currRawInput: &common.RawInput{},
		prevRawInput: &common.RawInput{},
		// server
		nextItemTime: ticktime.GetServerTime(),
	}
	// client
	if core != nil {
		world.core = core
		world.win = core.GetWindow()
		world.origin = imdraw.New(nil)
		world.hud = entity.NewHud(world)
		world.scope = entity.NewScope(world)
	}
	return world
}

func (w *world) GetObjectDB() common.ObjectDB {
	return w.objectDB
}

func (w *world) CheckCollision(id string, prevCollider, nextCollider pixel.Rect) (
	obj common.Object, staticAdjust, dynamicAdjust pixel.Vec) {
	for _, o := range w.objectDB.SelectAll() {
		if !o.Exists() || o.GetID() == id {
			continue
		}
		if collider, exists := o.GetCollider(); exists {
			staticAdjust, dynamicAdjust := util.CheckCollision(
				collider,
				prevCollider,
				nextCollider,
			)
			if staticAdjust.Len() > 0 {
				return o, staticAdjust, dynamicAdjust
			}
		}
	}
	return nil, pixel.ZV, pixel.ZV
}

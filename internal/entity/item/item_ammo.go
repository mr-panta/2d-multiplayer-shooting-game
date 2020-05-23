package item

import (
	"sync"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/animation"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/common"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/config"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/protocol"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/ticktime"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/util"
)

var (
	itemAmmoShape = pixel.R(0, 0, 45, 47)
)

type ItemAmmo struct {
	id            string
	world         common.World
	pos           pixel.Vec
	isDestroyed   bool
	tickSnapshots []*protocol.TickSnapshot
	lock          sync.RWMutex
}

func NewItemAmmo(world common.World, id string) *ItemAmmo {
	return &ItemAmmo{
		id:    id,
		world: world,
		pos:   util.GetHighVec(),
	}
}

func (o *ItemAmmo) GetID() string {
	return o.id
}

func (o *ItemAmmo) Destroy() {
	o.isDestroyed = true
}

func (o *ItemAmmo) Exists() bool {
	return !o.isDestroyed
}

func (o *ItemAmmo) SetPos(pos pixel.Vec) {
	o.pos = pos
}

func (o *ItemAmmo) GetShape() pixel.Rect {
	return itemAmmoShape.Moved(o.pos.Sub(itemAmmoShape.Center()))
}

func (o *ItemAmmo) GetCollider() (pixel.Rect, bool) {
	return pixel.ZR, false
}

func (o *ItemAmmo) GetRenderObjects() []common.RenderObject {
	return []common.RenderObject{common.NewRenderObject(itemZ, o.GetShape(), o.render)}
}

func (o *ItemAmmo) SetSnapshot(tick int64, ss *protocol.ObjectSnapshot) {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.tickSnapshots = append(o.tickSnapshots, &protocol.TickSnapshot{
		Tick:     tick,
		Snapshot: ss,
	})
}

func (o *ItemAmmo) GetSnapshot(tick int64) (ss *protocol.ObjectSnapshot) {
	o.lock.RLock()
	defer o.lock.RUnlock()
	for i := len(o.tickSnapshots) - 1; i >= 0; i-- {
		ts := o.tickSnapshots[i]
		if ts.Tick == tick {
			ss = ts.Snapshot
		} else if ts.Tick < tick {
			break
		}
	}
	if ss == nil {
		ss = o.getCurrentSnapshot()
	}
	return ss
}

func (o *ItemAmmo) ServerUpdate(tick int64) {
	o.SetSnapshot(tick, o.getCurrentSnapshot())
	o.cleanTickSnapshots()
}

func (o *ItemAmmo) ClientUpdate() {
	ss := o.getLerpSnapshot().Item.Ammo
	o.pos = ss.Pos.Convert()
	o.cleanTickSnapshots()
}

func (o *ItemAmmo) UsedBy(p common.Player) (ok bool) {
	if p.GetWeapon() != nil && p.GetWeapon().AddAmmo(-1) {
		return true
	}
	return false
}

func (o *ItemAmmo) GetType() int {
	return config.ItemObject
}

func (o *ItemAmmo) cleanTickSnapshots() {
	o.lock.Lock()
	defer o.lock.Unlock()
	if len(o.tickSnapshots) <= 1 {
		return
	}
	t := ticktime.GetServerTime().Add(-config.LerpPeriod * 2)
	tick := ticktime.GetTick(t)
	index := 0
	for i, ts := range o.tickSnapshots {
		if ts.Tick >= tick {
			index = i
			break
		}
	}
	if index > 0 {
		o.tickSnapshots = o.tickSnapshots[index:]
	}
}

func (o *ItemAmmo) getCurrentSnapshot() *protocol.ObjectSnapshot {
	return &protocol.ObjectSnapshot{
		ID:   o.GetID(),
		Type: o.GetType(),
		Item: &protocol.ItemSnapshot{
			Ammo: &protocol.ItemAmmoSnapshot{
				Pos: util.ConvertVec(o.pos),
			},
		},
	}
}

func (o *ItemAmmo) render(win *pixelgl.Window, viewPos pixel.Vec) {
	anim := animation.NewItemAmmo()
	anim.Pos = o.pos.Sub(viewPos)
	anim.Draw(win)
}

func (o *ItemAmmo) getLerpSnapshot() *protocol.ObjectSnapshot {
	return o.getSnapshotsByTime(ticktime.GetLerpTime())
}

func (o *ItemAmmo) getSnapshotsByTime(t time.Time) *protocol.ObjectSnapshot {
	o.lock.RLock()
	defer o.lock.RUnlock()
	a, b, d := protocol.GetSnapshotByTime(t, o.tickSnapshots)
	if a == nil || b == nil {
		a = o.getCurrentSnapshot()
		b = o.getCurrentSnapshot()
	}
	ssA := a.Item.Ammo
	ssB := b.Item.Ammo
	return &protocol.ObjectSnapshot{
		ID:   o.GetID(),
		Type: o.GetType(),
		Item: &protocol.ItemSnapshot{
			Ammo: &protocol.ItemAmmoSnapshot{
				Pos: util.ConvertVec(pixel.Lerp(ssA.Pos.Convert(), ssB.Pos.Convert(), d)),
			},
		},
	}
}

package item

import (
	"sync"
	"time"

	"github.com/faiface/pixel"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/animation"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/common"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/config"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/protocol"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/ticktime"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/util"
)

var (
	itemAmmoSMShape = pixel.R(0, 0, 20, 26)
)

type ItemAmmoSM struct {
	id            string
	world         common.World
	pos           pixel.Vec
	isDestroyed   bool
	tickSnapshots []*protocol.TickSnapshot
	lock          sync.RWMutex
}

func NewItemAmmoSM(world common.World, id string) *ItemAmmoSM {
	return &ItemAmmoSM{
		id:    id,
		world: world,
		pos:   util.GetHighVec(),
	}
}

func (o *ItemAmmoSM) GetID() string {
	return o.id
}

func (o *ItemAmmoSM) Destroy() {
	o.isDestroyed = true
}

func (o *ItemAmmoSM) Exists() bool {
	return !o.isDestroyed
}

func (o *ItemAmmoSM) SetPos(pos pixel.Vec) {
	o.pos = pos
}

func (o *ItemAmmoSM) GetShape() pixel.Rect {
	return itemAmmoSMShape.Moved(o.pos.Sub(pixel.V(itemAmmoSMShape.W()/2, 0)))
}

func (o *ItemAmmoSM) GetCollider() (pixel.Rect, bool) {
	return pixel.ZR, false
}

func (o *ItemAmmoSM) GetRenderObjects() []common.RenderObject {
	return []common.RenderObject{common.NewRenderObject(itemZ, o.GetShape(), o.render)}
}

func (o *ItemAmmoSM) SetSnapshot(tick int64, ss *protocol.ObjectSnapshot) {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.tickSnapshots = append(o.tickSnapshots, &protocol.TickSnapshot{
		Tick:     tick,
		Snapshot: ss,
	})
}

func (o *ItemAmmoSM) GetSnapshot(tick int64) (ss *protocol.ObjectSnapshot) {
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

func (o *ItemAmmoSM) ServerUpdate(tick int64) {
	o.SetSnapshot(tick, o.getCurrentSnapshot())
	o.cleanTickSnapshots()
}

func (o *ItemAmmoSM) ClientUpdate() {
	ss := o.getLerpSnapshot().Item.AmmoSM
	o.pos = ss.Pos.Convert()
	o.cleanTickSnapshots()
}

func (o *ItemAmmoSM) UsedBy(p common.Player) (ok bool) {
	if p.GetWeapon() != nil && p.GetWeapon().AddAmmo(-2) {
		return true
	}
	return false
}

func (o *ItemAmmoSM) GetType() int {
	return config.ItemObject
}

func (o *ItemAmmoSM) cleanTickSnapshots() {
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

func (o *ItemAmmoSM) getCurrentSnapshot() *protocol.ObjectSnapshot {
	return &protocol.ObjectSnapshot{
		ID:   o.GetID(),
		Type: o.GetType(),
		Item: &protocol.ItemSnapshot{
			AmmoSM: &protocol.ItemAmmoSMSnapshot{
				Pos: util.ConvertVec(o.pos),
			},
		},
	}
}

func (o *ItemAmmoSM) render(target pixel.Target, viewPos pixel.Vec) {
	anim := animation.NewItemAmmoSM()
	anim.Pos = o.pos.Sub(viewPos)
	anim.Draw(target)
}

func (o *ItemAmmoSM) getLerpSnapshot() *protocol.ObjectSnapshot {
	return o.getSnapshotsByTime(ticktime.GetLerpTime())
}

func (o *ItemAmmoSM) getSnapshotsByTime(t time.Time) *protocol.ObjectSnapshot {
	o.lock.RLock()
	defer o.lock.RUnlock()
	a, b, d := protocol.GetSnapshotByTime(t, o.tickSnapshots)
	if a == nil || b == nil {
		a = o.getCurrentSnapshot()
		b = o.getCurrentSnapshot()
	}
	ssA := a.Item.AmmoSM
	ssB := b.Item.AmmoSM
	return &protocol.ObjectSnapshot{
		ID:   o.GetID(),
		Type: o.GetType(),
		Item: &protocol.ItemSnapshot{
			AmmoSM: &protocol.ItemAmmoSMSnapshot{
				Pos: util.ConvertVec(pixel.Lerp(ssA.Pos.Convert(), ssB.Pos.Convert(), d)),
			},
		},
	}
}

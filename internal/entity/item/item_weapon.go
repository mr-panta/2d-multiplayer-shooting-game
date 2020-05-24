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
	itemWeaponShape = pixel.R(0, 0, 45, 47)
)

type ItemWeapon struct {
	id            string
	weaponID      string
	world         common.World
	pos           pixel.Vec
	createTime    time.Time
	isDestroyed   bool
	tickSnapshots []*protocol.TickSnapshot
	lock          sync.RWMutex
}

func NewItemWeapon(world common.World, id string, weaponID string) *ItemWeapon {
	return &ItemWeapon{
		id:         id,
		weaponID:   weaponID,
		world:      world,
		pos:        util.GetHighVec(),
		createTime: ticktime.GetServerTime(),
	}
}

func (w *ItemWeapon) GetID() string {
	return w.id
}

func (w *ItemWeapon) Destroy() {
	w.isDestroyed = true
}

func (w *ItemWeapon) Exists() bool {
	return !w.isDestroyed
}

func (w *ItemWeapon) SetPos(pos pixel.Vec) {
	w.pos = pos
}

func (w *ItemWeapon) GetShape() pixel.Rect {
	return itemWeaponShape.Moved(w.pos.Sub(pixel.V(itemWeaponShape.W()/2, 0)))
}

func (w *ItemWeapon) GetCollider() (pixel.Rect, bool) {
	return pixel.ZR, false
}

func (w *ItemWeapon) GetRenderObjects() []common.RenderObject {
	return []common.RenderObject{common.NewRenderObject(itemZ, w.GetShape(), w.render)}
}

func (w *ItemWeapon) SetSnapshot(tick int64, ss *protocol.ObjectSnapshot) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.tickSnapshots = append(w.tickSnapshots, &protocol.TickSnapshot{
		Tick:     tick,
		Snapshot: ss,
	})
}

func (w *ItemWeapon) GetSnapshot(tick int64) (ss *protocol.ObjectSnapshot) {
	w.lock.RLock()
	defer w.lock.RUnlock()
	for i := len(w.tickSnapshots) - 1; i >= 0; i-- {
		ts := w.tickSnapshots[i]
		if ts.Tick == tick {
			ss = ts.Snapshot
		} else if ts.Tick < tick {
			break
		}
	}
	if ss == nil {
		ss = w.getCurrentSnapshot()
	}
	return ss
}

func (w *ItemWeapon) ServerUpdate(tick int64) {
	w.SetSnapshot(tick, w.getCurrentSnapshot())
	w.cleanTickSnapshots()
	now := ticktime.GetServerTime()
	if now.Sub(w.createTime) > itemLifeTime {
		w.world.GetObjectDB().Delete(w.id)
		w.world.GetObjectDB().Delete(w.weaponID)
		w.isDestroyed = true
	}
}

func (w *ItemWeapon) ClientUpdate() {
	ss := w.getLerpSnapshot().Item.Weapon
	w.pos = ss.Pos.Convert()
	w.cleanTickSnapshots()
}

func (w *ItemWeapon) UsedBy(p common.Player) (ok bool) {
	if o, exists := w.world.GetObjectDB().SelectOne(w.weaponID); exists &&
		o.GetType() == config.WeaponObject && p.GetWeapon() == nil {
		weapon := o.(common.Weapon)
		weapon.SetPlayerID(p.GetID())
		p.SetWeapon(weapon)
		return true
	}
	return false
}

func (w *ItemWeapon) GetType() int {
	return config.ItemObject
}

func (w *ItemWeapon) cleanTickSnapshots() {
	w.lock.Lock()
	defer w.lock.Unlock()
	if len(w.tickSnapshots) <= 1 {
		return
	}
	t := ticktime.GetServerTime().Add(-config.LerpPeriod * 2)
	tick := ticktime.GetTick(t)
	index := 0
	for i, ts := range w.tickSnapshots {
		if ts.Tick >= tick {
			index = i
			break
		}
	}
	if index > 0 {
		w.tickSnapshots = w.tickSnapshots[index:]
	}
}

func (w *ItemWeapon) getCurrentSnapshot() *protocol.ObjectSnapshot {
	return &protocol.ObjectSnapshot{
		ID:   w.GetID(),
		Type: w.GetType(),
		Item: &protocol.ItemSnapshot{
			Weapon: &protocol.ItemWeaponSnapshot{
				Pos: util.ConvertVec(w.pos),
			},
		},
	}
}

func (w *ItemWeapon) render(target pixel.Target, viewPos pixel.Vec) {
	anim := animation.NewItemWeapon()
	anim.Pos = w.pos.Sub(viewPos)
	anim.Draw(target)
}

func (w *ItemWeapon) getLerpSnapshot() *protocol.ObjectSnapshot {
	return w.getSnapshotsByTime(ticktime.GetLerpTime())
}

func (w *ItemWeapon) getSnapshotsByTime(t time.Time) *protocol.ObjectSnapshot {
	w.lock.RLock()
	defer w.lock.RUnlock()
	a, b, d := protocol.GetSnapshotByTime(t, w.tickSnapshots)
	if a == nil || b == nil {
		a = w.getCurrentSnapshot()
		b = w.getCurrentSnapshot()
	}
	ssA := a.Item.Weapon
	ssB := b.Item.Weapon
	return &protocol.ObjectSnapshot{
		ID:   w.GetID(),
		Type: w.GetType(),
		Item: &protocol.ItemSnapshot{
			Weapon: &protocol.ItemWeaponSnapshot{
				Pos: util.ConvertVec(pixel.Lerp(ssA.Pos.Convert(), ssB.Pos.Convert(), d)),
			},
		},
	}
}

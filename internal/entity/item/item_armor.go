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
	itemArmorShape  = pixel.R(0, 0, 38, 40)
	itemArmorSize   = 30.
	itemArmorXLSize = 100.
)

type ItemArmor struct {
	id            string
	armor         float64
	world         common.World
	pos           pixel.Vec
	createTime    time.Time
	isDestroyed   bool
	tickSnapshots []*protocol.TickSnapshot
	lock          sync.RWMutex
}

func NewItemArmor(world common.World, id string, armor float64) *ItemArmor {
	return &ItemArmor{
		id:         id,
		world:      world,
		armor:      armor,
		pos:        util.GetHighVec(),
		createTime: ticktime.GetServerTime(),
	}
}

func (w *ItemArmor) GetID() string {
	return w.id
}

func (w *ItemArmor) Destroy() {
	w.isDestroyed = true
}

func (w *ItemArmor) Exists() bool {
	return !w.isDestroyed
}

func (w *ItemArmor) SetPos(pos pixel.Vec) {
	w.pos = pos
}

func (w *ItemArmor) GetShape() pixel.Rect {
	return itemArmorShape.Moved(w.pos.Sub(pixel.V(itemArmorShape.W()/2, 0)))
}

func (w *ItemArmor) GetCollider() (pixel.Rect, bool) {
	return pixel.ZR, false
}

func (w *ItemArmor) GetRenderObjects() []common.RenderObject {
	return []common.RenderObject{common.NewRenderObject(itemZ, w.GetShape(), w.render)}
}

func (w *ItemArmor) SetSnapshot(tick int64, ss *protocol.ObjectSnapshot) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.tickSnapshots = append(w.tickSnapshots, &protocol.TickSnapshot{
		Tick:     tick,
		Snapshot: ss,
	})
}

func (w *ItemArmor) GetSnapshot(tick int64) (ss *protocol.ObjectSnapshot) {
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

func (w *ItemArmor) ServerUpdate(tick int64) {
	w.SetSnapshot(tick, w.getCurrentSnapshot())
	w.cleanTickSnapshots()
	now := ticktime.GetServerTime()
	if now.Sub(w.createTime) > itemLifeTime {
		w.world.GetObjectDB().Delete(w.id)
		w.isDestroyed = true
	}
}

func (w *ItemArmor) ClientUpdate() {
	ss := w.getLerpSnapshot().Item.Armor
	w.pos = ss.Pos.Convert()
	w.armor = ss.Armor
	w.cleanTickSnapshots()
}

func (w *ItemArmor) UsedBy(p common.Player) (ok bool) {
	return p.AddArmorHP(w.armor, 0)
}

func (w *ItemArmor) GetType() int {
	return config.ItemObject
}

func (w *ItemArmor) cleanTickSnapshots() {
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

func (w *ItemArmor) getCurrentSnapshot() *protocol.ObjectSnapshot {
	return &protocol.ObjectSnapshot{
		ID:   w.GetID(),
		Type: w.GetType(),
		Item: &protocol.ItemSnapshot{
			Armor: &protocol.ItemArmorSnapshot{
				Pos:   util.ConvertVec(w.pos),
				Armor: w.armor,
			},
		},
	}
}

func (w *ItemArmor) render(target pixel.Target, viewPos pixel.Vec) {
	var anim *animation.Item
	if w.armor >= itemArmorXLSize {
		anim = animation.NewItemArmorBlue()
	} else {
		anim = animation.NewItemArmor()
	}
	anim.Pos = w.pos.Sub(viewPos)
	anim.Draw(target)
}

func (w *ItemArmor) getLerpSnapshot() *protocol.ObjectSnapshot {
	return w.getSnapshotsByTime(ticktime.GetLerpTime())
}

func (w *ItemArmor) getSnapshotsByTime(t time.Time) *protocol.ObjectSnapshot {
	w.lock.RLock()
	defer w.lock.RUnlock()
	a, b, d := protocol.GetSnapshotByTime(t, w.tickSnapshots)
	if a == nil || b == nil {
		a = w.getCurrentSnapshot()
		b = w.getCurrentSnapshot()
	}
	ssA := a.Item.Armor
	ssB := b.Item.Armor
	return &protocol.ObjectSnapshot{
		ID:   w.GetID(),
		Type: w.GetType(),
		Item: &protocol.ItemSnapshot{
			Armor: &protocol.ItemArmorSnapshot{
				Pos:   util.ConvertVec(pixel.Lerp(ssA.Pos.Convert(), ssB.Pos.Convert(), d)),
				Armor: ssB.Armor,
			},
		},
	}
}

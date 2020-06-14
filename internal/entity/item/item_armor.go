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
	itemArmorShape    = pixel.R(0, 0, 38, 40)
	itemArmorBlueSize = 100.
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

func (o *ItemArmor) GetID() string {
	return o.id
}

func (o *ItemArmor) Destroy() {
	o.isDestroyed = true
}

func (o *ItemArmor) Exists() bool {
	return !o.isDestroyed
}

func (o *ItemArmor) SetPos(pos pixel.Vec) {
	o.pos = pos
}

func (o *ItemArmor) GetShape() pixel.Rect {
	return itemArmorShape.Moved(o.pos.Sub(pixel.V(itemArmorShape.W()/2, 0)))
}

func (o *ItemArmor) GetCollider() (pixel.Rect, bool) {
	return pixel.ZR, false
}

func (o *ItemArmor) GetRenderObjects() []common.RenderObject {
	return []common.RenderObject{common.NewRenderObject(itemZ, o.GetShape(), o.render)}
}

func (o *ItemArmor) SetSnapshot(tick int64, ss *protocol.ObjectSnapshot) {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.tickSnapshots = append(o.tickSnapshots, &protocol.TickSnapshot{
		Tick:     tick,
		Snapshot: ss,
	})
}

func (o *ItemArmor) GetSnapshot(tick int64) (ss *protocol.ObjectSnapshot) {
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

func (o *ItemArmor) ServerUpdate(tick int64) {
	o.SetSnapshot(tick, o.getCurrentSnapshot())
	o.cleanTickSnapshots()
	now := ticktime.GetServerTime()
	if now.Sub(o.createTime) > itemLifeTime {
		o.world.GetObjectDB().Delete(o.id)
		o.isDestroyed = true
	}
}

func (o *ItemArmor) ClientUpdate() {
	ss := o.getLerpSnapshot().Item.Armor
	o.pos = ss.Pos.Convert()
	o.armor = ss.Armor
	o.cleanTickSnapshots()
}

func (o *ItemArmor) UsedBy(p common.Player) (ok bool) {
	o.world.GetObjectDB().Delete(o.GetID())
	return p.AddArmorHP(o.armor, 0)
}

func (o *ItemArmor) CollectedBy(p common.Player, index int) (ok bool) {
	return false
}

func (o *ItemArmor) GetItemType() int {
	return config.InstanceUsedItem
}

func (o *ItemArmor) GetType() int {
	return config.ItemObject
}

func (o *ItemArmor) cleanTickSnapshots() {
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

func (o *ItemArmor) getCurrentSnapshot() *protocol.ObjectSnapshot {
	return &protocol.ObjectSnapshot{
		ID:   o.GetID(),
		Type: o.GetType(),
		Item: &protocol.ItemSnapshot{
			Armor: &protocol.ItemArmorSnapshot{
				Pos:   util.ConvertVec(o.pos),
				Armor: o.armor,
			},
		},
	}
}

func (o *ItemArmor) render(target pixel.Target, viewPos pixel.Vec) {
	var anim *animation.Item
	if o.armor >= itemArmorBlueSize {
		anim = animation.NewItemArmorBlue()
	} else {
		anim = animation.NewItemArmor()
	}
	anim.Pos = o.pos.Sub(viewPos)
	anim.Draw(target)
}

func (o *ItemArmor) getLerpSnapshot() *protocol.ObjectSnapshot {
	return o.getSnapshotsByTime(ticktime.GetLerpTime())
}

func (o *ItemArmor) getSnapshotsByTime(t time.Time) *protocol.ObjectSnapshot {
	o.lock.RLock()
	defer o.lock.RUnlock()
	a, b, d := protocol.GetSnapshotByTime(t, o.tickSnapshots)
	if a == nil || b == nil {
		a = o.getCurrentSnapshot()
		b = o.getCurrentSnapshot()
	}
	ssA := a.Item.Armor
	ssB := b.Item.Armor
	return &protocol.ObjectSnapshot{
		ID:   o.GetID(),
		Type: o.GetType(),
		Item: &protocol.ItemSnapshot{
			Armor: &protocol.ItemArmorSnapshot{
				Pos:   util.ConvertVec(pixel.Lerp(ssA.Pos.Convert(), ssB.Pos.Convert(), d)),
				Armor: ssB.Armor,
			},
		},
	}
}

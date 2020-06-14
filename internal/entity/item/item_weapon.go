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

func (o *ItemWeapon) GetID() string {
	return o.id
}

func (o *ItemWeapon) Destroy() {
	o.isDestroyed = true
}

func (o *ItemWeapon) Exists() bool {
	return !o.isDestroyed
}

func (o *ItemWeapon) SetPos(pos pixel.Vec) {
	o.pos = pos
}

func (o *ItemWeapon) GetShape() pixel.Rect {
	return itemWeaponShape.Moved(o.pos.Sub(pixel.V(itemWeaponShape.W()/2, 0)))
}

func (o *ItemWeapon) GetCollider() (pixel.Rect, bool) {
	return pixel.ZR, false
}

func (o *ItemWeapon) GetRenderObjects() []common.RenderObject {
	return []common.RenderObject{common.NewRenderObject(itemZ, o.GetShape(), o.render)}
}

func (o *ItemWeapon) SetSnapshot(tick int64, ss *protocol.ObjectSnapshot) {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.tickSnapshots = append(o.tickSnapshots, &protocol.TickSnapshot{
		Tick:     tick,
		Snapshot: ss,
	})
}

func (o *ItemWeapon) GetSnapshot(tick int64) (ss *protocol.ObjectSnapshot) {
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

func (o *ItemWeapon) ServerUpdate(tick int64) {
	o.SetSnapshot(tick, o.getCurrentSnapshot())
	o.cleanTickSnapshots()
	now := ticktime.GetServerTime()
	if now.Sub(o.createTime) > itemLifeTime {
		o.world.GetObjectDB().Delete(o.id)
		o.world.GetObjectDB().Delete(o.weaponID)
		o.isDestroyed = true
	}
}

func (o *ItemWeapon) ClientUpdate() {
	ss := o.getLerpSnapshot().Item.Weapon
	o.pos = ss.Pos.Convert()
	o.weaponID = ss.WeaponID
	o.cleanTickSnapshots()
}

func (o *ItemWeapon) UsedBy(p common.Player) (ok bool) {
	if obj, exists := o.world.GetObjectDB().SelectOne(o.weaponID); exists &&
		obj.GetType() == config.WeaponObject && p.GetWeapon() == nil {
		weapon := obj.(common.Weapon)
		weapon.SetPlayerID(p.GetID())
		p.SetWeapon(weapon)
		o.world.GetObjectDB().Delete(o.GetID())
		return true
	}
	return false
}

func (o *ItemWeapon) CollectedBy(p common.Player, index int) (ok bool) {
	return false
}

func (o *ItemWeapon) GetItemType() int {
	return config.InstanceUsedItem
}

func (o *ItemWeapon) GetType() int {
	return config.ItemObject
}

func (o *ItemWeapon) cleanTickSnapshots() {
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

func (o *ItemWeapon) getCurrentSnapshot() *protocol.ObjectSnapshot {
	return &protocol.ObjectSnapshot{
		ID:   o.GetID(),
		Type: o.GetType(),
		Item: &protocol.ItemSnapshot{
			Weapon: &protocol.ItemWeaponSnapshot{
				WeaponID: o.weaponID,
				Pos:      util.ConvertVec(o.pos),
			},
		},
	}
}

func (o *ItemWeapon) render(target pixel.Target, viewPos pixel.Vec) {
	anim := animation.NewItemWeapon()
	anim.Pos = o.pos.Sub(viewPos)
	anim.Draw(target)
}

func (o *ItemWeapon) getLerpSnapshot() *protocol.ObjectSnapshot {
	return o.getSnapshotsByTime(ticktime.GetLerpTime())
}

func (o *ItemWeapon) getSnapshotsByTime(t time.Time) *protocol.ObjectSnapshot {
	o.lock.RLock()
	defer o.lock.RUnlock()
	a, b, d := protocol.GetSnapshotByTime(t, o.tickSnapshots)
	if a == nil || b == nil {
		a = o.getCurrentSnapshot()
		b = o.getCurrentSnapshot()
	}
	ssA := a.Item.Weapon
	ssB := b.Item.Weapon
	return &protocol.ObjectSnapshot{
		ID:   o.GetID(),
		Type: o.GetType(),
		Item: &protocol.ItemSnapshot{
			Weapon: &protocol.ItemWeaponSnapshot{
				WeaponID: ssB.WeaponID,
				Pos:      util.ConvertVec(pixel.Lerp(ssA.Pos.Convert(), ssB.Pos.Convert(), d)),
			},
		},
	}
}

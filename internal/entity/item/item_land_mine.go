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
	itemLandMineShape    = pixel.R(0, 0, 45, 47)
	itemLandMineCollider = pixel.R(0, 0, 40, 16)
)

const (
	itemLandMineDropRange = 72
	itemLandMineDamage    = 120.0
	itemLandMineRadius    = 300.0
)

type ItemLandMine struct {
	id            string
	playerID      string
	world         common.World
	pos           pixel.Vec
	createTime    time.Time
	deleteTime    time.Time
	slotIndex     int
	isVisible     bool
	isAcitve      bool
	isDestroyed   bool
	tickSnapshots []*protocol.TickSnapshot
	lock          sync.RWMutex
}

func NewItemLandMine(world common.World, id string) *ItemLandMine {
	return &ItemLandMine{
		id:         id,
		world:      world,
		pos:        util.GetHighVec(),
		createTime: ticktime.GetServerTime(),
	}
}

func (o *ItemLandMine) GetID() string {
	return o.id
}

func (o *ItemLandMine) Destroy() {
	o.isDestroyed = true
}

func (o *ItemLandMine) Exists() bool {
	return !o.isDestroyed
}

// set pos and reset
func (o *ItemLandMine) SetPos(pos pixel.Vec) {
	o.playerID = ""
	o.createTime = ticktime.GetServerTime()
	o.pos = pos
}

func (o *ItemLandMine) GetShape() pixel.Rect {
	return itemLandMineShape.Moved(o.pos.Sub(pixel.V(itemLandMineShape.W()/2, 0)))
}

func (o *ItemLandMine) GetCollider() (pixel.Rect, bool) {
	return itemLandMineCollider.Moved(o.pos.Sub(pixel.V(itemLandMineCollider.W()/2, 0))), false
}

func (o *ItemLandMine) GetRenderObjects() []common.RenderObject {
	return []common.RenderObject{common.NewRenderObject(itemZ, o.GetShape(), o.render)}
}

func (o *ItemLandMine) SetSnapshot(tick int64, ss *protocol.ObjectSnapshot) {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.tickSnapshots = append(o.tickSnapshots, &protocol.TickSnapshot{
		Tick:     tick,
		Snapshot: ss,
	})
}

func (o *ItemLandMine) GetSnapshot(tick int64) (ss *protocol.ObjectSnapshot) {
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

func (o *ItemLandMine) ServerUpdate(tick int64) {
	if !o.isDestroyed && o.isAcitve {
		isTriggered := false
		players := []common.Player{}
		playerDamages := []float64{}
		for _, obj := range o.world.GetObjectDB().SelectAll() {
			if obj.Exists() && obj.GetType() == config.PlayerObject {
				if player := obj.(common.Player); player.IsAlive() {
					col, _ := o.GetCollider()
					playerCol, _ := player.GetCollider()
					if col.Intersects(playerCol) {
						isTriggered = true
					}
					dist := player.GetPos().Sub(o.pos).Len()
					if dist < itemLandMineRadius {
						damage := ((itemLandMineRadius - dist) / itemLandMineRadius) * itemLandMineDamage
						playerDamages = append(playerDamages, damage)
						players = append(players, player)
					}
				}
			}
		}
		if isTriggered {
			for i, player := range players {
				player.AddDamage(o.playerID, o.id, playerDamages[i])
			}
			o.isDestroyed = true
			o.deleteTime = ticktime.GetServerTime().Add(config.LerpPeriod * 2)
		}
	}
	now := ticktime.GetServerTime()
	if (!ticktime.IsZeroTime(o.deleteTime) && now.Sub(o.deleteTime) > 0) ||
		(now.Sub(o.createTime) > itemLifeTime && o.playerID == "" && !o.isAcitve) {
		o.world.GetObjectDB().Delete(o.id)
	}
	o.SetSnapshot(tick, o.getCurrentSnapshot())
	o.cleanTickSnapshots()
}

func (o *ItemLandMine) ClientUpdate() {
	ss := o.getLerpSnapshot().Item.LandMine
	o.pos = ss.Pos.Convert()
	o.playerID = ss.PlayerID
	o.slotIndex = ss.SlotIndex
	o.isAcitve = ss.IsActive
	collider, _ := o.GetCollider()
	o.isVisible = o.world.GetScope().Intersects(collider)
	o.cleanTickSnapshots()
}

func (o *ItemLandMine) UsedBy(p common.Player) (ok bool) {
	if o.playerID == "" || o.isAcitve {
		return false
	}
	o.pos = p.GetPivot().Add(p.GetCursorDir().Unit().Scaled(itemLandMineDropRange))
	o.isAcitve = true
	return true
}

func (o *ItemLandMine) CollectedBy(p common.Player, index int) (ok bool) {
	if o.playerID != "" {
		return false
	}
	o.playerID = p.GetID()
	o.slotIndex = index
	return true
}

func (o *ItemLandMine) GetItemType() int {
	return config.CollectibleItem
}

func (o *ItemLandMine) GetType() int {
	return config.ItemObject
}

func (o *ItemLandMine) cleanTickSnapshots() {
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

func (o *ItemLandMine) getCurrentSnapshot() *protocol.ObjectSnapshot {
	return &protocol.ObjectSnapshot{
		ID:   o.GetID(),
		Type: o.GetType(),
		Item: &protocol.ItemSnapshot{
			LandMine: &protocol.ItemLandMineSnapshot{
				Pos:       util.ConvertVec(o.pos),
				PlayerID:  o.playerID,
				SlotIndex: o.slotIndex,
				IsActive:  o.isAcitve,
			},
		},
	}
}

func (o *ItemLandMine) render(target pixel.Target, viewPos pixel.Vec) {
	if o.isDestroyed {
		return
	}
	if o.isAcitve {
		if o.isVisible {
			anim := animation.NewItemLandMine()
			anim.Pos = o.pos.Sub(viewPos)
			anim.Draw(target)
		}
	} else if o.playerID == "" {
		anim := animation.NewItemMystery()
		anim.Pos = o.pos.Sub(viewPos)
		anim.Draw(target)
	} else {
		anim := animation.NewIconLandMine()
		anim.Pos = o.pos.Sub(viewPos)
		anim.Draw(target)
	}
}

func (o *ItemLandMine) getLerpSnapshot() *protocol.ObjectSnapshot {
	return o.getSnapshotsByTime(ticktime.GetLerpTime())
}

func (o *ItemLandMine) getSnapshotsByTime(t time.Time) *protocol.ObjectSnapshot {
	o.lock.RLock()
	defer o.lock.RUnlock()
	_, b, _ := protocol.GetSnapshotByTime(t, o.tickSnapshots)
	if b == nil {
		b = o.getCurrentSnapshot()
	}
	ssB := b.Item.LandMine
	return &protocol.ObjectSnapshot{
		ID:   o.GetID(),
		Type: o.GetType(),
		Item: &protocol.ItemSnapshot{
			LandMine: &protocol.ItemLandMineSnapshot{
				Pos:       ssB.Pos,
				PlayerID:  ssB.PlayerID,
				SlotIndex: ssB.SlotIndex,
				IsActive:  ssB.IsActive,
			},
		},
	}
}

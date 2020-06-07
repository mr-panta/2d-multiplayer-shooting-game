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
	itemSkullShape    = pixel.R(0, 0, 38, 43)
	itemSkullInitTime = 99 * time.Second
)

type itemSkullRecord struct {
	remainingTime time.Duration
	pickupTime    time.Time
	dropTime      time.Time
}

type ItemSkull struct {
	id            string
	world         common.World
	pos           pixel.Vec
	playerID      string
	recordMap     map[string]*itemSkullRecord
	recordLock    sync.RWMutex
	tickSnapshots []*protocol.TickSnapshot
	lock          sync.RWMutex
	// render
	remainingTime time.Duration
}

func NewItemSkull(world common.World, id string) *ItemSkull {
	return &ItemSkull{
		id:        id,
		world:     world,
		pos:       util.GetHighVec(),
		recordMap: make(map[string]*itemSkullRecord),
	}
}

func (o *ItemSkull) GetID() string {
	return o.id
}

func (o *ItemSkull) Destroy() {
	// NOOP
}

func (o *ItemSkull) Exists() bool {
	return true
}

func (o *ItemSkull) SetPos(pos pixel.Vec) {
	o.pos = pos
}

func (o *ItemSkull) GetShape() pixel.Rect {
	return itemSkullShape.Moved(o.pos.Sub(pixel.V(itemSkullShape.W()/2, 0)))
}

func (o *ItemSkull) GetCollider() (pixel.Rect, bool) {
	return pixel.ZR, false
}

func (o *ItemSkull) GetRenderObjects() []common.RenderObject {
	return []common.RenderObject{common.NewRenderObject(itemZ, o.GetShape(), o.render)}
}

func (o *ItemSkull) SetSnapshot(tick int64, ss *protocol.ObjectSnapshot) {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.tickSnapshots = append(o.tickSnapshots, &protocol.TickSnapshot{
		Tick:     tick,
		Snapshot: ss,
	})
}

func (o *ItemSkull) GetSnapshot(tick int64) (ss *protocol.ObjectSnapshot) {
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

func (o *ItemSkull) ServerUpdate(tick int64) {
	if player := o.getPlayer(); player != nil {
		if player.IsAlive() {
			o.pos = player.GetPos()
		} else {
			if record, exists := o.getRecord(o.playerID); exists {
				now := ticktime.GetServerTime()
				record.dropTime = now
				record.remainingTime -= now.Sub(record.pickupTime)
			}
			player.SetVisibleCause(o.GetID(), false)
			o.playerID = ""
		}
	}
	o.SetSnapshot(tick, o.getCurrentSnapshot())
	o.cleanTickSnapshots()
}

func (o *ItemSkull) ClientUpdate() {
	now := ticktime.GetServerTime()
	ss := o.getLastSnapshot().Item.Skull
	o.playerID = ss.PlayerID
	recordMap := make(map[string]*itemSkullRecord)
	for playerID, record := range ss.RecordMap {
		recordMap[playerID] = &itemSkullRecord{
			remainingTime: time.Duration(record.RemainingMS) * time.Millisecond,
			dropTime:      time.Unix(0, record.DropTime),
			pickupTime:    time.Unix(0, record.PickupTime),
		}
	}
	if record, exists := recordMap[o.playerID]; o.playerID != "" && exists {
		o.remainingTime = record.remainingTime - now.Sub(record.pickupTime)
	}
	o.recordMap = recordMap
	o.pos = ss.Pos.Convert()
	o.cleanTickSnapshots()
}

func (o *ItemSkull) UsedBy(player common.Player) (ok bool) {
	if player := o.getPlayer(); player != nil {
		return false
	}
	o.playerID = player.GetID()
	record, exists := o.getRecord(o.playerID)
	if !exists {
		record = &itemSkullRecord{
			remainingTime: itemSkullInitTime,
		}
	}
	record.pickupTime = ticktime.GetServerTime()
	o.setRecord(o.playerID, record)
	player.SetVisibleCause(o.GetID(), true)
	return true
}

func (o *ItemSkull) setRecord(playerID string, record *itemSkullRecord) {
	o.recordLock.Lock()
	defer o.recordLock.Unlock()
	o.recordMap[playerID] = record
}

func (o *ItemSkull) getRecord(playerID string) (record *itemSkullRecord, exists bool) {
	o.recordLock.RLock()
	defer o.recordLock.RUnlock()
	record, exists = o.recordMap[playerID]
	return record, exists
}

func (o *ItemSkull) getPlayer() common.Player {
	if o.playerID == "" {
		return nil
	}
	obj, exists := o.world.GetObjectDB().SelectOne(o.playerID)
	if !exists {
		return nil
	}
	return obj.(common.Player)
}

func (o *ItemSkull) GetType() int {
	return config.ItemObject
}

func (o *ItemSkull) cleanTickSnapshots() {
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

func (o *ItemSkull) getCurrentSnapshot() *protocol.ObjectSnapshot {
	recordMap := make(map[string]*protocol.ItemSkullRecord)
	for playerID, record := range o.recordMap {
		recordMap[playerID] = &protocol.ItemSkullRecord{
			RemainingMS: int(record.remainingTime.Seconds() * 1000),
			PickupTime:  record.pickupTime.UnixNano(),
			DropTime:    record.dropTime.UnixNano(),
		}
	}
	return &protocol.ObjectSnapshot{
		ID:   o.GetID(),
		Type: o.GetType(),
		Item: &protocol.ItemSnapshot{
			Skull: &protocol.ItemSkullSnapshot{
				Pos:       util.ConvertVec(o.pos),
				PlayerID:  o.playerID,
				RecordMap: recordMap,
			},
		},
	}
}

func (o *ItemSkull) render(target pixel.Target, viewPos pixel.Vec) {
	if player := o.getPlayer(); player != nil && player.IsAlive() {
		anim := animation.NewIconSkull()
		anim.Pos = player.GetPos().Add(pixel.V(0, 180)).Sub(viewPos)
		anim.Draw(target)
	} else {
		anim := animation.NewItemSkull()
		anim.Pos = o.pos.Sub(viewPos)
		anim.Draw(target)
	}
}

func (o *ItemSkull) getLastSnapshot() *protocol.ObjectSnapshot {
	o.lock.RLock()
	defer o.lock.RUnlock()
	if len(o.tickSnapshots) > 0 {
		return o.tickSnapshots[len(o.tickSnapshots)-1].Snapshot
	}
	return o.getCurrentSnapshot()
}

func (o *ItemSkull) getLerpSnapshot() *protocol.ObjectSnapshot {
	return o.getSnapshotsByTime(ticktime.GetLerpTime())
}

func (o *ItemSkull) getSnapshotsByTime(t time.Time) *protocol.ObjectSnapshot {
	o.lock.RLock()
	defer o.lock.RUnlock()
	a, b, d := protocol.GetSnapshotByTime(t, o.tickSnapshots)
	if a == nil || b == nil {
		a = o.getCurrentSnapshot()
		b = o.getCurrentSnapshot()
	}
	ssA := a.Item.Skull
	ssB := b.Item.Skull
	return &protocol.ObjectSnapshot{
		ID:   o.GetID(),
		Type: o.GetType(),
		Item: &protocol.ItemSnapshot{
			Skull: &protocol.ItemSkullSnapshot{
				Pos:       util.ConvertVec(pixel.Lerp(ssA.Pos.Convert(), ssB.Pos.Convert(), d)),
				PlayerID:  ssB.PlayerID,
				RecordMap: ssB.RecordMap,
			},
		},
	}
}

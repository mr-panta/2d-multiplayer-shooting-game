package item

import (
	"fmt"
	"image/color"
	"math"
	"sync"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/text"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/animation"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/common"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/config"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/protocol"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/ticktime"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/util"
	"golang.org/x/image/colornames"
)

const (
	itemSkullBlinkDiv = 3
)

var (
	itemSkullShape                = pixel.R(0, 0, 38, 43)
	itemSkullIconOffset           = pixel.V(0, 180)
	itemSkullIconOutScreenOffsets = []pixel.Vec{
		pixel.V(-itemSkullShape.W()/2, 0),
		pixel.V(0, itemSkullShape.H()/2),
		pixel.V(itemSkullShape.W()/2, 0),
		pixel.V(0, -itemSkullShape.H()/2),
	}
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
	winnerTxt     *text.Text
}

func NewItemSkull(world common.World, id string) *ItemSkull {
	return &ItemSkull{
		id:        id,
		world:     world,
		pos:       util.GetHighVec(),
		recordMap: make(map[string]*itemSkullRecord),
		winnerTxt: animation.NewText(),
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

func (o *ItemSkull) GetRenderObjects() (objs []common.RenderObject) {
	player := o.world.GetMainPlayer()
	if player == nil {
		return nil
	}
	p := player.GetPivot()
	shape := pixel.Rect{Min: p, Max: p}
	objs = append(objs, common.NewRenderObject(itemZ+1, shape, o.renderIcon))
	if player := o.getPlayer(""); !(player != nil && player.IsAlive()) {
		objs = append(objs, common.NewRenderObject(itemZ, o.GetShape(), o.render))
	}
	if o.remainingTime <= 0 {
		objs = append(objs, common.NewRenderObject(config.MinWindowRenderZ+1, shape, o.renderWinner))
	}
	return objs
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
	if player := o.getPlayer(""); player != nil {
		if player.IsAlive() {
			if record, exists := o.getRecord(o.playerID); exists {
				now := ticktime.GetServerTime()
				if remainingTime := record.remainingTime - now.Sub(record.pickupTime); remainingTime <= 0 {
					o.world.Destroy()
				}
			}
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
	oldPlayerID := o.playerID
	o.playerID = ss.PlayerID
	recordMap := make(map[string]*itemSkullRecord)
	for playerID, record := range ss.RecordMap {
		recordMap[playerID] = &itemSkullRecord{
			remainingTime: time.Duration(record.RemainingMS) * time.Millisecond,
			dropTime:      time.Unix(0, record.DropTime),
			pickupTime:    time.Unix(0, record.PickupTime),
		}
	}
	if player := o.getPlayer(""); player != nil {
		if record, exists := recordMap[o.playerID]; exists {
			o.remainingTime = record.remainingTime - now.Sub(record.pickupTime)
			if t := o.remainingTime; t > 0 {
				subfix := fmt.Sprint(int(math.Ceil(t.Seconds())))
				player.SetPlayerSubfix(fmt.Sprint(subfix))
			} else {
				player.SetPlayerSubfix("")
			}
		}
	}
	if oldPlayerID != o.playerID && oldPlayerID != "" {
		// Remove subfix
		if oldPlayer := o.getPlayer(oldPlayerID); oldPlayer != nil {
			oldPlayer.SetPlayerSubfix("")
		}
	}
	o.recordMap = recordMap
	o.pos = ss.Pos.Convert()
	o.cleanTickSnapshots()
}

func (o *ItemSkull) UsedBy(player common.Player) (ok bool) {
	if player := o.getPlayer(""); player != nil {
		return false
	}
	o.playerID = player.GetID()
	record, exists := o.getRecord(o.playerID)
	if !exists {
		record = &itemSkullRecord{
			remainingTime: config.DefaultWorldInitTime,
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

func (o *ItemSkull) getPlayer(playerID string) common.Player {
	if playerID == "" {
		playerID = o.playerID
	}
	if playerID == "" {
		return nil
	}
	obj, exists := o.world.GetObjectDB().SelectOne(playerID)
	if !exists {
		return nil
	}
	return obj.(common.Player)
}

func (o *ItemSkull) GetType() int {
	return config.ItemObject
}

func (o *ItemSkull) GetRemainingTimeMap() map[string]time.Duration {
	o.recordLock.RLock()
	defer o.recordLock.RUnlock()
	now := ticktime.GetServerTime()
	remainingTimeMap := make(map[string]time.Duration)
	for playerID, record := range o.recordMap {
		remainingTime := record.remainingTime
		if record.pickupTime.After(record.dropTime) {
			remainingTime -= now.Sub(record.pickupTime)
		}
		if remainingTime < 0 {
			remainingTime = 0
		}
		remainingTimeMap[playerID] = remainingTime
	}
	return remainingTimeMap
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
	o.recordLock.RLock()
	defer o.recordLock.RUnlock()
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
	anim := animation.NewItemSkull()
	anim.Pos = o.pos.Sub(viewPos)
	anim.Draw(target)
}

func (o *ItemSkull) renderIcon(target pixel.Target, viewPos pixel.Vec) {
	mainPlayer := o.world.GetMainPlayer()
	if mainPlayer == nil {
		return
	}
	winBound := o.world.GetWindow().Bounds()
	winBound = pixel.Rect{
		Min: winBound.Min.Add(pixel.V(1, 1)),
		Max: winBound.Max.Sub(pixel.V(1, 1)),
	}
	pos := o.pos
	if player := o.getPlayer(""); player != nil {
		pos = o.getPlayer("").GetPos().Add(itemSkullIconOffset)
	}
	pos = pos.Sub(viewPos)
	mpPos := mainPlayer.GetPos().Add(itemSkullIconOffset).Sub(viewPos)
	var c color.Color
	if !winBound.Contains(pos) {
		line := pixel.L(pos, mpPos)
		edges := winBound.Edges()
		for i, edge := range edges {
			if v, ok := line.Intersect(edge); ok {
				pos = v.Sub(itemSkullIconOutScreenOffsets[i])
			}
		}
		ratio := uint8((ticktime.GetServerTimeMS() / itemSkullBlinkDiv) % 256)
		c = &color.RGBA{R: ratio, G: ratio, B: ratio, A: ratio}
	} else if player := o.getPlayer(""); player == nil {
		return
	}
	anim := animation.NewIconSkull()
	anim.Pos = pos
	anim.Color = c
	anim.Draw(target)
}

func (o *ItemSkull) renderWinner(target pixel.Target, viewPos pixel.Vec) {
	win := o.world.GetWindow()
	smooth := win.Smooth()
	win.SetSmooth(false)
	defer win.SetSmooth(smooth)
	if player := o.getPlayer(""); player != nil {
		animation.DrawStrokeTextCenter(
			o.winnerTxt,
			target,
			win.Bounds().Center(),
			fmt.Sprintf("WINNER: %s", player.GetPlayerName()),
			4,
			colornames.White,
			colornames.Black,
		)
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

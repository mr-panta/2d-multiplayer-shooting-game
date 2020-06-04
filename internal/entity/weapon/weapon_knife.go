package weapon

import (
	"sync"
	"time"

	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/sound"

	"github.com/faiface/pixel/imdraw"

	"github.com/faiface/pixel"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/animation"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/common"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/config"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/protocol"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/ticktime"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/util"
)

const (
	knifeTriggerVisibleTime = time.Second
	knifeTriggerCooldown    = 250 * time.Millisecond
	knifeTriggerMinRange    = 8
	knifeTriggerMaxRange    = 32
	knifeDamage             = 20
)

var (
	knifeShape       = pixel.R(-10, -10, 10, 10)
	knifeShapeOffset = pixel.V(-40, 0)
)

type WeaponKnife struct {
	world         common.World
	id            string
	playerID      string
	radius        float64
	pos           pixel.Vec
	dir           pixel.Vec
	isHit         bool
	triggerTime   time.Time
	tickSnapshots []*protocol.TickSnapshot
	lock          sync.RWMutex
	imd           *imdraw.IMDraw
}

func NewWeaponKnife(world common.World, id string) common.Weapon {
	return &WeaponKnife{
		world: world,
		id:    id,
		pos:   util.GetHighVec(),
		imd:   imdraw.New(nil),
	}
}

func (o *WeaponKnife) GetID() string {
	return o.id
}

func (o *WeaponKnife) GetType() int {
	return config.WeaponObject
}

func (o *WeaponKnife) Destroy() {
	// NOOP
}

func (o *WeaponKnife) Exists() bool {
	return true
}

func (o *WeaponKnife) GetShape() pixel.Rect {
	offset := knifeShapeOffset.Sub(pixel.V(o.radius, 0))
	offset = offset.Rotated(pixel.ZV.Sub(o.dir).Angle())
	return knifeShape.Moved(o.pos).Moved(offset)
}

func (o *WeaponKnife) GetCollider() (pixel.Rect, bool) {
	return pixel.ZR, false
}

func (o *WeaponKnife) GetRenderObjects() []common.RenderObject {
	return nil
}

func (o *WeaponKnife) GetSnapshot(tick int64) (snapshot *protocol.ObjectSnapshot) {
	o.lock.RLock()
	defer o.lock.RUnlock()
	for i := len(o.tickSnapshots) - 1; i >= 0; i-- {
		ts := o.tickSnapshots[i]
		if ts.Tick == tick {
			snapshot = ts.Snapshot
		} else if ts.Tick < tick {
			break
		}
	}
	if snapshot == nil {
		snapshot = o.getCurrentSnapshot()
	}
	return snapshot
}

func (o *WeaponKnife) getCurrentSnapshot() *protocol.ObjectSnapshot {
	return &protocol.ObjectSnapshot{
		ID:   o.GetID(),
		Type: o.GetType(),
		Weapon: &protocol.WeaponSnapshot{
			Knife: &protocol.WeaponKnifeSnapshot{
				PlayerID:    o.playerID,
				TriggerTime: o.triggerTime.UnixNano(),
			},
		},
	}
}

func (o *WeaponKnife) SetSnapshot(tick int64, snapshot *protocol.ObjectSnapshot) {
	o.lock.Lock()
	defer o.lock.Unlock()
	o.tickSnapshots = append(o.tickSnapshots, &protocol.TickSnapshot{
		Tick:     tick,
		Snapshot: snapshot,
	})
}

func (o *WeaponKnife) checkPlayerCollision() common.Player {
	for _, obj := range o.world.GetObjectDB().SelectAll() {
		if !obj.Exists() || obj.GetID() == o.GetID() {
			continue
		}
		if obj.GetType() == config.PlayerObject {
			player := obj.(common.Player)
			if !player.IsAlive() || player.GetID() == o.playerID {
				continue
			}
			if player.GetShape().Intersects(o.GetShape()) {
				return player
			}
		}
	}
	return nil
}

func (o *WeaponKnife) ServerUpdate(tick int64) {
	o.updateRadius()
	if o.radius <= knifeTriggerMinRange {
		o.isHit = false
	} else if player := o.checkPlayerCollision(); !o.isHit && player != nil {
		player.AddDamage(o.playerID, o.GetID(), knifeDamage)
		o.isHit = true
	}
	o.SetSnapshot(tick, o.getCurrentSnapshot())
	o.cleanTickSnapshots()
}

func (o *WeaponKnife) ClientUpdate() {
	var ss *protocol.WeaponKnifeSnapshot
	if o.playerID != o.world.GetMainPlayerID() {
		snapshot := o.getLastSnapshot()
		ss = snapshot.Weapon.Knife
	} else {
		snapshot := o.getLerpSnapshot()
		ss = snapshot.Weapon.Knife
	}
	o.playerID = ss.PlayerID
	o.triggerTime = time.Unix(0, ss.TriggerTime)
	radius := o.radius
	o.updateRadius()
	if mainPlayer := o.world.GetMainPlayer(); mainPlayer != nil {
		dist := o.world.GetMainPlayer().GetPivot().Sub(o.pos).Len()
		if radius < o.radius && radius == knifeTriggerMinRange {
			sound.PlayWeaponKnifeStab(dist)
		}
	}
	o.cleanTickSnapshots()
}

func (o *WeaponKnife) getLastSnapshot() *protocol.ObjectSnapshot {
	o.lock.RLock()
	defer o.lock.RUnlock()
	if len(o.tickSnapshots) > 0 {
		return o.tickSnapshots[len(o.tickSnapshots)-1].Snapshot
	}
	return o.getCurrentSnapshot()
}

func (o *WeaponKnife) getLerpSnapshot() *protocol.ObjectSnapshot {
	return o.getSnapshotsByTime(ticktime.GetLerpTime())
}

func (o *WeaponKnife) getSnapshotsByTime(t time.Time) *protocol.ObjectSnapshot {
	o.lock.RLock()
	defer o.lock.RUnlock()
	_, b, _ := protocol.GetSnapshotByTime(t, o.tickSnapshots)
	if b == nil {
		b = o.getCurrentSnapshot()
	}
	ssB := b.Weapon.Knife
	return &protocol.ObjectSnapshot{
		ID:   o.GetID(),
		Type: o.GetType(),
		Weapon: &protocol.WeaponSnapshot{
			Knife: &protocol.WeaponKnifeSnapshot{
				PlayerID:    ssB.PlayerID,
				TriggerTime: ssB.TriggerTime,
			},
		},
	}
}

func (o *WeaponKnife) updateRadius() {
	radius := 0.0
	now := ticktime.GetServerTime()
	rng := float64(knifeTriggerMaxRange - knifeTriggerMinRange)
	diff := float64(now.Sub(o.triggerTime))
	cooldown := float64(knifeTriggerCooldown)
	if diff <= cooldown/2 {
		radius = (diff / (cooldown / 2)) * rng
	} else if diff <= cooldown {
		d := diff - cooldown/2
		radius = (1.0 - d/(cooldown/2)) * rng
	}
	o.radius = radius + knifeTriggerMinRange
}

func (o *WeaponKnife) cleanTickSnapshots() {
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

func (o *WeaponKnife) GetWeaponType() int {
	return config.KnifeWeapon
}

func (o *WeaponKnife) SetPos(pos pixel.Vec) {
	o.pos = pos
}

func (o *WeaponKnife) SetDir(dir pixel.Vec) {
	o.dir = dir
}

func (o *WeaponKnife) SetPlayerID(playerID string) {
	o.playerID = playerID
}

func (o *WeaponKnife) Render(target pixel.Target, viewPos pixel.Vec) {
	anim := animation.NewWeaponKnife()
	anim.Radius = o.radius
	anim.Dir = o.dir
	anim.Pos = o.pos.Sub(viewPos)
	anim.Draw(target)
	if config.EnvDebug() {
		o.renderShape(target, viewPos)
	}
}

func (o *WeaponKnife) renderShape(target pixel.Target, viewPos pixel.Vec) { // for debugging
	r := o.GetShape().Moved(pixel.ZV.Sub(viewPos))
	o.imd.Clear()
	o.imd.Color = config.ShapeColor
	o.imd.Push(r.Min, r.Max)
	o.imd.Rectangle(1)
	o.imd.Draw(target)
}

func (o *WeaponKnife) AddAmmo(ammo int) (canAdd bool) {
	return false
}

func (o *WeaponKnife) GetAmmo() (mag, ammo int) {
	return 0, 0
}

func (o *WeaponKnife) Trigger() bool {
	now := ticktime.GetServerTime()
	if now.Sub(o.triggerTime) > knifeTriggerCooldown {
		o.triggerTime = now
		return true
	}
	return false
}

func (o *WeaponKnife) Reload() bool {
	return false
}

func (o *WeaponKnife) StopReloading() {
	// NOOP
}

func (o *WeaponKnife) GetScopeRadius(dist float64) float64 {
	return 0
}

func (o *WeaponKnife) GetTriggerVisibleTime() time.Duration {
	return knifeTriggerVisibleTime
}

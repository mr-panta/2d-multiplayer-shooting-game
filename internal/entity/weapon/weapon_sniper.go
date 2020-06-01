package weapon

import (
	"sync"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/animation"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/common"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/config"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/protocol"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/sound"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/ticktime"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/util"
	"golang.org/x/image/colornames"
)

const (
	sniperWidth              = 196
	sniperBulletSpeed        = 4000
	sniperMaxRange           = 4000
	sniperBulletLength       = 12
	sniperDamage             = 100
	sniperTriggerVisibleTime = 2 * time.Second
	sniperTriggerCooldown    = 2000 * time.Millisecond
	sniperReloadCooldown     = 3 * time.Second
	sniperAmmo               = 10
	sniperMag                = 5
	sniperMaxScopeRadius     = 80
	sniperMaxScopeRange      = 240
)

type WeaponSniper struct {
	id            string
	playerID      string
	world         common.World
	pos           pixel.Vec
	dir           pixel.Vec
	posImd        *imdraw.IMDraw
	dirImd        *imdraw.IMDraw
	isDestroyed   bool
	isTriggering  bool
	isReloading   bool
	triggerTime   time.Time
	reloadTime    time.Time
	mag           int
	ammo          int
	tickSnapshots []*protocol.TickSnapshot
	lock          sync.RWMutex
}

func NewWeaponSniper(world common.World, id string) common.Weapon {
	return &WeaponSniper{
		id:     id,
		world:  world,
		pos:    util.GetHighVec(),
		posImd: imdraw.New(nil),
		dirImd: imdraw.New(nil),
		mag:    sniperMag,
	}
}

func (m *WeaponSniper) GetID() string {
	return m.id
}

func (m *WeaponSniper) Destroy() {
	m.isDestroyed = true
}

func (m *WeaponSniper) Exists() bool {
	return !m.isDestroyed
}

func (m *WeaponSniper) SetPos(pos pixel.Vec) {
	m.pos = pos
}

func (m *WeaponSniper) SetDir(dir pixel.Vec) {
	m.dir = dir
}

func (m *WeaponSniper) GetShape() pixel.Rect {
	min := m.pos.Sub(pixel.V(sniperWidth, sniperWidth).Scaled(0.5))
	max := m.pos.Add(pixel.V(sniperWidth, sniperWidth).Scaled(0.5))
	return pixel.Rect{Min: min, Max: max}
}

func (m *WeaponSniper) GetCollider() (pixel.Rect, bool) {
	return pixel.ZR, false
}

func (m *WeaponSniper) GetRenderObjects() []common.RenderObject {
	return nil
}

func (m *WeaponSniper) Render(target pixel.Target, viewPos pixel.Vec) {
	anim := animation.NewWeaponSniper()
	anim.Pos = m.pos.Sub(viewPos)
	anim.Dir = m.dir
	anim.TriggerTime = m.triggerTime
	anim.TriggerCooldown = sniperTriggerCooldown
	if m.isReloading {
		anim.State = animation.WeaponReloadState
	} else if m.isTriggering {
		anim.State = animation.WeaponTriggerState
	} else {
		anim.State = animation.WeaponIdleState
	}
	anim.Draw(target)
	// debug
	if config.EnvDebug() {
		m.renderDir(target, viewPos)
		m.renderPos(target, viewPos)
	}
}

func (m *WeaponSniper) GetType() int {
	return config.WeaponObject
}

func (m *WeaponSniper) renderPos(target pixel.Target, viewPos pixel.Vec) { // For debugging
	m.posImd.Clear()
	m.posImd.Color = colornames.Red
	m.posImd.Push(m.pos)
	m.posImd.Circle(2, 1)
	m.posImd.SetMatrix(pixel.IM.Moved(pixel.ZV.Sub(viewPos)))
	m.posImd.Draw(target)
}

func (m *WeaponSniper) renderDir(target pixel.Target, viewPos pixel.Vec) { // For debugging
	m.dirImd.Clear()
	m.dirImd.Color = colornames.Blue
	m.dirImd.Push(m.pos, m.pos.Add(m.dir.Unit().Scaled(80)))
	m.dirImd.Line(1)
	m.dirImd.SetMatrix(pixel.IM.Moved(pixel.ZV.Sub(viewPos)))
	m.dirImd.Draw(target)
}

func (m *WeaponSniper) ServerUpdate(tick int64) {
	if ticktime.IsZeroTime(m.triggerTime) {
		m.triggerTime = ticktime.GetServerTime()
	}
	if ticktime.IsZeroTime(m.reloadTime) {
		m.reloadTime = ticktime.GetServerTime()
	}
	if m.playerID == "" {
		m.isReloading = false
		m.reloadTime = ticktime.GetServerStartTime()
	} else {
		now := ticktime.GetServerTime()
		m.isTriggering = now.Sub(m.triggerTime) < sniperTriggerCooldown
		isReloading := now.Sub(m.reloadTime) < sniperReloadCooldown
		if !isReloading && m.isReloading {
			m.finishReloading()
		}
		m.isReloading = isReloading
	}
	// Add snapshot
	m.SetSnapshot(tick, m.getCurrentSnapshot())
	m.cleanTickSnapshots()
}

func (m *WeaponSniper) ClientUpdate() {
	var now time.Time
	var ss *protocol.WeaponSniperSnapshot
	if m.playerID != m.world.GetMainPlayerID() {
		now = ticktime.GetServerTime()
		snapshot := m.getLastSnapshot()
		ss = snapshot.Weapon.Sniper
	} else {
		now = ticktime.GetLerpTime()
		snapshot := m.getLerpSnapshot()
		ss = snapshot.Weapon.Sniper
	}
	prevTriggerTime := m.triggerTime
	prevReloadTime := m.reloadTime
	m.playerID = ss.PlayerID
	m.mag = ss.Mag
	m.ammo = ss.Ammo
	m.triggerTime = time.Unix(0, ss.TriggerTime)
	m.reloadTime = time.Unix(0, ss.ReloadTime)
	m.isTriggering = now.Sub(m.triggerTime) < sniperTriggerCooldown
	m.isReloading = now.Sub(m.reloadTime) < sniperReloadCooldown
	if mainPlayer := m.world.GetMainPlayer(); mainPlayer != nil {
		dist := m.world.GetMainPlayer().GetPivot().Sub(m.pos).Len()
		if !ticktime.IsZeroTime(prevTriggerTime) && prevTriggerTime.Before(m.triggerTime) {
			sound.PlayWeaponSniperFire(dist)
		}
		if !ticktime.IsZeroTime(prevReloadTime) && prevReloadTime.Before(m.reloadTime) {
			sound.PlayWeaponSniperReload(dist)
		}
	}
	// Clean snapshot
	m.cleanTickSnapshots()
}

func (m *WeaponSniper) GetSnapshot(tick int64) (snapshot *protocol.ObjectSnapshot) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	for i := len(m.tickSnapshots) - 1; i >= 0; i-- {
		ts := m.tickSnapshots[i]
		if ts.Tick == tick {
			snapshot = ts.Snapshot
		} else if ts.Tick < tick {
			break
		}
	}
	if snapshot == nil {
		snapshot = m.getCurrentSnapshot()
	}
	return snapshot
}

func (m *WeaponSniper) SetSnapshot(tick int64, snapshot *protocol.ObjectSnapshot) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.tickSnapshots = append(m.tickSnapshots, &protocol.TickSnapshot{
		Tick:     tick,
		Snapshot: snapshot,
	})
}

func (m *WeaponSniper) Trigger() (ok bool) {
	ok = false
	if !m.isTriggering && m.mag > 0 && !m.isReloading {
		bullet := NewBullet(m.world, m.world.GetObjectDB().GetAvailableID())
		bullet.Fire(
			m.playerID,
			m.id,
			m.pos.Add(m.dir.Unit().Scaled(sniperWidth/2)),
			m.dir,
			sniperBulletSpeed,
			sniperMaxRange,
			sniperDamage,
			sniperBulletLength,
		)
		m.world.GetObjectDB().Set(bullet)
		m.triggerTime = ticktime.GetServerTime()
		m.mag--
		ok = true
	}
	if !m.isReloading && m.mag == 0 {
		_ = m.Reload()
	}
	return ok
}

func (m *WeaponSniper) Reload() bool {
	if !m.isReloading && m.mag < sniperMag && m.ammo > 0 {
		m.reloadTime = ticktime.GetServerTime()
		return true
	}
	return false
}

func (m *WeaponSniper) SetPlayerID(playerID string) {
	m.playerID = playerID
}

func (m *WeaponSniper) AddAmmo(ammo int) bool {
	if m.ammo >= sniperAmmo {
		return false
	}
	if ammo == -1 {
		ammo = sniperAmmo
	} else if ammo == -2 {
		ammo = sniperMag
	}
	if m.ammo += ammo; m.ammo > sniperAmmo {
		m.ammo = sniperAmmo
	}
	return true
}
func (m *WeaponSniper) GetAmmo() (mag, ammo int) {
	return m.mag, m.ammo
}

func (m *WeaponSniper) GetScopeRadius(dist float64) float64 {
	if dist > sniperMaxScopeRange {
		return sniperMaxScopeRadius
	}
	return sniperMaxScopeRadius * (dist / sniperMaxScopeRange)
}

func (m *WeaponSniper) GetWeaponType() int {
	return config.SniperWeapon
}

func (m *WeaponSniper) GetTriggerVisibleTime() time.Duration {
	return sniperTriggerVisibleTime
}

func (m *WeaponSniper) finishReloading() {
	if m.mag < sniperMag && m.ammo > 0 {
		totalAmmo := m.ammo + m.mag
		if totalAmmo > sniperMag {
			m.mag = sniperMag
		} else {
			m.mag = totalAmmo
		}
		m.ammo = totalAmmo - m.mag
	}
}

func (m *WeaponSniper) getLastSnapshot() *protocol.ObjectSnapshot {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if len(m.tickSnapshots) > 0 {
		return m.tickSnapshots[len(m.tickSnapshots)-1].Snapshot
	}
	return m.getCurrentSnapshot()
}

func (m *WeaponSniper) getLerpSnapshot() *protocol.ObjectSnapshot {
	return m.getSnapshotsByTime(ticktime.GetLerpTime())
}

func (m *WeaponSniper) getSnapshotsByTime(t time.Time) *protocol.ObjectSnapshot {
	m.lock.RLock()
	defer m.lock.RUnlock()
	_, b, _ := protocol.GetSnapshotByTime(t, m.tickSnapshots)
	if b == nil {
		b = m.getCurrentSnapshot()
	}
	ssB := b.Weapon.Sniper
	return &protocol.ObjectSnapshot{
		ID:   m.GetID(),
		Type: m.GetType(),
		Weapon: &protocol.WeaponSnapshot{
			Sniper: &protocol.WeaponSniperSnapshot{
				PlayerID:    ssB.PlayerID,
				Mag:         ssB.Mag,
				Ammo:        ssB.Ammo,
				TriggerTime: ssB.TriggerTime,
				ReloadTime:  ssB.ReloadTime,
			},
		},
	}
}

func (m *WeaponSniper) getCurrentSnapshot() *protocol.ObjectSnapshot {
	return &protocol.ObjectSnapshot{
		ID:   m.GetID(),
		Type: m.GetType(),
		Weapon: &protocol.WeaponSnapshot{
			Sniper: &protocol.WeaponSniperSnapshot{
				PlayerID:    m.playerID,
				Mag:         m.mag,
				Ammo:        m.ammo,
				TriggerTime: m.triggerTime.UnixNano(),
				ReloadTime:  m.reloadTime.UnixNano(),
			},
		},
	}
}

func (m *WeaponSniper) cleanTickSnapshots() {
	m.lock.Lock()
	defer m.lock.Unlock()
	if len(m.tickSnapshots) <= 1 {
		return
	}
	t := ticktime.GetServerTime().Add(-config.LerpPeriod * 2)
	tick := ticktime.GetTick(t)
	index := 0
	for i, ts := range m.tickSnapshots {
		if ts.Tick >= tick {
			index = i
			break
		}
	}
	if index > 0 {
		m.tickSnapshots = m.tickSnapshots[index:]
	}
}

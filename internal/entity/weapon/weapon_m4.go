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
	m4Width           = 124
	m4BulletSpeed     = 1000
	m4MaxRange        = 1500
	m4BulletLength    = 12
	m4Damage          = 20
	m4TriggerCooldown = 200 * time.Millisecond
	m4ReloadCooldown  = 2 * time.Second
	m4Ammo            = 60
	m4Mag             = 30
	m4MaxScopeRadius  = 160
	m4MaxScopeRange   = 600
)

type WeaponM4 struct {
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

func NewWeaponM4(world common.World, id string) *WeaponM4 {
	return &WeaponM4{
		id:     id,
		world:  world,
		pos:    util.GetHighVec(),
		posImd: imdraw.New(nil),
		dirImd: imdraw.New(nil),
		mag:    m4Mag,
	}
}

func (m *WeaponM4) GetID() string {
	return m.id
}

func (m *WeaponM4) Destroy() {
	m.isDestroyed = true
}

func (m *WeaponM4) Exists() bool {
	return !m.isDestroyed
}

func (m *WeaponM4) SetPos(pos pixel.Vec) {
	m.pos = pos
}

func (m *WeaponM4) SetDir(dir pixel.Vec) {
	m.dir = dir
}

func (m *WeaponM4) GetShape() pixel.Rect {
	min := m.pos.Sub(pixel.V(m4Width, m4Width).Scaled(0.5))
	max := m.pos.Add(pixel.V(m4Width, m4Width).Scaled(0.5))
	return pixel.Rect{Min: min, Max: max}
}

func (m *WeaponM4) GetCollider() (pixel.Rect, bool) {
	return pixel.ZR, false
}

func (m *WeaponM4) GetRenderObjects() []common.RenderObject {
	return nil
}

func (m *WeaponM4) Render(target pixel.Target, viewPos pixel.Vec) {
	anim := animation.NewWeaponM4()
	anim.Pos = m.pos.Sub(viewPos)
	anim.Dir = m.dir
	if m.isReloading {
		anim.State = animation.WeaponM4ReloadState
	} else {
		anim.State = animation.WeaponM4IdleState
	}
	anim.Draw(target)
	// debug
	if config.EnvDebug() {
		m.renderDir(target, viewPos)
		m.renderPos(target, viewPos)
	}
}

func (m *WeaponM4) GetType() int {
	return config.WeaponObject
}

func (m *WeaponM4) renderPos(target pixel.Target, viewPos pixel.Vec) { // For debugging
	m.posImd.Clear()
	m.posImd.Color = colornames.Red
	m.posImd.Push(m.pos)
	m.posImd.Circle(2, 1)
	m.posImd.SetMatrix(pixel.IM.Moved(pixel.ZV.Sub(viewPos)))
	m.posImd.Draw(target)
}

func (m *WeaponM4) renderDir(target pixel.Target, viewPos pixel.Vec) { // For debugging
	m.dirImd.Clear()
	m.dirImd.Color = colornames.Blue
	m.dirImd.Push(m.pos, m.pos.Add(m.dir.Unit().Scaled(80)))
	m.dirImd.Line(1)
	m.dirImd.SetMatrix(pixel.IM.Moved(pixel.ZV.Sub(viewPos)))
	m.dirImd.Draw(target)
}

func (m *WeaponM4) ServerUpdate(tick int64) {
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
		m.isTriggering = now.Sub(m.triggerTime) < m4TriggerCooldown
		isReloading := now.Sub(m.reloadTime) < m4ReloadCooldown
		if !isReloading && m.isReloading {
			m.finishReloading()
		}
		m.isReloading = isReloading
	}
}

func (m *WeaponM4) ClientUpdate() {
	var now time.Time
	var ss *protocol.WeaponM4Snapshot
	if m.playerID != m.world.GetMainPlayerID() {
		now = ticktime.GetServerTime()
		snapshot := m.getLastSnapshot()
		ss = snapshot.Weapon.M4
	} else {
		now = ticktime.GetLerpTime()
		snapshot := m.getLerpSnapshot()
		ss = snapshot.Weapon.M4
	}
	prevTriggerTime := m.triggerTime
	prevReloadTime := m.reloadTime
	m.playerID = ss.PlayerID
	m.mag = ss.Mag
	m.ammo = ss.Ammo
	m.triggerTime = time.Unix(0, ss.TriggerTime)
	m.reloadTime = time.Unix(0, ss.ReloadTime)
	m.isTriggering = now.Sub(m.triggerTime) < m4TriggerCooldown
	m.isReloading = now.Sub(m.reloadTime) < m4ReloadCooldown
	if mainPlayer := m.world.GetMainPlayer(); mainPlayer != nil {
		dist := m.world.GetMainPlayer().GetPivot().Sub(m.pos).Len()
		if !ticktime.IsZeroTime(prevTriggerTime) && prevTriggerTime.Before(m.triggerTime) {
			sound.PlayWeaponM4Fire(dist)
		}
		if !ticktime.IsZeroTime(prevReloadTime) && prevReloadTime.Before(m.reloadTime) {
			sound.PlayWeaponM4Reload(dist)
		}
	}
}

func (m *WeaponM4) GetSnapshot(tick int64) (snapshot *protocol.ObjectSnapshot) {
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

func (m *WeaponM4) SetSnapshot(tick int64, snapshot *protocol.ObjectSnapshot) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.tickSnapshots = append(m.tickSnapshots, &protocol.TickSnapshot{
		Tick:     tick,
		Snapshot: snapshot,
	})
}

func (m *WeaponM4) Trigger() (ok bool) {
	ok = false
	if !m.isTriggering && m.mag > 0 && !m.isReloading {
		bullet := NewBullet(m.world, m.world.GetObjectDB().GetAvailableID())
		bullet.Fire(
			m.playerID,
			m.id,
			m.pos.Add(m.dir.Unit().Scaled(m4Width/2)),
			m.dir,
			m4BulletSpeed,
			m4MaxRange,
			m4Damage,
			m4BulletLength,
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

func (m *WeaponM4) Reload() bool {
	if !m.isReloading && m.mag < m4Mag && m.ammo > 0 {
		m.reloadTime = ticktime.GetServerTime()
		return true
	}
	return false
}

func (m *WeaponM4) SetPlayerID(playerID string) {
	m.playerID = playerID
}

func (m *WeaponM4) AddAmmo(ammo int) bool {
	if m.ammo >= m4Ammo {
		return false
	}
	if ammo == -1 {
		ammo = m4Ammo
	} else if ammo == -2 {
		ammo = m4Mag
	}
	if m.ammo += ammo; m.ammo > m4Ammo {
		m.ammo = m4Ammo
	}
	return true
}
func (m *WeaponM4) GetAmmo() (mag, ammo int) {
	return m.mag, m.ammo
}

func (m *WeaponM4) GetScopeRadius(dist float64) float64 {
	if dist > m4MaxScopeRange {
		return 0
	}
	return m4MaxScopeRadius * (1.0 - (dist / m4MaxScopeRange))
}

func (m *WeaponM4) GetWeaponType() int {
	return config.M4Weapon
}

func (m *WeaponM4) finishReloading() {
	if m.mag < m4Mag && m.ammo > 0 {
		totalAmmo := m.ammo + m.mag
		if totalAmmo > m4Mag {
			m.mag = m4Mag
		} else {
			m.mag = totalAmmo
		}
		m.ammo = totalAmmo - m.mag
	}
}

func (m *WeaponM4) getLastSnapshot() *protocol.ObjectSnapshot {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if len(m.tickSnapshots) > 0 {
		return m.tickSnapshots[len(m.tickSnapshots)-1].Snapshot
	}
	return m.getCurrentSnapshot()
}

func (m *WeaponM4) getLerpSnapshot() *protocol.ObjectSnapshot {
	return m.getSnapshotsByTime(ticktime.GetLerpTime())
}

func (m *WeaponM4) getSnapshotsByTime(t time.Time) *protocol.ObjectSnapshot {
	m.lock.RLock()
	defer m.lock.RUnlock()
	_, b, _ := protocol.GetSnapshotByTime(t, m.tickSnapshots)
	if b == nil {
		b = m.getCurrentSnapshot()
	}
	ssB := b.Weapon.M4
	return &protocol.ObjectSnapshot{
		ID:   m.GetID(),
		Type: m.GetType(),
		Weapon: &protocol.WeaponSnapshot{
			M4: &protocol.WeaponM4Snapshot{
				PlayerID:    ssB.PlayerID,
				Mag:         ssB.Mag,
				Ammo:        ssB.Ammo,
				TriggerTime: ssB.TriggerTime,
				ReloadTime:  ssB.ReloadTime,
			},
		},
	}
}

func (m *WeaponM4) getCurrentSnapshot() *protocol.ObjectSnapshot {
	return &protocol.ObjectSnapshot{
		ID:   m.GetID(),
		Type: m.GetType(),
		Weapon: &protocol.WeaponSnapshot{
			M4: &protocol.WeaponM4Snapshot{
				PlayerID:    m.playerID,
				Mag:         m.mag,
				Ammo:        m.ammo,
				TriggerTime: m.triggerTime.UnixNano(),
				ReloadTime:  m.reloadTime.UnixNano(),
			},
		},
	}
}

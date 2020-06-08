package weapon

import (
	"math"
	"math/rand"
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
	shotgunDropRate           = 20
	shotgunWidth              = 124
	shotgunBulletAmount       = 8
	shotgunBulletSpeed        = 1500
	shotgunMaxRange           = 450
	shotgunBulletLength       = 6
	shotgunDamage             = 10
	shotgunTriggerVisibleTime = time.Second
	shotgunTriggerCooldown    = 1000 * time.Millisecond
	shotgunReloadCooldown     = 2 * time.Second
	shotgunAmmo               = 16
	shotgunMag                = 8
	shotgunMaxScopeRadius     = 200
	shotgunMaxScopeRange      = 400
	shotgunRecoilAngle        = math.Pi / 180 * 52
)

type WeaponShotgun struct {
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

func NewWeaponShotgun(world common.World, id string) common.Weapon {
	return &WeaponShotgun{
		id:     id,
		world:  world,
		pos:    util.GetHighVec(),
		posImd: imdraw.New(nil),
		dirImd: imdraw.New(nil),
		mag:    shotgunMag,
	}
}

func (m *WeaponShotgun) GetID() string {
	return m.id
}

func (m *WeaponShotgun) Destroy() {
	m.isDestroyed = true
}

func (m *WeaponShotgun) Exists() bool {
	return !m.isDestroyed
}

func (m *WeaponShotgun) SetPos(pos pixel.Vec) {
	m.pos = pos
}

func (m *WeaponShotgun) SetDir(dir pixel.Vec) {
	m.dir = dir
}

func (m *WeaponShotgun) GetShape() pixel.Rect {
	min := m.pos.Sub(pixel.V(shotgunWidth, shotgunWidth).Scaled(0.5))
	max := m.pos.Add(pixel.V(shotgunWidth, shotgunWidth).Scaled(0.5))
	return pixel.Rect{Min: min, Max: max}
}

func (m *WeaponShotgun) GetCollider() (pixel.Rect, bool) {
	return pixel.ZR, false
}

func (m *WeaponShotgun) GetRenderObjects() []common.RenderObject {
	return nil
}

func (m *WeaponShotgun) Render(target pixel.Target, viewPos pixel.Vec) {
	now := ticktime.GetLerpTime()
	if m.playerID == m.world.GetMainPlayerID() {
		now = ticktime.GetServerTime()
	}
	anim := animation.NewWeaponShotgun()
	anim.Pos = m.pos.Sub(viewPos)
	anim.Dir = m.dir
	anim.TriggerDuration = now.Sub(m.triggerTime)
	anim.TriggerCooldown = shotgunTriggerCooldown
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

func (m *WeaponShotgun) GetType() int {
	return config.WeaponObject
}

func (m *WeaponShotgun) renderPos(target pixel.Target, viewPos pixel.Vec) { // For debugging
	m.posImd.Clear()
	m.posImd.Color = colornames.Red
	m.posImd.Push(m.pos)
	m.posImd.Circle(2, 1)
	m.posImd.SetMatrix(pixel.IM.Moved(pixel.ZV.Sub(viewPos)))
	m.posImd.Draw(target)
}

func (m *WeaponShotgun) renderDir(target pixel.Target, viewPos pixel.Vec) { // For debugging
	m.dirImd.Clear()
	m.dirImd.Color = colornames.Blue
	m.dirImd.Push(m.pos, m.pos.Add(m.dir.Unit().Scaled(80)))
	m.dirImd.Line(1)
	m.dirImd.SetMatrix(pixel.IM.Moved(pixel.ZV.Sub(viewPos)))
	m.dirImd.Draw(target)
}

func (m *WeaponShotgun) ServerUpdate(tick int64) {
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
		m.isTriggering = now.Sub(m.triggerTime) < shotgunTriggerCooldown
		isReloading := now.Sub(m.reloadTime) < shotgunReloadCooldown
		if !isReloading && m.isReloading {
			m.finishReloading()
		}
		m.isReloading = isReloading
	}
	// Add snapshot
	m.SetSnapshot(tick, m.getCurrentSnapshot())
	m.cleanTickSnapshots()
}

func (m *WeaponShotgun) ClientUpdate() {
	var now time.Time
	var ss *protocol.WeaponShotgunSnapshot
	if m.playerID == m.world.GetMainPlayerID() {
		now = ticktime.GetServerTime()
		snapshot := m.getLastSnapshot()
		ss = snapshot.Weapon.Shotgun
	} else {
		now = ticktime.GetLerpTime()
		snapshot := m.getLerpSnapshot()
		ss = snapshot.Weapon.Shotgun
	}
	prevTriggerTime := m.triggerTime
	prevReloadTime := m.reloadTime
	m.playerID = ss.PlayerID
	m.mag = ss.Mag
	m.ammo = ss.Ammo
	m.triggerTime = time.Unix(0, ss.TriggerTime)
	m.reloadTime = time.Unix(0, ss.ReloadTime)
	m.isTriggering = now.Sub(m.triggerTime) < shotgunTriggerCooldown
	m.isReloading = now.Sub(m.reloadTime) < shotgunReloadCooldown
	if mainPlayer := m.world.GetMainPlayer(); mainPlayer != nil {
		dist := m.world.GetMainPlayer().GetPivot().Sub(m.pos).Len()
		if !ticktime.IsZeroTime(prevTriggerTime) && prevTriggerTime.Before(m.triggerTime) {
			sound.PlayWeaponShotgunFire(dist)
		}
		if !ticktime.IsZeroTime(prevReloadTime) && prevReloadTime.Before(m.reloadTime) {
			sound.PlayWeaponShotgunReload(dist)
		}
	}

	// Clean snapshot
	m.cleanTickSnapshots()
}

func (m *WeaponShotgun) GetSnapshot(tick int64) (snapshot *protocol.ObjectSnapshot) {
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

func (m *WeaponShotgun) SetSnapshot(tick int64, snapshot *protocol.ObjectSnapshot) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.tickSnapshots = append(m.tickSnapshots, &protocol.TickSnapshot{
		Tick:     tick,
		Snapshot: snapshot,
	})
}

func (m *WeaponShotgun) Trigger() (ok bool) {
	ok = false
	if !m.isTriggering && m.mag > 0 && !m.isReloading {
		for i := 0; i < shotgunBulletAmount; i++ {
			bullet := NewBullet(m.world, m.world.GetObjectDB().GetAvailableID())
			recoilAngle := rand.Float64()*shotgunRecoilAngle - shotgunRecoilAngle/2
			dir := m.dir.Rotated(recoilAngle)
			bullet.Fire(
				m.playerID,
				m.id,
				m.pos.Add(m.dir.Unit().Scaled(shotgunWidth/2)),
				dir,
				shotgunBulletSpeed,
				shotgunMaxRange,
				shotgunDamage,
				shotgunBulletLength,
			)
			m.world.GetObjectDB().Set(bullet)
		}
		m.triggerTime = ticktime.GetServerTime()
		m.mag--
		ok = true
	}
	if !m.isReloading && m.mag == 0 {
		_ = m.Reload()
	}
	return ok
}

func (m *WeaponShotgun) Reload() bool {
	if !m.isReloading && m.mag < shotgunMag && m.ammo > 0 {
		m.reloadTime = ticktime.GetServerTime()
		return true
	}
	return false
}

func (m *WeaponShotgun) StopReloading() {
	m.isReloading = false
	m.reloadTime = ticktime.GetServerStartTime()
}

func (m *WeaponShotgun) SetPlayerID(playerID string) {
	m.playerID = playerID
}

func (m *WeaponShotgun) AddAmmo(ammo int) bool {
	if m.ammo >= shotgunAmmo {
		return false
	}
	if ammo == -1 {
		ammo = shotgunAmmo
	} else if ammo == -2 {
		ammo = shotgunMag
	}
	if m.ammo += ammo; m.ammo > shotgunAmmo {
		m.ammo = shotgunAmmo
	}
	return true
}
func (m *WeaponShotgun) GetAmmo() (mag, ammo int) {
	return m.mag, m.ammo
}

func (m *WeaponShotgun) GetScopeRadius(dist float64) float64 {
	if dist > shotgunMaxScopeRange {
		return 0
	}
	return shotgunMaxScopeRadius * (1.0 - (dist / shotgunMaxScopeRange))
}

func (m *WeaponShotgun) GetWeaponType() int {
	return config.ShotgunWeapon
}

func (m *WeaponShotgun) GetTriggerVisibleTime() time.Duration {
	return shotgunTriggerVisibleTime
}

func (m *WeaponShotgun) finishReloading() {
	if m.mag < shotgunMag && m.ammo > 0 {
		totalAmmo := m.ammo + m.mag
		if totalAmmo > shotgunMag {
			m.mag = shotgunMag
		} else {
			m.mag = totalAmmo
		}
		m.ammo = totalAmmo - m.mag
	}
}

func (m *WeaponShotgun) getLastSnapshot() *protocol.ObjectSnapshot {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if len(m.tickSnapshots) > 0 {
		return m.tickSnapshots[len(m.tickSnapshots)-1].Snapshot
	}
	return m.getCurrentSnapshot()
}

func (m *WeaponShotgun) getLerpSnapshot() *protocol.ObjectSnapshot {
	return m.getSnapshotsByTime(ticktime.GetLerpTime())
}

func (m *WeaponShotgun) getSnapshotsByTime(t time.Time) *protocol.ObjectSnapshot {
	m.lock.RLock()
	defer m.lock.RUnlock()
	_, b, _ := protocol.GetSnapshotByTime(t, m.tickSnapshots)
	if b == nil {
		b = m.getCurrentSnapshot()
	}
	ssB := b.Weapon.Shotgun
	return &protocol.ObjectSnapshot{
		ID:   m.GetID(),
		Type: m.GetType(),
		Weapon: &protocol.WeaponSnapshot{
			Shotgun: &protocol.WeaponShotgunSnapshot{
				PlayerID:    ssB.PlayerID,
				Mag:         ssB.Mag,
				Ammo:        ssB.Ammo,
				TriggerTime: ssB.TriggerTime,
				ReloadTime:  ssB.ReloadTime,
			},
		},
	}
}

func (m *WeaponShotgun) getCurrentSnapshot() *protocol.ObjectSnapshot {
	return &protocol.ObjectSnapshot{
		ID:   m.GetID(),
		Type: m.GetType(),
		Weapon: &protocol.WeaponSnapshot{
			Shotgun: &protocol.WeaponShotgunSnapshot{
				PlayerID:    m.playerID,
				Mag:         m.mag,
				Ammo:        m.ammo,
				TriggerTime: m.triggerTime.UnixNano(),
				ReloadTime:  m.reloadTime.UnixNano(),
			},
		},
	}
}

func (m *WeaponShotgun) cleanTickSnapshots() {
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

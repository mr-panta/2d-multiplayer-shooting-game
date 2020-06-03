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
	smgDropRate           = 20
	smgWidth              = 124
	smgBulletSpeed        = 1000
	smgMaxRange           = 1000
	smgBulletLength       = 12
	smgDamage             = 10
	smgTriggerVisibleTime = time.Second
	smgTriggerCooldown    = 100 * time.Millisecond
	smgReloadCooldown     = 2 * time.Second
	smgAmmo               = 60
	smgMag                = 30
	smgMaxScopeRadius     = 160
	smgMaxScopeRange      = 600
	smgRecoilAngle        = math.Pi / 180 * 6
)

type WeaponSMG struct {
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

func NewWeaponSMG(world common.World, id string) common.Weapon {
	return &WeaponSMG{
		id:     id,
		world:  world,
		pos:    util.GetHighVec(),
		posImd: imdraw.New(nil),
		dirImd: imdraw.New(nil),
		mag:    smgMag,
	}
}

func (m *WeaponSMG) GetID() string {
	return m.id
}

func (m *WeaponSMG) Destroy() {
	m.isDestroyed = true
}

func (m *WeaponSMG) Exists() bool {
	return !m.isDestroyed
}

func (m *WeaponSMG) SetPos(pos pixel.Vec) {
	m.pos = pos
}

func (m *WeaponSMG) SetDir(dir pixel.Vec) {
	m.dir = dir
}

func (m *WeaponSMG) GetShape() pixel.Rect {
	min := m.pos.Sub(pixel.V(smgWidth, smgWidth).Scaled(0.5))
	max := m.pos.Add(pixel.V(smgWidth, smgWidth).Scaled(0.5))
	return pixel.Rect{Min: min, Max: max}
}

func (m *WeaponSMG) GetCollider() (pixel.Rect, bool) {
	return pixel.ZR, false
}

func (m *WeaponSMG) GetRenderObjects() []common.RenderObject {
	return nil
}

func (m *WeaponSMG) Render(target pixel.Target, viewPos pixel.Vec) {
	anim := animation.NewWeaponSMG()
	anim.Pos = m.pos.Sub(viewPos)
	anim.Dir = m.dir
	anim.TriggerTime = m.triggerTime
	anim.TriggerCooldown = smgTriggerCooldown
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

func (m *WeaponSMG) GetType() int {
	return config.WeaponObject
}

func (m *WeaponSMG) renderPos(target pixel.Target, viewPos pixel.Vec) { // For debugging
	m.posImd.Clear()
	m.posImd.Color = colornames.Red
	m.posImd.Push(m.pos)
	m.posImd.Circle(2, 1)
	m.posImd.SetMatrix(pixel.IM.Moved(pixel.ZV.Sub(viewPos)))
	m.posImd.Draw(target)
}

func (m *WeaponSMG) renderDir(target pixel.Target, viewPos pixel.Vec) { // For debugging
	m.dirImd.Clear()
	m.dirImd.Color = colornames.Blue
	m.dirImd.Push(m.pos, m.pos.Add(m.dir.Unit().Scaled(80)))
	m.dirImd.Line(1)
	m.dirImd.SetMatrix(pixel.IM.Moved(pixel.ZV.Sub(viewPos)))
	m.dirImd.Draw(target)
}

func (m *WeaponSMG) ServerUpdate(tick int64) {
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
		m.isTriggering = now.Sub(m.triggerTime) < smgTriggerCooldown
		isReloading := now.Sub(m.reloadTime) < smgReloadCooldown
		if !isReloading && m.isReloading {
			m.finishReloading()
		}
		m.isReloading = isReloading
	}
	// Add snapshot
	m.SetSnapshot(tick, m.getCurrentSnapshot())
	m.cleanTickSnapshots()
}

func (m *WeaponSMG) ClientUpdate() {
	var now time.Time
	var ss *protocol.WeaponSMGSnapshot
	if m.playerID != m.world.GetMainPlayerID() {
		now = ticktime.GetServerTime()
		snapshot := m.getLastSnapshot()
		ss = snapshot.Weapon.SMG
	} else {
		now = ticktime.GetLerpTime()
		snapshot := m.getLerpSnapshot()
		ss = snapshot.Weapon.SMG
	}
	prevTriggerTime := m.triggerTime
	prevReloadTime := m.reloadTime
	m.playerID = ss.PlayerID
	m.mag = ss.Mag
	m.ammo = ss.Ammo
	m.triggerTime = time.Unix(0, ss.TriggerTime)
	m.reloadTime = time.Unix(0, ss.ReloadTime)
	m.isTriggering = now.Sub(m.triggerTime) < smgTriggerCooldown
	m.isReloading = now.Sub(m.reloadTime) < smgReloadCooldown
	if mainPlayer := m.world.GetMainPlayer(); mainPlayer != nil {
		dist := m.world.GetMainPlayer().GetPivot().Sub(m.pos).Len()
		if !ticktime.IsZeroTime(prevTriggerTime) && prevTriggerTime.Before(m.triggerTime) {
			sound.PlayWeaponSMGFire(dist)
		}
		if !ticktime.IsZeroTime(prevReloadTime) && prevReloadTime.Before(m.reloadTime) {
			sound.PlayWeaponSMGReload(dist)
		}
	}
	// Clean snapshot
	m.cleanTickSnapshots()
}

func (m *WeaponSMG) GetSnapshot(tick int64) (snapshot *protocol.ObjectSnapshot) {
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

func (m *WeaponSMG) SetSnapshot(tick int64, snapshot *protocol.ObjectSnapshot) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.tickSnapshots = append(m.tickSnapshots, &protocol.TickSnapshot{
		Tick:     tick,
		Snapshot: snapshot,
	})
}

func (m *WeaponSMG) Trigger() (ok bool) {
	ok = false
	if !m.isTriggering && m.mag > 0 && !m.isReloading {
		bullet := NewBullet(m.world, m.world.GetObjectDB().GetAvailableID())
		recoilAngle := rand.Float64()*smgRecoilAngle - smgRecoilAngle/2
		dir := m.dir.Rotated(recoilAngle)
		bullet.Fire(
			m.playerID,
			m.id,
			m.pos.Add(m.dir.Unit().Scaled(smgWidth/2)),
			dir,
			smgBulletSpeed,
			smgMaxRange,
			smgDamage,
			smgBulletLength,
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

func (m *WeaponSMG) Reload() bool {
	if !m.isReloading && m.mag < smgMag && m.ammo > 0 {
		m.reloadTime = ticktime.GetServerTime()
		return true
	}
	return false
}

func (m *WeaponSMG) StopReloading() {
	m.isReloading = false
	m.reloadTime = ticktime.GetServerStartTime()
}

func (m *WeaponSMG) SetPlayerID(playerID string) {
	m.playerID = playerID
}

func (m *WeaponSMG) AddAmmo(ammo int) bool {
	if m.ammo >= smgAmmo {
		return false
	}
	if ammo == -1 {
		ammo = smgAmmo
	} else if ammo == -2 {
		ammo = smgMag
	}
	if m.ammo += ammo; m.ammo > smgAmmo {
		m.ammo = smgAmmo
	}
	return true
}
func (m *WeaponSMG) GetAmmo() (mag, ammo int) {
	return m.mag, m.ammo
}

func (m *WeaponSMG) GetScopeRadius(dist float64) float64 {
	if dist > smgMaxScopeRange {
		return 0
	}
	return smgMaxScopeRadius * (1.0 - (dist / smgMaxScopeRange))
}

func (m *WeaponSMG) GetTriggerVisibleTime() time.Duration {
	return smgTriggerVisibleTime
}

func (m *WeaponSMG) GetWeaponType() int {
	return config.SMGWeapon
}

func (m *WeaponSMG) finishReloading() {
	if m.mag < smgMag && m.ammo > 0 {
		totalAmmo := m.ammo + m.mag
		if totalAmmo > smgMag {
			m.mag = smgMag
		} else {
			m.mag = totalAmmo
		}
		m.ammo = totalAmmo - m.mag
	}
}

func (m *WeaponSMG) getLastSnapshot() *protocol.ObjectSnapshot {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if len(m.tickSnapshots) > 0 {
		return m.tickSnapshots[len(m.tickSnapshots)-1].Snapshot
	}
	return m.getCurrentSnapshot()
}

func (m *WeaponSMG) getLerpSnapshot() *protocol.ObjectSnapshot {
	return m.getSnapshotsByTime(ticktime.GetLerpTime())
}

func (m *WeaponSMG) getSnapshotsByTime(t time.Time) *protocol.ObjectSnapshot {
	m.lock.RLock()
	defer m.lock.RUnlock()
	_, b, _ := protocol.GetSnapshotByTime(t, m.tickSnapshots)
	if b == nil {
		b = m.getCurrentSnapshot()
	}
	ssB := b.Weapon.SMG
	return &protocol.ObjectSnapshot{
		ID:   m.GetID(),
		Type: m.GetType(),
		Weapon: &protocol.WeaponSnapshot{
			SMG: &protocol.WeaponSMGSnapshot{
				PlayerID:    ssB.PlayerID,
				Mag:         ssB.Mag,
				Ammo:        ssB.Ammo,
				TriggerTime: ssB.TriggerTime,
				ReloadTime:  ssB.ReloadTime,
			},
		},
	}
}

func (m *WeaponSMG) getCurrentSnapshot() *protocol.ObjectSnapshot {
	return &protocol.ObjectSnapshot{
		ID:   m.GetID(),
		Type: m.GetType(),
		Weapon: &protocol.WeaponSnapshot{
			SMG: &protocol.WeaponSMGSnapshot{
				PlayerID:    m.playerID,
				Mag:         m.mag,
				Ammo:        m.ammo,
				TriggerTime: m.triggerTime.UnixNano(),
				ReloadTime:  m.reloadTime.UnixNano(),
			},
		},
	}
}

func (m *WeaponSMG) cleanTickSnapshots() {
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

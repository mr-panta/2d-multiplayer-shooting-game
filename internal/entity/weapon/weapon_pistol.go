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
	pistolDropRate           = 40
	pistolWidth              = 72
	pistolBulletSpeed        = 2000
	pistolMaxRange           = 800
	pistolBulletLength       = 8
	pistolDamage             = 15
	pistolTriggerVisibleTime = time.Second
	pistolTriggerCooldown    = 400 * time.Millisecond
	pistolReloadCooldown     = 2 * time.Second
	pistolAmmo               = 40
	pistolMag                = 20
	pistolMaxScopeRadius     = 140
	pistolMaxScopeRange      = 400
	pistolRecoilAngle        = math.Pi / 180 * 4
)

type WeaponPistol struct {
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

func NewWeaponPistol(world common.World, id string) common.Weapon {
	return &WeaponPistol{
		id:     id,
		world:  world,
		pos:    util.GetHighVec(),
		posImd: imdraw.New(nil),
		dirImd: imdraw.New(nil),
		mag:    pistolMag,
	}
}

func (m *WeaponPistol) GetID() string {
	return m.id
}

func (m *WeaponPistol) Destroy() {
	m.isDestroyed = true
}

func (m *WeaponPistol) Exists() bool {
	return !m.isDestroyed
}

func (m *WeaponPistol) SetPos(pos pixel.Vec) {
	m.pos = pos
}

func (m *WeaponPistol) SetDir(dir pixel.Vec) {
	m.dir = dir
}

func (m *WeaponPistol) GetShape() pixel.Rect {
	min := m.pos.Sub(pixel.V(pistolWidth, pistolWidth).Scaled(0.5))
	max := m.pos.Add(pixel.V(pistolWidth, pistolWidth).Scaled(0.5))
	return pixel.Rect{Min: min, Max: max}
}

func (m *WeaponPistol) GetCollider() (pixel.Rect, bool) {
	return pixel.ZR, false
}

func (m *WeaponPistol) GetRenderObjects() []common.RenderObject {
	return nil
}

func (m *WeaponPistol) Render(target pixel.Target, viewPos pixel.Vec) {
	anim := animation.NewWeaponPistol()
	anim.Pos = m.pos.Sub(viewPos)
	anim.Dir = m.dir
	anim.TriggerTime = m.triggerTime
	anim.TriggerCooldown = pistolTriggerCooldown
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

func (m *WeaponPistol) GetType() int {
	return config.WeaponObject
}

func (m *WeaponPistol) renderPos(target pixel.Target, viewPos pixel.Vec) { // For debugging
	m.posImd.Clear()
	m.posImd.Color = colornames.Red
	m.posImd.Push(m.pos)
	m.posImd.Circle(2, 1)
	m.posImd.SetMatrix(pixel.IM.Moved(pixel.ZV.Sub(viewPos)))
	m.posImd.Draw(target)
}

func (m *WeaponPistol) renderDir(target pixel.Target, viewPos pixel.Vec) { // For debugging
	m.dirImd.Clear()
	m.dirImd.Color = colornames.Blue
	m.dirImd.Push(m.pos, m.pos.Add(m.dir.Unit().Scaled(80)))
	m.dirImd.Line(1)
	m.dirImd.SetMatrix(pixel.IM.Moved(pixel.ZV.Sub(viewPos)))
	m.dirImd.Draw(target)
}

func (m *WeaponPistol) ServerUpdate(tick int64) {
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
		m.isTriggering = now.Sub(m.triggerTime) < pistolTriggerCooldown
		isReloading := now.Sub(m.reloadTime) < pistolReloadCooldown
		if !isReloading && m.isReloading {
			m.finishReloading()
		}
		m.isReloading = isReloading
	}
	// Add snapshot
	m.SetSnapshot(tick, m.getCurrentSnapshot())
	m.cleanTickSnapshots()

}

func (m *WeaponPistol) ClientUpdate() {
	var now time.Time
	var ss *protocol.WeaponPistolSnapshot
	if m.playerID != m.world.GetMainPlayerID() {
		now = ticktime.GetServerTime()
		snapshot := m.getLastSnapshot()
		ss = snapshot.Weapon.Pistol
	} else {
		now = ticktime.GetLerpTime()
		snapshot := m.getLerpSnapshot()
		ss = snapshot.Weapon.Pistol
	}
	prevTriggerTime := m.triggerTime
	prevReloadTime := m.reloadTime
	m.playerID = ss.PlayerID
	m.mag = ss.Mag
	m.ammo = ss.Ammo
	m.triggerTime = time.Unix(0, ss.TriggerTime)
	m.reloadTime = time.Unix(0, ss.ReloadTime)
	m.isTriggering = now.Sub(m.triggerTime) < pistolTriggerCooldown
	m.isReloading = now.Sub(m.reloadTime) < pistolReloadCooldown
	if mainPlayer := m.world.GetMainPlayer(); mainPlayer != nil {
		dist := m.world.GetMainPlayer().GetPivot().Sub(m.pos).Len()
		if !ticktime.IsZeroTime(prevTriggerTime) && prevTriggerTime.Before(m.triggerTime) {
			sound.PlayWeaponPistolFire(dist)
		}
		if !ticktime.IsZeroTime(prevReloadTime) && prevReloadTime.Before(m.reloadTime) {
			sound.PlayWeaponPistolReload(dist)
		}
	}
	// Clean snapshot
	m.cleanTickSnapshots()
}

func (m *WeaponPistol) GetSnapshot(tick int64) (snapshot *protocol.ObjectSnapshot) {
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

func (m *WeaponPistol) SetSnapshot(tick int64, snapshot *protocol.ObjectSnapshot) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.tickSnapshots = append(m.tickSnapshots, &protocol.TickSnapshot{
		Tick:     tick,
		Snapshot: snapshot,
	})
}

func (m *WeaponPistol) Trigger() (ok bool) {
	ok = false
	if !m.isTriggering && m.mag > 0 && !m.isReloading {
		bullet := NewBullet(m.world, m.world.GetObjectDB().GetAvailableID())
		recoilAngle := rand.Float64()*pistolRecoilAngle - pistolRecoilAngle/2
		dir := m.dir.Rotated(recoilAngle)
		bullet.Fire(
			m.playerID,
			m.id,
			m.pos.Add(m.dir.Unit().Scaled(pistolWidth/2)),
			dir,
			pistolBulletSpeed,
			pistolMaxRange,
			pistolDamage,
			pistolBulletLength,
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

func (m *WeaponPistol) Reload() bool {
	if !m.isReloading && m.mag < pistolMag && m.ammo > 0 {
		m.reloadTime = ticktime.GetServerTime()
		return true
	}
	return false
}

func (m *WeaponPistol) StopReloading() {
	m.isReloading = false
	m.reloadTime = ticktime.GetServerStartTime()
}

func (m *WeaponPistol) SetPlayerID(playerID string) {
	m.playerID = playerID
}

func (m *WeaponPistol) AddAmmo(ammo int) bool {
	if m.ammo >= pistolAmmo {
		return false
	}
	if ammo == -1 {
		ammo = pistolAmmo
	} else if ammo == -2 {
		ammo = pistolMag
	}
	if m.ammo += ammo; m.ammo > pistolAmmo {
		m.ammo = pistolAmmo
	}
	return true
}
func (m *WeaponPistol) GetAmmo() (mag, ammo int) {
	return m.mag, m.ammo
}

func (m *WeaponPistol) GetScopeRadius(dist float64) float64 {
	if dist > pistolMaxScopeRange {
		return 0
	}
	return pistolMaxScopeRadius * (1.0 - (dist / pistolMaxScopeRange))
}

func (m *WeaponPistol) GetWeaponType() int {
	return config.PistolWeapon
}

func (m *WeaponPistol) GetTriggerVisibleTime() time.Duration {
	return pistolTriggerVisibleTime
}

func (m *WeaponPistol) finishReloading() {
	if m.mag < pistolMag && m.ammo > 0 {
		totalAmmo := m.ammo + m.mag
		if totalAmmo > pistolMag {
			m.mag = pistolMag
		} else {
			m.mag = totalAmmo
		}
		m.ammo = totalAmmo - m.mag
	}
}

func (m *WeaponPistol) getLastSnapshot() *protocol.ObjectSnapshot {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if len(m.tickSnapshots) > 0 {
		return m.tickSnapshots[len(m.tickSnapshots)-1].Snapshot
	}
	return m.getCurrentSnapshot()
}

func (m *WeaponPistol) getLerpSnapshot() *protocol.ObjectSnapshot {
	return m.getSnapshotsByTime(ticktime.GetLerpTime())
}

func (m *WeaponPistol) getSnapshotsByTime(t time.Time) *protocol.ObjectSnapshot {
	m.lock.RLock()
	defer m.lock.RUnlock()
	_, b, _ := protocol.GetSnapshotByTime(t, m.tickSnapshots)
	if b == nil {
		b = m.getCurrentSnapshot()
	}
	ssB := b.Weapon.Pistol
	return &protocol.ObjectSnapshot{
		ID:   m.GetID(),
		Type: m.GetType(),
		Weapon: &protocol.WeaponSnapshot{
			Pistol: &protocol.WeaponPistolSnapshot{
				PlayerID:    ssB.PlayerID,
				Mag:         ssB.Mag,
				Ammo:        ssB.Ammo,
				TriggerTime: ssB.TriggerTime,
				ReloadTime:  ssB.ReloadTime,
			},
		},
	}
}

func (m *WeaponPistol) getCurrentSnapshot() *protocol.ObjectSnapshot {
	return &protocol.ObjectSnapshot{
		ID:   m.GetID(),
		Type: m.GetType(),
		Weapon: &protocol.WeaponSnapshot{
			Pistol: &protocol.WeaponPistolSnapshot{
				PlayerID:    m.playerID,
				Mag:         m.mag,
				Ammo:        m.ammo,
				TriggerTime: m.triggerTime.UnixNano(),
				ReloadTime:  m.reloadTime.UnixNano(),
			},
		},
	}
}

func (m *WeaponPistol) cleanTickSnapshots() {
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

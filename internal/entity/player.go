package entity

import (
	"sync"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/animation"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/common"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/config"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/entity/item"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/protocol"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/ticktime"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/util"
)

const (
	playerShapeHeigth        = 128
	playerShapeWidth         = 48
	playerColliderSize       = 40
	playerWeaponMarginTop    = 24
	playerBaseMoveSpeed      = 300
	playerFrameTime          = 150
	playerZ                  = 10
	playerDropDiff           = 32
	playerInitHP             = 100
	playerRespawnTime        = 3 * time.Second
	playerHitHeightlightTime = 100 * time.Millisecond
	playerVisibleTime        = 1000 * time.Millisecond
	playerMaxScopeRadius     = 240
	playerMaxScopeRange      = 400
	playerStartRegenTime     = 3 * time.Second
	playerRegenRate          = 1
	playerSpeedCooldown      = 300 * time.Millisecond
	playerMaxPosError        = 100
)

type player struct {
	id            string
	weaponID      string
	world         common.World
	tickSnapshots []*protocol.TickSnapshot
	pos           pixel.Vec
	posError      pixel.Vec
	errorTime     time.Time
	updateTime    time.Time
	respawnTime   time.Time
	hitTime       time.Time
	triggerTime   time.Time
	isDestroyed   bool
	isMainPlayer  bool
	isDropping    bool
	isTriggering  bool
	isReloading   bool
	hp            float64
	maxMoveSpeed  float64
	moveSpeed     float64
	moveDir       pixel.Vec
	cursorDir     pixel.Vec
	lock          sync.RWMutex
	colliderImd   *imdraw.IMDraw
	shapeImd      *imdraw.IMDraw
}

func NewPlayer(world common.World, id string) common.Player {
	return &player{
		id:           id,
		world:        world,
		pos:          util.GetHighVec(),
		maxMoveSpeed: playerBaseMoveSpeed,
		updateTime:   ticktime.GetServerTime(),
		colliderImd:  imdraw.New(nil),
		shapeImd:     imdraw.New(nil),
		hp:           playerInitHP,
	}
}

func (p *player) GetID() string {
	return p.id
}

func (p *player) Destroy() {
	p.isDestroyed = true
}

func (p *player) Exists() bool {
	now := ticktime.GetServerTime()
	return now.After(p.respawnTime) && !p.isDestroyed
}

func (p *player) GetShape() pixel.Rect {
	return p.getShapeByPos(p.pos)
}

func (p *player) GetCollider() (pixel.Rect, bool) {
	return p.getCollider(), true
}

func (p *player) GetRenderObjects() (objs []common.RenderObject) {
	objs = append(objs, common.NewRenderObject(playerZ, p.GetShape(), p.render))
	// debug
	if config.EnvDebug() {
		if p.isMainPlayer {
			objs = append(objs, common.NewRenderObject(playerZ-2, p.GetShape(), p.renderLerp))
		}
		objs = append(objs, common.NewRenderObject(playerZ-1, p.GetShape(), p.renderLastSnapshot))
		objs = append(objs, common.NewRenderObject(playerZ+2, p.getCollider(), p.renderCollider))
		objs = append(objs, common.NewRenderObject(playerZ+1, p.GetShape(), p.renderShape))
	}
	return objs
}

func (p *player) GetPivot() pixel.Vec {
	ss := p.getLerpSnapshot().Player
	pos := ss.Pos.Convert()
	if p.isMainPlayer {
		pos = p.pos
	}
	return pos.Sub(pixel.V(0, playerWeaponMarginTop))
}

func (p *player) GetSnapshot(tick int64) (snapshot *protocol.ObjectSnapshot) {
	p.lock.RLock()
	defer p.lock.RUnlock()
	for i := len(p.tickSnapshots) - 1; i >= 0; i-- {
		ts := p.tickSnapshots[i]
		if ts.Tick == tick {
			snapshot = ts.Snapshot
		} else if ts.Tick < tick {
			break
		}
	}
	if snapshot == nil {
		snapshot = p.getCurrentSnapshot()
	}
	return snapshot
}

func (p *player) SetSnapshot(tick int64, snapshot *protocol.ObjectSnapshot) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.tickSnapshots = append(p.tickSnapshots, &protocol.TickSnapshot{
		Tick:     tick,
		Snapshot: snapshot,
	})
	if p.isMainPlayer {
		p.posError = p.pos.Sub(snapshot.Player.Pos.Convert())
		p.errorTime = ticktime.GetServerTime()
	}
}

func (p *player) ServerUpdate(tick int64) {
	now := ticktime.GetServerTime()
	// Check item
	if p.Exists() {
		for _, o := range p.world.GetObjectDB().SelectAll() {
			if o.GetType() != config.ItemObject || !o.GetShape().Intersects(p.getCollider()) {
				continue
			}
			item := o.(common.Item)
			if ok := item.UsedBy(p); ok {
				p.world.GetObjectDB().Delete(item.GetID())
			}
		}
	}
	// Update weapon
	if weapon := p.GetWeapon(); weapon != nil {
		weapon.SetPos(p.GetPivot())
		weapon.SetDir(p.cursorDir)
		// Interact weapon
		if p.isDropping {
			p.DropWeapon()
			p.isDropping = false
		} else if p.isReloading {
			weapon.Reload()
			p.isReloading = false
		} else if p.isTriggering {
			if weapon.Trigger() {
				p.triggerTime = now
			}
		}
	}
	// Update position
	moveSpeed := p.moveSpeed
	if now.Sub(p.triggerTime) < playerSpeedCooldown {
		moveSpeed /= 2
	}
	pos := p.pos
	diff := now.Sub(p.updateTime).Seconds()
	diffDist := p.moveDir.Unit().Scaled(moveSpeed * diff)
	pos = pos.Add(diffDist)
	// Check collision
	_, _, dynamicAdjust := p.world.CheckCollision(p.id, p.getCollider(), p.getColliderByPos(pos))
	p.pos = pos.Sub(dynamicAdjust)
	// Update HP
	if now.Sub(p.hitTime) > playerStartRegenTime {
		if p.hp += diff * playerRegenRate; p.hp > playerInitHP {
			p.hp = playerInitHP
		}
	}
	p.updateTime = now
	// Add snapshot
	p.SetSnapshot(tick, p.getCurrentSnapshot())
	p.cleanTickSnapshots()
}

func (p *player) ClientUpdate() {
	if p.isMainPlayer {
		// Set weapon
		ss := p.getLastSnapshot().Player
		p.weaponID = ss.WeaponID
		// Set status
		p.respawnTime = time.Unix(0, ss.RespawnTime)
		p.hitTime = time.Unix(0, ss.HitTime)
		p.triggerTime = time.Unix(0, ss.TriggerTime)
		p.hp = ss.HP
		// Update position
		now := ticktime.GetServerTime()
		moveSpeed := p.moveSpeed
		if now.Sub(p.triggerTime) < playerSpeedCooldown {
			moveSpeed /= 2
		}
		pos := p.pos
		diff := now.Sub(p.updateTime).Seconds()
		diffDist := p.moveDir.Unit().Scaled(moveSpeed * diff)
		pos = pos.Add(diffDist)
		// Correct error
		if ss.Pos.Convert().Sub(pos).Len() >= playerMaxPosError {
			pos = ss.Pos.Convert()
		} else {
			ms := 1000.0
			d := time.Duration(ms/config.ServerSyncRate) * time.Millisecond
			if now.Sub(p.errorTime) <= d {
				errorCorrectionDist := p.posError.Scaled(diff / d.Seconds())
				pos = pos.Sub(errorCorrectionDist)
			}
		}
		// Check collision
		_, _, dynamicAdjust := p.world.CheckCollision(p.id, p.getCollider(), p.getColliderByPos(pos))
		p.pos = pos.Sub(dynamicAdjust)
		p.updateTime = now

	} else {
		ss := p.getLerpSnapshot().Player
		// Set weapon
		p.weaponID = ss.WeaponID
		// Set status
		lastSS := p.getLastSnapshot().Player
		p.respawnTime = time.Unix(0, lastSS.RespawnTime)
		p.hitTime = time.Unix(0, lastSS.HitTime)
		p.triggerTime = time.Unix(0, lastSS.TriggerTime)
		// Update position
		p.hp = ss.HP
		p.pos = ss.Pos.Convert()
		p.moveDir = ss.MoveDir.Convert()
		p.cursorDir = ss.CursorDir.Convert()
		p.moveSpeed = ss.MoveSpeed
		p.maxMoveSpeed = ss.MaxMoveSpeed
	}
	// Update weapon
	if weapon := p.GetWeapon(); weapon != nil {
		weapon.SetPos(p.GetPivot())
		weapon.SetDir(p.cursorDir)
	}
	p.cleanTickSnapshots()
}

func (p *player) SetPos(pos pixel.Vec) {
	p.pos = pos
}

func (p *player) SetInput(input *protocol.InputSnapshot) {
	if input == nil {
		return
	}
	var moveSpeed float64
	moveDir := pixel.ZV
	if input.Up {
		moveDir.Y = 1
	} else if input.Down {
		moveDir.Y = -1
	}
	if input.Left {
		moveDir.X = -1
	} else if input.Right {
		moveDir.X = 1
	}
	if !moveDir.Eq(pixel.ZV) {
		moveSpeed = p.maxMoveSpeed
	}
	p.moveSpeed = moveSpeed
	p.moveDir = moveDir
	p.cursorDir = input.CursorDir.Convert()
	p.isDropping = input.Drop
	p.isTriggering = input.Fire
	p.isReloading = input.Reload
}

func (p *player) SetMainPlayer() {
	p.isMainPlayer = true
}

func (p *player) GetWeapon() common.Weapon {
	if p.weaponID == "" {
		return nil
	}
	if o, exists := p.world.GetObjectDB().SelectOne(p.weaponID); exists && o.GetType() == config.WeaponObject {
		return o.(common.Weapon)
	}
	return nil
}

func (p *player) SetWeapon(weapon common.Weapon) {
	if weapon != nil {
		p.weaponID = weapon.GetID()
	} else {
		p.weaponID = ""
	}
}

func (p *player) GetType() int {
	return config.PlayerObject
}

func (p *player) AddDamage(damage float64) {
	p.hp -= damage
	p.hitTime = ticktime.GetServerTime()
	if p.hp <= 0 {
		p.hp = playerInitHP
		p.respawnTime = ticktime.GetServerTime().Add(playerRespawnTime)
		p.DropWeapon()
		p.world.SpawnPlayer(p.id)
	}
}

func (p *player) GetHP() float64 {
	return p.hp
}

func (p *player) GetRespawnTime() time.Time {
	return p.respawnTime
}

func (p *player) GetScopeRadius(dist float64) float64 {
	if w := p.GetWeapon(); w != nil {
		return w.GetScopeRadius(dist)
	}
	if dist > playerMaxScopeRange {
		return 0
	}
	return playerMaxScopeRadius * (1.0 - (dist / playerMaxScopeRange))
}

func (p *player) DropWeapon() {
	if weapon := p.GetWeapon(); weapon != nil {
		p.SetWeapon(nil)
		weapon.SetPlayerID("")
		itemID := util.GenerateID()
		item := item.NewItemWeapon(p.world, itemID, weapon.GetID())
		item.SetPos(p.GetShape().Min.Sub(pixel.V(playerDropDiff, 0)))
		p.world.GetObjectDB().Set(item)
	}
}

func (p *player) GetHitTime() time.Time {
	return p.hitTime
}

func (p *player) GetTriggerTime() time.Time {
	return p.triggerTime
}

func (p *player) IsVisible() bool {
	now := ticktime.GetServerTime()
	return now.Sub(p.hitTime) <= playerVisibleTime || now.Sub(p.triggerTime) <= playerVisibleTime
}

func (p *player) getShapeByPos(pos pixel.Vec) pixel.Rect {
	min := pos.Sub(pixel.V(playerShapeWidth, playerShapeHeigth).Scaled(0.5))
	max := pos.Add(pixel.V(playerShapeWidth, playerShapeHeigth).Scaled(0.5))
	return pixel.Rect{Min: min, Max: max}
}

func (p *player) getCollider() pixel.Rect {
	return p.getColliderByPos(p.pos)
}

func (p *player) getColliderByPos(pos pixel.Vec) pixel.Rect {
	min := pos.Sub(pixel.V(playerColliderSize/2, playerShapeHeigth/2))
	max := pos.Add(pixel.V(playerColliderSize/2, playerColliderSize-playerShapeHeigth/2))
	return pixel.Rect{Min: min, Max: max}
}

func (p *player) render(target pixel.Target, viewPos pixel.Vec) {
	now := ticktime.GetServerTime()
	pos := p.GetShape().Center()
	base := pos.Sub(viewPos)
	anim := animation.NewCharacter()
	anim.Pos = base
	anim.Right = p.cursorDir.X > 0
	anim.FrameTime = playerFrameTime
	anim.Hit = now.Sub(p.hitTime) <= playerHitHeightlightTime
	anim.Shadow = true
	if p.moveSpeed == 0 {
		anim.State = animation.CharacterIdleState
		anim.FrameTime *= 2
	} else {
		anim.State = animation.CharacterRunState
	}
	moveSpeed := p.moveSpeed
	if now.Sub(p.triggerTime) < playerSpeedCooldown {
		moveSpeed /= 2
	}
	if moveSpeed > 0 {
		anim.FrameTime = int(float64(playerFrameTime*playerBaseMoveSpeed) / moveSpeed)
	}
	anim.Draw(target)
	if weapon := p.GetWeapon(); weapon != nil {
		weapon.Render(target, viewPos)
	}
}

func (p *player) renderLerp(target pixel.Target, viewPos pixel.Vec) { // For debugging
	objectSS := p.getLerpSnapshot()
	ss := objectSS.Player
	pos := p.getShapeByPos(ss.Pos.Convert()).Center()
	base := pos.Sub(viewPos)
	sprite := animation.NewCharacter()
	sprite.Color = config.LerpColor
	sprite.Pos = base
	sprite.Right = ss.CursorDir.X > 0
	sprite.FrameTime = playerFrameTime
	if ss.MoveSpeed == 0 {
		sprite.State = animation.CharacterIdleState
		ss.MoveSpeed *= 2
	} else {
		sprite.State = animation.CharacterRunState
	}
	if ss.MoveSpeed > 0 {
		sprite.FrameTime = int(float64(playerFrameTime*playerBaseMoveSpeed) / ss.MoveSpeed)
	}
	sprite.Draw(target)
}

func (p *player) renderLastSnapshot(target pixel.Target, viewPos pixel.Vec) { // For debugging
	objectSS := p.getLastSnapshot()
	ss := objectSS.Player
	pos := p.getShapeByPos(ss.Pos.Convert()).Center()
	base := pos.Sub(viewPos)
	sprite := animation.NewCharacter()
	sprite.Color = config.LashSnapshotColor
	sprite.Pos = base
	sprite.Right = ss.CursorDir.X > 0
	sprite.FrameTime = playerFrameTime
	if ss.MoveSpeed == 0 {
		sprite.State = animation.CharacterIdleState
		ss.MoveSpeed *= 2
	} else {
		sprite.State = animation.CharacterRunState
	}
	if ss.MoveSpeed > 0 {
		sprite.FrameTime = int(float64(playerFrameTime*playerBaseMoveSpeed) / ss.MoveSpeed)
	}
	sprite.Draw(target)
}

func (p *player) renderCollider(target pixel.Target, viewPos pixel.Vec) { // For debugging
	p.colliderImd.Clear()
	p.colliderImd.Color = config.ColliderColor
	p.colliderImd.Push(p.getCollider().Min, p.getCollider().Max)
	p.colliderImd.Rectangle(1)
	p.colliderImd.SetMatrix(pixel.IM.Moved(pixel.ZV.Sub(viewPos)))
	p.colliderImd.Draw(target)
}

func (p *player) renderShape(target pixel.Target, viewPos pixel.Vec) { // For debugging
	p.shapeImd.Clear()
	p.shapeImd.Color = config.ShapeColor
	p.shapeImd.Push(p.GetShape().Min, p.GetShape().Max)
	p.shapeImd.Rectangle(1)
	p.shapeImd.SetMatrix(pixel.IM.Moved(pixel.ZV.Sub(viewPos)))
	p.shapeImd.Draw(target)
}

func (p *player) getLastSnapshot() *protocol.ObjectSnapshot {
	p.lock.RLock()
	defer p.lock.RUnlock()
	if len(p.tickSnapshots) > 0 {
		return p.tickSnapshots[len(p.tickSnapshots)-1].Snapshot
	}
	return p.getCurrentSnapshot()
}

func (p *player) getLerpSnapshot() *protocol.ObjectSnapshot {
	return p.getSnapshotsByTime(ticktime.GetLerpTime())
}

func (p *player) getSnapshotsByTime(t time.Time) *protocol.ObjectSnapshot {
	p.lock.RLock()
	defer p.lock.RUnlock()
	a, b, d := protocol.GetSnapshotByTime(t, p.tickSnapshots)
	if a == nil || b == nil {
		a = p.getCurrentSnapshot()
		b = p.getCurrentSnapshot()
	}
	ssA := a.Player
	ssB := b.Player
	return &protocol.ObjectSnapshot{
		ID:   p.id,
		Type: config.PlayerObject,
		Player: &protocol.PlayerSnapshot{
			WeaponID:     ssB.WeaponID,
			CursorDir:    util.ConvertVec(pixel.Lerp(ssA.CursorDir.Convert(), ssA.CursorDir.Convert(), d)),
			Pos:          util.ConvertVec(pixel.Lerp(ssA.Pos.Convert(), ssB.Pos.Convert(), d)),
			MoveDir:      util.ConvertVec(pixel.Lerp(ssA.MoveDir.Convert(), ssB.MoveDir.Convert(), d)),
			MoveSpeed:    util.LerpScalar(ssA.MoveSpeed, ssB.MoveSpeed, d),
			MaxMoveSpeed: util.LerpScalar(ssA.MaxMoveSpeed, ssB.MaxMoveSpeed, d),
			HP:           util.LerpScalar(ssA.HP, ssB.HP, d),
			RespawnTime:  ssB.RespawnTime,
			HitTime:      ssB.HitTime,
			TriggerTime:  ssB.TriggerTime,
		},
	}
}

func (p *player) cleanTickSnapshots() {
	p.lock.Lock()
	defer p.lock.Unlock()
	if len(p.tickSnapshots) <= 1 {
		return
	}
	t := ticktime.GetServerTime().Add(-config.LerpPeriod * 2)
	tick := ticktime.GetTick(t)
	index := 0
	for i, ts := range p.tickSnapshots {
		if ts.Tick >= tick {
			index = i
			break
		}
	}
	if index > 0 {
		p.tickSnapshots = p.tickSnapshots[index:]
	}
}

func (p *player) getCurrentSnapshot() *protocol.ObjectSnapshot {
	return &protocol.ObjectSnapshot{
		ID:   p.id,
		Type: config.PlayerObject,
		Player: &protocol.PlayerSnapshot{
			WeaponID:     p.weaponID,
			CursorDir:    util.ConvertVec(p.cursorDir),
			Pos:          util.ConvertVec(p.pos),
			MoveDir:      util.ConvertVec(p.moveDir),
			MoveSpeed:    p.moveSpeed,
			MaxMoveSpeed: p.maxMoveSpeed,
			HP:           p.hp,
			RespawnTime:  p.respawnTime.UnixNano(),
			HitTime:      p.hitTime.UnixNano(),
			TriggerTime:  p.triggerTime.UnixNano(),
		},
	}
}

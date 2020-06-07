package common

import (
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/protocol"
)

// Menu

type Menu interface {
	UpdateAndRender()
}

// World

type World interface {
	// Common
	GetID() string
	GetType() int
	GetObjectDB() ObjectDB
	GetSize() (width, height int)
	CheckCollision(id string, prevCollider, nextCollider pixel.Rect) (
		obj Object, staticAdjust, dynamicAdjust pixel.Vec)
	GetHud() Hud
	// Client
	Render()
	GetWindow() *pixelgl.Window
	ClientUpdate() (exists bool)
	SetMainPlayerID(playerID string)
	GetMainPlayerID() string
	GetMainPlayer() Player
	SetSnapshot(tick int64, snapshot *protocol.WorldSnapshot)
	GetInputSnapshot() *protocol.InputSnapshot
	GetCameraViewPos() pixel.Vec
	GetScope() Scope
	// Server
	ServerUpdate(tick int64) (exists bool)
	SpawnPlayer(playerID string, playerName string)
	GetSnapshot(all bool) (tick int64, snapshot *protocol.WorldSnapshot)
	SetInputSnapshot(playerID string, snapshot *protocol.InputSnapshot)
}

// Processors

type ClientProcessor interface {
	Restart()
	ToggleFPSLimit()
	Close()
	GetWindow() *pixelgl.Window
	Run()
	StartWorld(hostIP, playerName string) (err error)
}

type ServerProcessor interface {
	Wait()
	UpdateWorld()
	BroadcastSnapshot()
	CleanWorld()
}

// Objects

type Object interface {
	GetID() string
	GetType() int
	Destroy()
	Exists() bool
	GetShape() pixel.Rect
	GetCollider() (pixel.Rect, bool)
	GetRenderObjects() []RenderObject
	GetSnapshot(tick int64) *protocol.ObjectSnapshot
	SetSnapshot(tick int64, snapshot *protocol.ObjectSnapshot)
	ServerUpdate(tick int64)
	ClientUpdate()
}

type Player interface {
	Object
	GetPos() pixel.Vec
	SetPos(pos pixel.Vec)
	SetMainPlayer()
	GetPivot() pixel.Vec
	SetInput(input *protocol.InputSnapshot)
	GetMeleeWeapon() Weapon
	SetMeleeWeapon(w Weapon)
	GetWeapon() Weapon
	SetWeapon(w Weapon)
	DropWeapon()
	IncreaseKill()
	GetStats() (kill, death, streak, maxStreak int)
	AddDamage(firingPlayerID, weaponID string, damage float64)
	GetArmorHP() (float64, float64)
	AddArmorHP(armor, hp float64) (canAdd bool)
	GetRespawnTime() time.Time
	GetHitTime() time.Time
	GetTriggerTime() time.Time
	GetScopeRadius(dist float64) float64
	SetVisibleCause(id string, visible bool)
	IsVisible() bool
	IsAlive() bool
	SetPlayerName(name string)
	GetPlayerName() string
}

type Item interface {
	Object
	SetPos(pos pixel.Vec)
	UsedBy(p Player) (ok bool)
}

type Weapon interface {
	Object
	GetWeaponType() int
	SetPos(pos pixel.Vec)
	SetDir(dir pixel.Vec)
	SetPlayerID(playerID string)
	Render(target pixel.Target, viewPos pixel.Vec)
	AddAmmo(ammo int) (canAdd bool)
	GetAmmo() (mag, ammo int)
	Trigger() bool
	Reload() bool
	StopReloading()
	GetScopeRadius(dist float64) float64
	GetTriggerVisibleTime() time.Duration
}

type Bullet interface {
	Object
	Fire(playerID, weaponID string, initPos, dir pixel.Vec,
		speed, maxRange, damage, length float64)
}

type Tree interface {
	Object
	SetState(pos pixel.Vec, treeType string, right bool)
}

type Terrain interface {
	Object
	GetTerrainType() int
	SetState(pos pixel.Vec, terrainType int)
}

type Boundary interface {
	Object
}

// Etc

type Hud interface {
	Render(target pixel.Target)
	GetRenderObjects() []RenderObject
	ClientUpdate()
	ServerUpdate()
	GetKillFeedSnapshot() *protocol.KillFeedSnapshot
	SetKillFeedSnapshot(snapshot *protocol.KillFeedSnapshot)
	AddKillFeedRow(killerPlayerID, victimPlayerID, weaponID string)
	GetScoreboardPlayers() []Player
}

type Scope interface {
	Update()
	GetRenderObject() RenderObject
	Intersects(shape pixel.Rect) bool
}

type Field interface {
	GetShape() pixel.Rect
	Render(target pixel.Target, viewPos pixel.Vec)
}

type Water interface {
	GetRenderObjects() []RenderObject
}

package common

import (
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/protocol"
)

// Core

type Core interface {
	Restart()
	Close()
	GetWindow() *pixelgl.Window
}

// World

type World interface {
	// Common
	GetObjectDB() ObjectDB
	CheckCollision(id string, prevCollider, nextCollider pixel.Rect) (
		obj Object, staticAdjust, dynamicAdjust pixel.Vec)
	GetSize() (width, height int)
	// Client
	Render()
	GetWindow() *pixelgl.Window
	ClientUpdate()
	SetMainPlayerID(playerID string)
	GetMainPlayerID() string
	SetSnapshot(tick int64, snapshot *protocol.WorldSnapshot)
	GetInputSnapshot() *protocol.InputSnapshot
	GetCameraViewPos() pixel.Vec
	GetScope() Scope
	// Server
	ServerUpdate(tick int64)
	SpawnPlayer(playerID string)
	GetSnapshot(all bool) (tick int64, snapshot *protocol.WorldSnapshot)
	SetInputSnapshot(playerID string, snapshot *protocol.InputSnapshot)
}

// Processors

type ClientProcessor interface {
	Restart()
	Close()
	GetWindow() *pixelgl.Window
	Run()
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
	SetPos(pos pixel.Vec)
	SetMainPlayer()
	GetPivot() pixel.Vec
	SetInput(input *protocol.InputSnapshot)
	GetWeapon() Weapon
	SetWeapon(w Weapon)
	DropWeapon()
	AddDamage(damage float64)
	GetHP() float64
	GetRespawnTime() time.Time
	GetHitTime() time.Time
	GetTriggerTime() time.Time
	GetScopeRadius(dist float64) float64
	IsVisible() bool
}

type Item interface {
	Object
	SetPos(pos pixel.Vec)
	UsedBy(p Player) (ok bool)
}

type Weapon interface {
	Object
	SetPos(pos pixel.Vec)
	SetDir(dir pixel.Vec)
	SetPlayerID(playerID string)
	Render(target pixel.Target, viewPos pixel.Vec)
	AddAmmo(ammo int) (canAdd bool)
	GetAmmo() (mag, ammo int)
	Trigger() bool
	Reload() bool
	GetScopeRadius(dist float64) float64
}

type Bullet interface {
	Object
	Fire(playerID string, initPos, dir pixel.Vec, speed, maxRange, damage, length float64)
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

// Etc

type Hud interface {
	Update()
	Render(target pixel.Target)
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

package protocol

import (
	"github.com/faiface/pixel"
)

const (
	// Server APIs
	CmdRegisterPlayer = 1
	CmdSetPlayerInput = 2
	// Client APIs
	CmdAddWorldSnapshot = 1
)

// Wrapper Objects

type CmdData struct {
	Cmd  int
	Data interface{}
}

type WrappedData struct {
	Cmd int `json:"cmd"`
	// Server APIs
	RegisterPlayer *RegisterPlayerRequest `json:"register_player,omitempty"`
	SetPlayerInput *SetPlayerInputRequest `json:"set_player_input,omitempty"`
	// Client APIs
	AddWorldSnapshot *AddWorldSnapshotRequest `json:"add_world_snapshot,omitempty"`
}

// Common Objects

type Vec struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

func (v *Vec) Convert() pixel.Vec {
	return pixel.V(v.X, v.Y)
}

type WorldSnapshot struct {
	ObjectSnapshots []*ObjectSnapshot `json:"object_snapshots,omitempty"`
}

type InputSnapshot struct {
	CursorDir *Vec `json:"cursor_dir,omitempty"`
	Fire      bool `json:"fire,omitempty"`
	Focus     bool `json:"focus,omitempty"`
	Up        bool `json:"up,omitempty"`
	Left      bool `json:"left,omitempty"`
	Down      bool `json:"down,omitempty"`
	Right     bool `json:"right,omitempty"`
	Reload    bool `json:"reload,omitempty"`
	Drop      bool `json:"drop,omitempty"`
}

type ObjectSnapshot struct {
	ID      string           `json:"id,omitempty"`
	Type    int              `json:"type,omitempty"`
	Player  *PlayerSnapshot  `json:"player,omitempty"`
	Item    *ItemSnapshot    `json:"item,omitempty"`
	Weapon  *WeaponSnapshot  `json:"weapon,omitempty"`
	Bullet  *BulletSnapshot  `json:"bullet,omitempty"`
	Tree    *TreeSnapshot    `json:"tree,omitempty"`
	Terrain *TerrainSnapshot `json:"terrain,omitempty"`
}

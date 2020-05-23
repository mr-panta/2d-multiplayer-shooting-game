package weapon

import (
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/common"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/protocol"
)

type weaponTickSnapshot struct {
	tick     int64
	snapshot *protocol.WeaponSnapshot
}

func (ts *weaponTickSnapshot) GetTick() int64 {
	return ts.tick
}

func (ts *weaponTickSnapshot) GetSnapshot() interface{} {
	return ts.snapshot
}

func New(world common.World, itemID string, snapshot *protocol.ObjectSnapshot) common.Weapon {
	if snapshot != nil && snapshot.Weapon != nil {
		ss := snapshot.Weapon
		if ss.M4 != nil {
			return NewWeaponM4(world, itemID)
		}
	}
	return nil
}

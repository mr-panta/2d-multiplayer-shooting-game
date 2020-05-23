package item

import (
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/common"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/protocol"
)

const (
	itemZ = 10
)

func New(world common.World, itemID string, snapshot *protocol.ObjectSnapshot) common.Item {
	if snapshot != nil && snapshot.Item != nil {
		if snapshot.Item.Weapon != nil {
			return NewItemWeapon(world, itemID, "")
		} else if snapshot.Item.Ammo != nil {
			return NewItemAmmo(world, itemID)
		} else if snapshot.Item.AmmoSM != nil {
			return NewItemAmmoSM(world, itemID)
		}
	}
	return nil
}

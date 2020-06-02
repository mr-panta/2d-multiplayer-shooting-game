package item

import (
	"time"

	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/common"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/protocol"
)

const (
	itemZ        = 10
	itemLifeTime = 60 * time.Second
)

func New(world common.World, itemID string, snapshot *protocol.ObjectSnapshot) common.Item {
	if snapshot != nil && snapshot.Item != nil {
		if snapshot.Item.Weapon != nil {
			return NewItemWeapon(world, itemID, "")
		} else if snapshot.Item.Ammo != nil {
			return NewItemAmmo(world, itemID)
		} else if snapshot.Item.AmmoSM != nil {
			return NewItemAmmoSM(world, itemID)
		} else if snapshot.Item.Armor != nil {
			return NewItemArmor(world, itemID, false)
		}
	}
	return nil
}

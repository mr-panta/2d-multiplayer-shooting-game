package weapon

import (
	"math/rand"

	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/common"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/protocol"
)

type newFunc func(world common.World, id string) common.Weapon

func New(world common.World, id string, snapshot *protocol.ObjectSnapshot) common.Weapon {
	if snapshot != nil && snapshot.Weapon != nil {
		ss := snapshot.Weapon
		if ss.M4 != nil {
			return NewWeaponM4(world, id)
		}
		if ss.Shotgun != nil {
			return NewWeaponShotgun(world, id)
		}
		if ss.Sniper != nil {
			return NewWeaponSniper(world, id)
		}
		if ss.Pistol != nil {
			return NewWeaponPistol(world, id)
		}
		if ss.SMG != nil {
			return NewWeaponSMG(world, id)
		}
		if ss.Knife != nil {
			return NewWeaponKnife(world, id)
		}
	}
	return nil
}

func Random(world common.World, id string) common.Weapon {
	dropRates := []int{
		sniperDropRate,
		pistolDropRate,
		shotgunDropRate,
		smgDropRate,
		m4DropRate,
	}
	newFnList := []newFunc{
		NewWeaponSniper,
		NewWeaponPistol,
		NewWeaponShotgun,
		NewWeaponSMG,
		NewWeaponM4,
	}
	totalDropRate := 0
	for _, dropRate := range dropRates {
		totalDropRate += dropRate
	}
	n := int(rand.Uint32()) % totalDropRate
	lower := 0
	for i, fn := range newFnList {
		upper := lower + dropRates[i]
		if lower <= n && n < upper {
			return fn(world, id)
		}
		lower = upper
	}
	return NewWeaponPistol(world, id)
}

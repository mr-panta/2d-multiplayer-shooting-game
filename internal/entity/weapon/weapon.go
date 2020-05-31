package weapon

import (
	"math/rand"

	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/config"

	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/common"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/protocol"
)

type newFunc func(world common.World, id string) common.Weapon

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
	}
	return nil
}

func Random(world common.World, id string) common.Weapon {
	sniperExists := false
	for _, obj := range world.GetObjectDB().SelectAll() {
		if obj.GetType() == config.WeaponObject {
			if w := obj.(common.Weapon); w.GetWeaponType() == config.SniperWeapon {
				sniperExists = true
				break
			}
		}
	}
	if sniperExists {
		return NewWeaponSniper(world, id)
	}
	dropRates := []int{
		pistolDropRate,
		shotgunDropRate,
		smgDropRate,
		m4DropRate,
	}
	newFnList := []newFunc{
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

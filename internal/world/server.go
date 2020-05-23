package world

import (
	"math/rand"
	"time"

	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/common"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/config"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/entity"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/entity/item"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/entity/weapon"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/protocol"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/ticktime"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/util"
	"github.com/mr-panta/go-logger"
)

const (
	minNextItemPerd = 15
	maxNextItemPerd = 30
)

// Common

func (w *world) ServerUpdate(tick int64) {
	w.tick = tick
	// Item
	if ticktime.GetServerTime().After(w.nextItemTime) {
		w.nextItemTime = w.spawnItem()
	}
	// Snapshot
	for _, o := range w.objectDB.SelectAll() {
		o.ServerUpdate(tick)
	}
}

func (w *world) GetSnapshot() (int64, *protocol.WorldSnapshot) {
	snapshot := &protocol.WorldSnapshot{}
	for _, o := range w.objectDB.SelectAll() {
		snapshot.ObjectSnapshots = append(
			snapshot.ObjectSnapshots,
			o.GetSnapshot(w.tick),
		)
	}
	return w.tick, snapshot
}

// Player

func (w *world) SpawnPlayer(playerID string) {
	var player common.Player
	o, exists := w.objectDB.SelectOne(playerID)
	if exists {
		player = o.(common.Player)
	} else {
		player = entity.NewPlayer(w, playerID)
	}
	// TODO: change random position
	player.SetPos(util.RandomVec(w.field))
	w.objectDB.Set(player)
}

func (w *world) SetInputSnapshot(playerID string, snapshot *protocol.InputSnapshot) {
	if o, exists := w.objectDB.SelectOne(playerID); exists && o.GetType() == config.PlayerObject {
		player := o.(common.Player)
		player.SetInput(snapshot)
	}
}

// Item

type spawnItemFn func() common.Item

func (w *world) spawnItem() (nextItemTime time.Time) {
	// Create item
	spawnItemFnList := []spawnItemFn{
		w.spawnWeaponItem,
		w.spawnAmmoItem,
		w.spawnAmmoSMItem,
	}
	// i := int(rand.Uint32()) % len(spawnItemFnList)
	for _, fn := range spawnItemFnList {
		item := fn()
		item.SetPos(util.RandomVec(w.field))
		w.objectDB.Set(item)
		logger.Debugf(nil, "spawn_item:%s", item.GetID())
	}
	// Random next item time
	n := rand.Int()%(maxNextItemPerd-minNextItemPerd) + minNextItemPerd
	return ticktime.GetServerTime().Add(time.Duration(n) * time.Second)
}

func (w *world) spawnWeaponItem() common.Item {
	itemID := util.GenerateID()
	weaponID := util.GenerateID()
	weapon := weapon.NewWeaponM4(w, weaponID)
	w.objectDB.Set(weapon)
	logger.Debugf(nil, "spawn_weapon:%s", weaponID)
	return item.NewItemWeapon(w, itemID, weaponID)
}

func (w *world) spawnAmmoItem() common.Item {
	itemID := util.GenerateID()
	return item.NewItemAmmo(w, itemID)
}

func (w *world) spawnAmmoSMItem() common.Item {
	itemID := util.GenerateID()
	return item.NewItemAmmoSM(w, itemID)
}

// Tree

func (w *world) createTrees() {
	for i := 0; i < w.treeAmount; i++ {
		treeID := util.GenerateID()
		logger.Debugf(nil, "create_tree:%s", treeID)
		tree := entity.NewTree(w, treeID)
		w.objectDB.Set(tree)
		pos := util.RandomVec(w.field)
		index := int(rand.Uint32()) % len(config.TreeTypes)
		treeType := config.TreeTypes[index]
		right := rand.Int()%2 != 0
		tree.SetState(pos, treeType, right)
	}
}

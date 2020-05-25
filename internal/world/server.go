package world

import (
	"math/rand"
	"time"

	"github.com/faiface/pixel"
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

func (w *world) GetSnapshot(all bool) (int64, *protocol.WorldSnapshot) {
	snapshot := &protocol.WorldSnapshot{}
	for _, o := range w.objectDB.SelectAll() {
		skip := o.GetType() == 0
		skip = skip || (!all && o.GetType() == config.TreeObject)
		skip = skip || (!all && o.GetType() == config.TerrainObject)
		if skip {
			continue
		}
		snapshot.ObjectSnapshots = append(
			snapshot.ObjectSnapshots,
			o.GetSnapshot(w.tick),
		)
	}
	return w.tick, snapshot
}

func (w *world) getFreePos() pixel.Vec {
	for i := 0; i < 10; i++ {
		pos := util.RandomVec(w.getSizeRect())
		rect := pixel.R(
			-worldMinSpawnDist,
			-worldMinSpawnDist,
			worldMinSpawnDist,
			worldMinSpawnDist,
		).Moved(pos)
		ok := true
		for _, obj := range w.objectDB.SelectAll() {
			if collider, exists := obj.GetCollider(); exists && collider.Intersects(rect) {
				ok = false
				break
			}
		}
		if ok {
			return pos
		}
	}
	return pixel.ZV
}

// Player

func (w *world) SpawnPlayer(playerID string, playerName string) {
	var player common.Player
	o, exists := w.objectDB.SelectOne(playerID)
	if exists {
		player = o.(common.Player)
	} else {
		player = entity.NewPlayer(w, playerID)
		player.SetPlayerName(playerName)
	}
	player.SetPos(w.getFreePos())
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
		item.SetPos(w.getFreePos())
		w.objectDB.Set(item)
		logger.Debugf(nil, "spawn_item:%s", item.GetID())
	}
	// Random next item time
	n := rand.Int()%(maxNextItemPerd-minNextItemPerd) + minNextItemPerd
	return ticktime.GetServerTime().Add(time.Duration(n) * time.Second)
}

func (w *world) spawnWeaponItem() common.Item {
	weaponID := w.objectDB.GetAvailableID()
	weapon := weapon.NewWeaponM4(w, weaponID)
	w.objectDB.Set(weapon)
	logger.Debugf(nil, "spawn_weapon:%s", weaponID)
	itemID := w.objectDB.GetAvailableID()
	return item.NewItemWeapon(w, itemID, weaponID)
}

func (w *world) spawnAmmoItem() common.Item {
	itemID := w.objectDB.GetAvailableID()
	return item.NewItemAmmo(w, itemID)
}

func (w *world) spawnAmmoSMItem() common.Item {
	itemID := w.objectDB.GetAvailableID()
	return item.NewItemAmmoSM(w, itemID)
}

// Props

func (w *world) createTrees() {
	for i := 0; i < worldTreeAmount; i++ {
		treeID := w.objectDB.GetAvailableID()
		logger.Debugf(nil, "create_tree:%s", treeID)
		tree := entity.NewTree(w, treeID)
		w.objectDB.Set(tree)
		pos := w.getFreePos()
		index := int(rand.Uint32()) % len(config.TreeTypes)
		treeType := config.TreeTypes[index]
		right := rand.Int()%2 != 0
		tree.SetState(pos, treeType, right)
	}
}

func (w *world) createTerrains() {
	for i := 0; i < worldTerrainAmount; i++ {
		terrainID := w.objectDB.GetAvailableID()
		logger.Debugf(nil, "create_terrain:%s", terrainID)
		terrain := entity.NewTerrain(w, terrainID)
		w.objectDB.Set(terrain)
		pos := w.getFreePos()
		terrainType := int(rand.Uint32()) % config.TerrainTypeAmount
		terrain.SetState(pos, terrainType)
	}
}

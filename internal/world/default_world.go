package world

import (
	"image/color"
	"math/rand"
	"sort"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/animation"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/common"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/config"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/entity"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/entity/item"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/entity/scoreboard"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/entity/weapon"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/protocol"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/sound"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/ticktime"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/util"
	"github.com/mr-panta/go-logger"
)

var (
	defaultWorldFieldSize = pixel.R(0, 0, 64, 64)
)

const (
	minNextItemPerd             = 10
	maxNextItemPerd             = 20
	defaultWorldFieldWidth      = 8
	defaultWorldFieldHeight     = 8
	defaultWorldTreeAmount      = 0
	defaultWorldTerrainAmount   = 12
	defaultWorldMinSpawnDist    = 48
	defaultWorldBoundarySize    = 200
	defaultWorldRestartCooldown = 5 * time.Second
)

type defaultWorld struct {
	// common
	id          string
	destroyed   bool
	objectDB    common.ObjectDB
	hud         common.Hud
	scoreboard  *scoreboard.DefaultScoreboard
	fieldWidth  int
	fieldHeight int
	skullID     string
	// client
	win              *pixelgl.Window
	toggleFPSLimit   func()
	batch            *pixel.Batch
	currRawInput     *common.RawInput
	prevRawInput     *common.RawInput
	currSettingInput *common.RawInput
	prevSettingInput *common.RawInput
	mainPlayerID     string
	cameraPos        pixel.Vec
	scope            common.Scope
	water            common.Water
	frameCount       int
	fpsUpdateTime    time.Time
	// server
	tick         int64
	nextItemTime time.Time
	destroyTime  time.Time
}

func NewDefaultWorld(clientProcessor common.ClientProcessor, id string) common.World {
	world := &defaultWorld{
		// common
		id:          id,
		objectDB:    common.NewObjectDB(),
		fieldWidth:  defaultWorldFieldWidth,
		fieldHeight: defaultWorldFieldHeight,
		// client
		currRawInput:     &common.RawInput{},
		prevRawInput:     &common.RawInput{},
		currSettingInput: &common.RawInput{},
		prevSettingInput: &common.RawInput{},
		// server
		nextItemTime: ticktime.GetServerTime(),
	}
	// common
	world.hud = entity.NewHud(world)
	world.scoreboard = scoreboard.NewDefaultScoreboard(world)
	if clientProcessor != nil {
		// client
		world.win = clientProcessor.GetWindow()
		world.toggleFPSLimit = clientProcessor.ToggleFPSLimit
		world.batch = pixel.NewBatch(&pixel.TrianglesData{}, animation.GetObjectSheet())
		world.scope = entity.NewScope(world)
		world.water = entity.NewWater(world)
		world.fpsUpdateTime = ticktime.GetServerTime()
	} else {
		// server
		world.createTrees()
		world.createTerrains()
		world.createBoundaries()
		world.createSkull()
	}
	return world
}

func (w *defaultWorld) GetID() string {
	return w.id
}

func (w *defaultWorld) GetType() int {
	return config.DefaultWorld
}

func (w *defaultWorld) GetObjectDB() common.ObjectDB {
	return w.objectDB
}

func (w *defaultWorld) CheckCollision(id string, prevCollider, nextCollider pixel.Rect) (
	obj common.Object, staticAdjust, dynamicAdjust pixel.Vec) {
	count := 0
	for _, o := range w.objectDB.SelectAll() {
		if !o.Exists() || o.GetID() == id {
			continue
		}
		if collider, exists := o.GetCollider(); exists {
			static, dynamic := util.CheckCollision(
				collider,
				prevCollider,
				nextCollider,
			)
			if static.Len() > 0 {
				obj = o
				staticAdjust = static
				dynamicAdjust = dynamic
				count++
			}
			if count > 1 {
				staticAdjust = nextCollider.Center().Sub(prevCollider.Center())
				dynamicAdjust = nextCollider.Center().Sub(prevCollider.Center())
				break
			}
		}
	}
	return obj, staticAdjust, dynamicAdjust
}

func (w *defaultWorld) GetHud() common.Hud {
	return w.hud
}

func (w *defaultWorld) GetSize() (width, height int) {
	return w.fieldWidth, w.fieldHeight
}

func (w *defaultWorld) getSizeRect() pixel.Rect {
	return pixel.R(
		0, 0,
		float64(w.fieldWidth)*defaultWorldFieldSize.W(),
		float64(w.fieldHeight)*defaultWorldFieldSize.H(),
	)
}

// Server

func (w *defaultWorld) ServerUpdate(tick int64) bool {
	if w.destroyed {
		return false
	}
	w.tick = tick
	// Item
	if ticktime.GetServerTime().After(w.nextItemTime) {
		w.nextItemTime = w.spawnItem()
	}
	// Snapshot
	for _, o := range w.objectDB.SelectAll() {
		o.ServerUpdate(tick)
	}
	// Kill feed
	w.hud.ServerUpdate()
	w.scoreboard.ServerUpdate()
	return true
}

func (w *defaultWorld) GetSnapshot(all bool) (int64, *protocol.WorldSnapshot) {
	snapshot := &protocol.WorldSnapshot{
		ID:               w.GetID(),
		Type:             w.GetType(),
		FieldWidth:       w.fieldWidth,
		FieldHeight:      w.fieldHeight,
		KillFeedSnapshot: w.hud.GetKillFeedSnapshot(),
	}
	for _, o := range w.objectDB.SelectAll() {
		skip := (!all && o.GetType() == config.BoundaryObject)
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

func (w *defaultWorld) Destroy() {
	w.destroyed = true
	w.destroyTime = ticktime.GetServerTime()
}

func (w *defaultWorld) createBoundaries() {
	size := w.getSizeRect()
	w.objectDB.Set(entity.NewBoundary(w, w.objectDB.GetAvailableID(), pixel.Rect{
		Min: pixel.V(size.Min.X-defaultWorldBoundarySize, size.Min.Y-defaultWorldBoundarySize),
		Max: pixel.V(size.Max.X+defaultWorldBoundarySize, size.Min.Y),
	}))
	w.objectDB.Set(entity.NewBoundary(w, w.objectDB.GetAvailableID(), pixel.Rect{
		Min: pixel.V(size.Min.X-defaultWorldBoundarySize, size.Max.Y),
		Max: pixel.V(size.Max.X+defaultWorldBoundarySize, size.Max.Y+defaultWorldBoundarySize),
	}))
	w.objectDB.Set(entity.NewBoundary(w, w.objectDB.GetAvailableID(), pixel.Rect{
		Min: pixel.V(size.Min.X-defaultWorldBoundarySize, size.Min.Y-defaultWorldBoundarySize),
		Max: pixel.V(size.Min.X, size.Max.Y+defaultWorldBoundarySize),
	}))
	w.objectDB.Set(entity.NewBoundary(w, w.objectDB.GetAvailableID(), pixel.Rect{
		Min: pixel.V(size.Max.X, size.Min.Y-defaultWorldBoundarySize),
		Max: pixel.V(size.Max.X+defaultWorldBoundarySize, size.Max.Y+defaultWorldBoundarySize),
	}))
}

func (w *defaultWorld) createSkull() {
	w.skullID = w.objectDB.GetAvailableID()
	skull := item.NewItemSkull(w, w.skullID)
	skull.SetPos(w.getFreePos())
	w.objectDB.Set(skull)
}

func (w *defaultWorld) getFreePos() pixel.Vec {
	for i := 0; i < 10; i++ {
		pos := util.RandomVec(w.getSizeRect())
		rect := pixel.R(
			-defaultWorldMinSpawnDist,
			-defaultWorldMinSpawnDist,
			defaultWorldMinSpawnDist,
			defaultWorldMinSpawnDist,
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

func (w *defaultWorld) SpawnPlayer(playerID string, playerName string) {
	var player common.Player
	o, exists := w.objectDB.SelectOne(playerID)
	if exists {
		player = o.(common.Player)
	} else {
		// Create Knife
		knifeID := w.GetObjectDB().GetAvailableID()
		weaponKnife := weapon.NewWeaponKnife(w, knifeID)
		weaponKnife.SetPlayerID(playerID)
		w.objectDB.Set(weaponKnife)
		// Create Player
		player = entity.NewPlayer(w, playerID)
		player.SetPlayerName(playerName)
		player.SetMeleeWeapon(weaponKnife)
	}
	player.SetPos(w.getFreePos())
	w.objectDB.Set(player)
}

func (w *defaultWorld) SetInputSnapshot(playerID string, snapshot *protocol.InputSnapshot) {
	if o, exists := w.objectDB.SelectOne(playerID); exists && o.GetType() == config.PlayerObject {
		player := o.(common.Player)
		player.SetInput(snapshot)
	}
}

// Item

type spawnItemFn func() common.Item

func (w *defaultWorld) spawnItem() (nextItemTime time.Time) {
	// Create item
	spawnItemFnList := []spawnItemFn{
		w.spawnWeaponItem,
		w.spawnAmmoItem,
		w.spawnAmmoSMItem,
		w.spawnLandMineItem,
	}
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

func (w *defaultWorld) spawnWeaponItem() common.Item {
	weaponID := w.objectDB.GetAvailableID()
	weapon := weapon.Random(w, weaponID)
	w.objectDB.Set(weapon)
	logger.Debugf(nil, "spawn_weapon:%s", weaponID)
	itemID := w.objectDB.GetAvailableID()
	return item.NewItemWeapon(w, itemID, weaponID)
}

func (w *defaultWorld) spawnAmmoItem() common.Item {
	itemID := w.objectDB.GetAvailableID()
	return item.NewItemAmmo(w, itemID)
}

func (w *defaultWorld) spawnAmmoSMItem() common.Item {
	itemID := w.objectDB.GetAvailableID()
	return item.NewItemAmmoSM(w, itemID)
}

func (w *defaultWorld) spawnLandMineItem() common.Item {
	itemID := w.objectDB.GetAvailableID()
	return item.NewItemLandMine(w, itemID)
}

// Props

func (w *defaultWorld) createTrees() {
	for i := 0; i < defaultWorldTreeAmount; i++ {
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

func (w *defaultWorld) createTerrains() {
	for i := 0; i < defaultWorldTerrainAmount; i++ {
		terrainID := w.objectDB.GetAvailableID()
		logger.Debugf(nil, "create_terrain:%s", terrainID)
		terrain := entity.NewTerrain(w, terrainID)
		w.objectDB.Set(terrain)
		pos := w.getFreePos()
		terrainType := int(rand.Uint32()) % config.TerrainTypeAmount
		terrain.SetState(pos, terrainType)
	}
}

// Client

func (w *defaultWorld) GetWindow() *pixelgl.Window {
	return w.win
}

func (w *defaultWorld) ClientUpdate() bool {
	now := ticktime.GetServerTime()
	if w.destroyed {
		return !(now.Sub(w.destroyTime) >= defaultWorldRestartCooldown)
	}
	w.updateRawInput()
	w.updateSetting()
	for _, o := range w.objectDB.SelectAll() {
		o.ClientUpdate()
	}
	w.hud.ClientUpdate()
	w.scoreboard.ClientUpdate()
	w.scope.Update()
	return true
}

func (w *defaultWorld) Render() {
	// Prepare rendered objects
	objects := []common.RenderObject{}
	renderObjects := []common.RenderObject{}
	for _, o := range w.objectDB.SelectAll() {
		playerVisible := true
		if o.GetType() == config.PlayerObject && w.mainPlayerID != o.GetID() {
			player := o.(common.Player)
			visible := player.IsVisible()
			visible = visible || w.scope.Intersects(o.GetShape())
			playerVisible = visible
		}
		if o.Exists() && playerVisible {
			objects = append(objects, o.GetRenderObjects()...)
		}
	}
	objects = append(objects, w.hud.GetRenderObjects()...)
	objects = append(objects, w.water.GetRenderObjects()...)
	// Filter
	for _, obj := range objects {
		if w.isInScreen(obj.GetShape()) {
			renderObjects = append(renderObjects, obj)
		}
	}
	// Add scope
	if obj := w.scope.GetRenderObject(); obj != nil {
		renderObjects = append(renderObjects, obj)
	}
	// Sort
	sort.Slice(renderObjects, func(i, j int) bool {
		if renderObjects[i].GetZ() == renderObjects[j].GetZ() {
			return renderObjects[i].GetShape().Min.Y > renderObjects[j].GetShape().Min.Y
		}
		return renderObjects[i].GetZ() < renderObjects[j].GetZ()
	})
	// Render
	w.win.Clear(color.RGBA{8, 168, 255, 255})
	windownRenderObjects := []common.RenderObject{}
	defaultSmoothObjects := []common.RenderObject{}
	nonSmoothObjects := []common.RenderObject{}
	for _, obj := range renderObjects {
		if obj.GetZ() >= config.MinWindowRenderZ {
			windownRenderObjects = append(windownRenderObjects, obj)
		} else if obj.GetZ() >= 0 {
			defaultSmoothObjects = append(defaultSmoothObjects, obj)
		} else {
			nonSmoothObjects = append(nonSmoothObjects, obj)
		}
	}
	// Render non smooth objects
	smooth := w.win.Smooth()
	w.win.SetSmooth(false)
	w.batch.Clear()
	for _, obj := range nonSmoothObjects {
		obj.Render(w.batch, w.GetCameraViewPos())
	}
	w.batch.Draw(w.win)
	w.win.SetSmooth(smooth)
	// Render default smooth objects
	w.batch.Clear()
	for _, obj := range defaultSmoothObjects {
		obj.Render(w.batch, w.GetCameraViewPos())
	}
	w.batch.Draw(w.win)
	// Render window render objects
	for _, obj := range windownRenderObjects {
		obj.Render(w.win, w.GetCameraViewPos())
	}
	// Render hud
	w.scoreboard.Render(w.win)
	w.hud.Render(w.win)
	// Update FPS
	w.updateFPS()
}

func (w *defaultWorld) SetSnapshot(tick int64, snapshot *protocol.WorldSnapshot) {
	if snapshot.ID != w.GetID() {
		if !w.destroyed {
			w.destroyed = true
			w.destroyTime = ticktime.GetServerTime()
		}
		logger.Debugf(nil, "get different world id, current_id=%s, new_id", w.GetID(), snapshot.ID)
		return
	}
	w.fieldWidth = snapshot.FieldWidth
	w.fieldHeight = snapshot.FieldHeight
	existsMap := make(map[string]bool)
	for _, ss := range snapshot.ObjectSnapshots {
		existsMap[ss.ID] = true
		o, exists := w.objectDB.SelectOne(ss.ID)
		if !exists {
			o = w.addObject(ss)
		}
		o.SetSnapshot(tick, ss)
	}
	for _, o := range w.objectDB.SelectAll() {
		if o.GetType() != 0 &&
			o.GetType() != config.BoundaryObject &&
			o.GetType() != config.TreeObject &&
			o.GetType() != config.TerrainObject &&
			!existsMap[o.GetID()] {
			w.removeObject(o.GetID())
		}
	}
	w.hud.SetKillFeedSnapshot(snapshot.KillFeedSnapshot)
}

func (w *defaultWorld) GetScope() common.Scope {
	return w.scope
}

func (w *defaultWorld) GetCameraViewPos() pixel.Vec {
	if w.GetMainPlayer() != nil {
		w.cameraPos = w.GetMainPlayer().GetPivot()
	}
	r := w.win.Bounds()
	return w.cameraPos.Sub(r.Center())
}

func (w *defaultWorld) addObject(ss *protocol.ObjectSnapshot) (o common.Object) {
	switch ss.Type {
	case config.PlayerObject:
		return w.addPlayer(ss)
	case config.ItemObject:
		return w.addItem(ss)
	case config.WeaponObject:
		return w.addWeapon(ss)
	case config.BulletObject:
		return w.addBullet(ss)
	case config.TreeObject:
		return w.addTree(ss)
	case config.TerrainObject:
		return w.addTerrain(ss)
	case config.BoundaryObject:
		return w.addBoundary(ss)
	}
	return nil
}

func (w *defaultWorld) removeObject(id string) {
	logger.Debugf(nil, "remove_object:%s", id)
	w.objectDB.Delete(id)
}

func (w *defaultWorld) isInScreen(r pixel.Rect) bool {
	r = r.Moved(pixel.ZV.Sub(w.GetCameraViewPos()))
	return r.Intersects(w.win.Bounds())
}

func (w *defaultWorld) updateFPS() {
	w.frameCount++
	now := ticktime.GetServerTime()
	if now.Sub(w.fpsUpdateTime) >= time.Second {
		ticktime.SetFPS(w.frameCount)
		w.frameCount = 0
		w.fpsUpdateTime = now
	}
}

// Input

func (w *defaultWorld) GetInputSnapshot() *protocol.InputSnapshot {
	player := w.GetMainPlayer()
	if player == nil {
		return nil
	}
	pivot := player.GetPivot().Sub(w.GetCameraViewPos())
	inputSS := &protocol.InputSnapshot{
		CursorDir:  util.ConvertVec(w.currRawInput.MousePos.Sub(pivot)),
		Fire:       w.currRawInput.PressedFireKey,
		Melee:      w.currRawInput.PressedMeleeKey,
		Up:         w.currRawInput.PressedUpKey,
		Left:       w.currRawInput.PressedLeftKey,
		Down:       w.currRawInput.PressedDownKey,
		Right:      w.currRawInput.PressedRightKey,
		Reload:     !w.prevRawInput.PressedReloadKey && w.currRawInput.PressedReloadKey,
		Drop:       !w.prevRawInput.PressedDropKey && w.currRawInput.PressedDropKey,
		Use1stItem: !w.prevRawInput.PressedUse1stItemKey && w.currRawInput.PressedUse1stItemKey,
		Use2ndItem: !w.prevRawInput.PressedUse2ndItemKey && w.currRawInput.PressedUse2ndItemKey,
		Use3rdItem: !w.prevRawInput.PressedUse3rdItemKey && w.currRawInput.PressedUse3rdItemKey,
	}
	w.prevRawInput = w.currRawInput
	w.currRawInput = &common.RawInput{}
	player.SetInput(inputSS)
	return inputSS
}

func (w *defaultWorld) getRawInput() *common.RawInput {
	return &common.RawInput{
		MousePos:                w.win.MousePosition(),
		PressedFireKey:          w.win.Pressed(config.FireKey),
		PressedMeleeKey:         w.win.Pressed(config.MeleeKey),
		PressedUpKey:            w.win.Pressed(config.UpKey),
		PressedLeftKey:          w.win.Pressed(config.LeftKey),
		PressedDownKey:          w.win.Pressed(config.DownKey),
		PressedRightKey:         w.win.Pressed(config.RightKey),
		PressedReloadKey:        w.win.Pressed(config.ReloadKey),
		PressedDropKey:          w.win.Pressed(config.DropKey),
		PressedUse1stItemKey:    w.win.Pressed(config.Use1stItemKey),
		PressedUse2ndItemKey:    w.win.Pressed(config.Use2ndItemKey),
		PressedUse3rdItemKey:    w.win.Pressed(config.Use3rdItemKey),
		PressedToggleMuteKey:    w.win.Pressed(config.ToggleMuteKey),
		PressedVolumeUpKey:      w.win.Pressed(config.VolumeUpKey),
		PressedVolumeDownKey:    w.win.Pressed(config.VolumeDownKey),
		PressedToggleFullScreen: w.win.Pressed(config.ToggleFullScreen),
		PressedToggleFPSLimit:   w.win.Pressed(config.ToggleFPSLimit),
	}
}

func (w *defaultWorld) updateRawInput() {
	rawInput := w.getRawInput()
	currRawInput := &common.RawInput{
		MousePos:             rawInput.MousePos,
		PressedFireKey:       rawInput.PressedFireKey || w.currRawInput.PressedFireKey,
		PressedMeleeKey:      rawInput.PressedMeleeKey || w.currRawInput.PressedMeleeKey,
		PressedUpKey:         rawInput.PressedUpKey || w.currRawInput.PressedUpKey,
		PressedLeftKey:       rawInput.PressedLeftKey || w.currRawInput.PressedLeftKey,
		PressedDownKey:       rawInput.PressedDownKey || w.currRawInput.PressedDownKey,
		PressedRightKey:      rawInput.PressedRightKey || w.currRawInput.PressedRightKey,
		PressedReloadKey:     rawInput.PressedReloadKey || w.currRawInput.PressedReloadKey,
		PressedDropKey:       rawInput.PressedDropKey || w.currRawInput.PressedDropKey,
		PressedUse1stItemKey: rawInput.PressedUse1stItemKey || w.currRawInput.PressedUse1stItemKey,
		PressedUse2ndItemKey: rawInput.PressedUse2ndItemKey || w.currRawInput.PressedUse2ndItemKey,
		PressedUse3rdItemKey: rawInput.PressedUse3rdItemKey || w.currRawInput.PressedUse3rdItemKey,
	}
	w.currRawInput = currRawInput
}

func (w *defaultWorld) updateSetting() {
	w.prevSettingInput = w.currSettingInput
	w.currSettingInput = w.getRawInput()
	// Settings
	if !w.prevSettingInput.PressedToggleMuteKey && w.currSettingInput.PressedToggleMuteKey {
		sound.ToggleMute()
	}
	if !w.prevSettingInput.PressedVolumeUpKey && w.currSettingInput.PressedVolumeUpKey {
		sound.VolumeUp()
	}
	if !w.prevSettingInput.PressedVolumeDownKey && w.currSettingInput.PressedVolumeDownKey {
		sound.VolumeDown()
	}
	if !w.prevSettingInput.PressedToggleFullScreen && w.currSettingInput.PressedToggleFullScreen {
		if w.win.Monitor() != nil {
			w.win.SetMonitor(nil)
		} else {
			w.win.SetMonitor(pixelgl.PrimaryMonitor())
		}
	}
	if !w.prevSettingInput.PressedToggleFPSLimit && w.currSettingInput.PressedToggleFPSLimit {
		w.toggleFPSLimit()
	}

}

// Player

func (w *defaultWorld) SetMainPlayerID(playerID string) {
	logger.Debugf(nil, "main_player_id:%s", playerID)
	w.mainPlayerID = playerID
}

func (w *defaultWorld) GetMainPlayerID() string {
	return w.mainPlayerID
}

func (w *defaultWorld) addPlayer(ss *protocol.ObjectSnapshot) common.Player {
	logger.Debugf(nil, "add_player:%s", ss.ID)
	player := entity.NewPlayer(w, ss.ID)
	if ss.ID == w.mainPlayerID {
		player.SetMainPlayer()
	}
	w.objectDB.Set(player)
	return player
}

func (w *defaultWorld) GetMainPlayer() common.Player {
	if w.mainPlayerID == "" {
		return nil
	}
	o, exists := w.objectDB.SelectOne(w.mainPlayerID)
	if !exists {
		return nil
	}
	return o.(common.Player)
}

// Item

func (w *defaultWorld) addItem(ss *protocol.ObjectSnapshot) common.Item {
	logger.Debugf(nil, "add_item:%s", ss.ID)
	item := item.New(w, ss.ID, ss)
	w.objectDB.Set(item)
	if ss.Item.Skull != nil {
		w.skullID = ss.ID
		w.scoreboard.SetSkullID(ss.ID)
	}
	return item
}

// Weapon

func (w *defaultWorld) addWeapon(snapshot *protocol.ObjectSnapshot) common.Weapon {
	logger.Debugf(nil, "add_weapon:%s", snapshot.ID)
	weapon := weapon.New(w, snapshot.ID, snapshot)
	w.objectDB.Set(weapon)
	return weapon
}

func (w *defaultWorld) addBullet(snapshot *protocol.ObjectSnapshot) common.Bullet {
	logger.Debugf(nil, "add_bullet:%s", snapshot.ID)
	bullet := weapon.NewBullet(w, snapshot.ID)
	w.objectDB.Set(bullet)
	return bullet
}

// Props

func (w *defaultWorld) addTree(ss *protocol.ObjectSnapshot) common.Tree {
	logger.Debugf(nil, "add_tree:%s", ss.ID)
	tree := entity.NewTree(w, ss.ID)
	w.objectDB.Set(tree)
	return tree
}

func (w *defaultWorld) addTerrain(ss *protocol.ObjectSnapshot) common.Terrain {
	logger.Debugf(nil, "add_terrain:%s", ss.ID)
	terrain := entity.NewTerrain(w, ss.ID)
	w.objectDB.Set(terrain)
	return terrain
}

func (w *defaultWorld) addBoundary(ss *protocol.ObjectSnapshot) common.Boundary {
	logger.Debugf(nil, "add_boundary:%s", ss.ID)
	boundary := entity.NewBoundary(w, ss.ID, pixel.ZR)
	w.objectDB.Set(boundary)
	return boundary
}

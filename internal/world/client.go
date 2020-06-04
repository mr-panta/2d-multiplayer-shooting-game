package world

import (
	"fmt"
	"image/color"
	"sort"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/common"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/config"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/entity"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/entity/item"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/entity/weapon"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/protocol"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/sound"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/ticktime"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/util"
	"github.com/mr-panta/go-logger"
)

// Common

func (w *world) GetWindow() *pixelgl.Window {
	return w.win
}

func (w *world) ClientUpdate() {
	w.updateRawInput()
	w.updateSetting()
	for _, o := range w.objectDB.SelectAll() {
		o.ClientUpdate()
	}
	w.hud.ClientUpdate()
	w.scope.Update()
}

func (w *world) Render() {
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
		if obj.GetZ() >= worldMinWindowRenderZ {
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
	w.hud.Render(w.win)
	// Update FPS
	w.updateFPS()
}

func (w *world) SetSnapshot(tick int64, snapshot *protocol.WorldSnapshot) {
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
			o.GetType() != config.TreeObject &&
			o.GetType() != config.TerrainObject &&
			!existsMap[o.GetID()] {
			w.removeObject(o.GetID())
		}
	}
	w.hud.SetKillFeedSnapshot(snapshot.KillFeedSnapshot)
}

func (w *world) GetScope() common.Scope {
	return w.scope
}

func (w *world) GetCameraViewPos() pixel.Vec {
	if w.GetMainPlayer() != nil {
		w.cameraPos = w.GetMainPlayer().GetPivot()
	}
	r := w.win.Bounds()
	return w.cameraPos.Sub(r.Center())
}

func (w *world) addObject(ss *protocol.ObjectSnapshot) (o common.Object) {
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
	}
	return nil
}

func (w *world) removeObject(id string) {
	logger.Debugf(nil, "remove_object:%s", id)
	w.objectDB.Delete(id)
}

func (w *world) isInScreen(r pixel.Rect) bool {
	r = r.Moved(pixel.ZV.Sub(w.GetCameraViewPos()))
	return r.Intersects(w.win.Bounds())
}

func (w *world) updateFPS() {
	w.frameCount++
	now := ticktime.GetServerTime()
	if now.Sub(w.fpsUpdateTime) >= time.Second {
		w.fps = w.frameCount
		w.frameCount = 0
		w.fpsUpdateTime = now
		fmt.Printf("FPS:%d|PING:%d\n", w.fps, ticktime.GetPing()/1000000)
	}
}

// Input

func (w *world) GetInputSnapshot() *protocol.InputSnapshot {
	player := w.GetMainPlayer()
	if player == nil {
		return nil
	}
	pivot := player.GetPivot().Sub(w.GetCameraViewPos())
	inputSS := &protocol.InputSnapshot{
		CursorDir: util.ConvertVec(w.currRawInput.MousePos.Sub(pivot)),
		Fire:      w.currRawInput.PressedFireKey,
		Melee:     w.currRawInput.PressedMeleeKey,
		Up:        w.currRawInput.PressedUpKey,
		Left:      w.currRawInput.PressedLeftKey,
		Down:      w.currRawInput.PressedDownKey,
		Right:     w.currRawInput.PressedRightKey,
		Reload:    !w.prevRawInput.PressedReloadKey && w.currRawInput.PressedReloadKey,
		Drop:      !w.prevRawInput.PressedDropKey && w.currRawInput.PressedDropKey,
	}
	w.prevRawInput = w.currRawInput
	w.currRawInput = &common.RawInput{}
	player.SetInput(inputSS)
	return inputSS
}

func (w *world) getRawInput() *common.RawInput {
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
		PressedToggleMuteKey:    w.win.Pressed(config.ToggleMuteKey),
		PressedVolumeUpKey:      w.win.Pressed(config.VolumeUpKey),
		PressedVolumeDownKey:    w.win.Pressed(config.VolumeDownKey),
		PressedToggleFullScreen: w.win.Pressed(config.ToggleFullScreen),
		PressedToggleFPSLimit:   w.win.Pressed(config.ToggleFPSLimit),
	}
}

func (w *world) updateRawInput() {
	rawInput := w.getRawInput()
	currRawInput := &common.RawInput{
		MousePos:         rawInput.MousePos,
		PressedFireKey:   rawInput.PressedFireKey || w.currRawInput.PressedFireKey,
		PressedMeleeKey:  rawInput.PressedMeleeKey || w.currRawInput.PressedMeleeKey,
		PressedUpKey:     rawInput.PressedUpKey || w.currRawInput.PressedUpKey,
		PressedLeftKey:   rawInput.PressedLeftKey || w.currRawInput.PressedLeftKey,
		PressedDownKey:   rawInput.PressedDownKey || w.currRawInput.PressedDownKey,
		PressedRightKey:  rawInput.PressedRightKey || w.currRawInput.PressedRightKey,
		PressedReloadKey: rawInput.PressedReloadKey || w.currRawInput.PressedReloadKey,
		PressedDropKey:   rawInput.PressedDropKey || w.currRawInput.PressedDropKey,
	}
	w.currRawInput = currRawInput
}

func (w *world) updateSetting() {
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

func (w *world) SetMainPlayerID(playerID string) {
	logger.Debugf(nil, "main_player_id:%s", playerID)
	w.mainPlayerID = playerID
}

func (w *world) GetMainPlayerID() string {
	return w.mainPlayerID
}

func (w *world) addPlayer(ss *protocol.ObjectSnapshot) common.Player {
	logger.Debugf(nil, "add_player:%s", ss.ID)
	player := entity.NewPlayer(w, ss.ID)
	if ss.ID == w.mainPlayerID {
		player.SetMainPlayer()
	}
	w.objectDB.Set(player)
	return player
}

func (w *world) GetMainPlayer() common.Player {
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

func (w *world) addItem(ss *protocol.ObjectSnapshot) common.Item {
	logger.Debugf(nil, "add_item:%s", ss.ID)
	item := item.New(w, ss.ID, ss)
	w.objectDB.Set(item)
	return item
}

// Weapon

func (w *world) addWeapon(snapshot *protocol.ObjectSnapshot) common.Weapon {
	logger.Debugf(nil, "add_weapon:%s", snapshot.ID)
	weapon := weapon.New(w, snapshot.ID, snapshot)
	w.objectDB.Set(weapon)
	return weapon
}

func (w *world) addBullet(snapshot *protocol.ObjectSnapshot) common.Bullet {
	logger.Debugf(nil, "add_bullet:%s", snapshot.ID)
	bullet := weapon.NewBullet(w, snapshot.ID)
	w.objectDB.Set(bullet)
	return bullet
}

// Props

func (w *world) addTree(ss *protocol.ObjectSnapshot) common.Tree {
	logger.Debugf(nil, "add_tree:%s", ss.ID)
	tree := entity.NewTree(w, ss.ID)
	w.objectDB.Set(tree)
	return tree
}

func (w *world) addTerrain(ss *protocol.ObjectSnapshot) common.Terrain {
	logger.Debugf(nil, "add_tree:%s", ss.ID)
	terrain := entity.NewTerrain(w, ss.ID)
	w.objectDB.Set(terrain)
	return terrain
}

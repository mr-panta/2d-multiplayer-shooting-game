package world

import (
	"image/color"
	"sort"
	"time"

	"github.com/faiface/pixel"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/animation"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/common"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/config"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/entity"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/entity/item"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/entity/weapon"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/protocol"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/ticktime"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/util"
)

const (
	playerVisibleTime = 1000 * time.Millisecond
)

// Common

func (w *world) ClientUpdate() {
	w.updateRawInput()
	for _, o := range w.objectDB.SelectAll() {
		o.ClientUpdate()
	}
	w.hud.Update()
	w.scope.Update(w.win)
}

func (w *world) Render() {
	// Prepare rendered objects
	objects := []common.RenderObject{}
	renderObjects := []common.RenderObject{}
	for _, o := range w.objectDB.SelectAll() {
		playerVisible := true
		if o.GetType() == config.PlayerObject && w.mainPlayerID != o.GetID() {
			player := o.(common.Player)
			visible := ticktime.GetServerTime().Sub(player.GetHitTime()) < playerVisibleTime
			visible = visible || ticktime.GetServerTime().Sub(player.GetTriggerTime()) < playerVisibleTime
			visible = visible || w.scope.Intersects(o.GetShape())
			playerVisible = visible
		}
		if o.Exists() && playerVisible {
			objects = append(objects, o.GetRenderObjects()...)
		}
	}
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
	w.win.Clear(color.RGBA{0xb0, 0xbb, 0x8d, 0xff})
	// Render fields
	w.renderFields()
	// Render objects
	for _, obj := range renderObjects {
		obj.Render(w.win, w.GetCameraViewPos())
	}
	// Render hud
	w.hud.Render(w.win)
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
		if o.GetType() != 0 && o.GetType() != config.TreeObject && !existsMap[o.GetID()] {
			w.removeObject(o.GetID())
		}
	}
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
	}
	return nil
}

func (w *world) removeObject(id string) {
	// logger.Debugf(nil, "remove_object:%s", id)
	w.objectDB.Delete(id)
}

func (w *world) isInScreen(r pixel.Rect) bool {
	r = r.Moved(pixel.ZV.Sub(w.GetCameraViewPos()))
	return r.Intersects(w.win.Bounds())
}

func (w *world) GetCameraViewPos() pixel.Vec {
	if w.getMainPlayer() != nil {
		w.cameraPos = w.getMainPlayer().GetPivot()
	}
	r := w.win.Bounds()
	return w.cameraPos.Sub(r.Center())
}

// Input

func (w *world) GetInputSnapshot() *protocol.InputSnapshot {
	player := w.getMainPlayer()
	if player == nil {
		return nil
	}
	pivot := player.GetPivot().Sub(w.GetCameraViewPos())
	inputSS := &protocol.InputSnapshot{
		CursorDir: util.ConvertVec(w.currRawInput.MousePos.Sub(pivot)),
		Fire:      w.currRawInput.PressedFireKey,
		Focus:     w.currRawInput.PressedFocusKey,
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
		MousePos:         w.win.MousePosition(),
		PressedFireKey:   w.win.Pressed(config.FireKey),
		PressedFocusKey:  w.win.Pressed(config.FocusKey),
		PressedUpKey:     w.win.Pressed(config.UpKey),
		PressedLeftKey:   w.win.Pressed(config.LeftKey),
		PressedDownKey:   w.win.Pressed(config.DownKey),
		PressedRightKey:  w.win.Pressed(config.RightKey),
		PressedReloadKey: w.win.Pressed(config.ReloadKey),
		PressedDropKey:   w.win.Pressed(config.DropKey),
	}
}

func (w *world) updateRawInput() {
	rawInput := w.getRawInput()
	currRawInput := &common.RawInput{
		MousePos:         rawInput.MousePos,
		PressedFireKey:   rawInput.PressedFireKey || w.currRawInput.PressedFireKey,
		PressedFocusKey:  rawInput.PressedFocusKey || w.currRawInput.PressedFocusKey,
		PressedUpKey:     rawInput.PressedUpKey || w.currRawInput.PressedUpKey,
		PressedLeftKey:   rawInput.PressedLeftKey || w.currRawInput.PressedLeftKey,
		PressedDownKey:   rawInput.PressedDownKey || w.currRawInput.PressedDownKey,
		PressedRightKey:  rawInput.PressedRightKey || w.currRawInput.PressedRightKey,
		PressedReloadKey: rawInput.PressedReloadKey || w.currRawInput.PressedReloadKey,
		PressedDropKey:   rawInput.PressedDropKey || w.currRawInput.PressedDropKey,
	}
	w.currRawInput = currRawInput
}

// Player

func (w *world) SetMainPlayerID(playerID string) {
	// logger.Debugf(nil, "main_player_id:%s", playerID)
	w.mainPlayerID = playerID
}

func (w *world) GetMainPlayerID() string {
	return w.mainPlayerID
}

func (w *world) addPlayer(ss *protocol.ObjectSnapshot) common.Player {
	// logger.Debugf(nil, "add_player:%s", ss.ID)
	player := entity.NewPlayer(w, ss.ID)
	if ss.ID == w.mainPlayerID {
		player.SetMainPlayer()
	}
	w.objectDB.Set(player)
	return player
}

func (w *world) getMainPlayer() common.Player {
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
	// logger.Debugf(nil, "add_item:%s", ss.ID)
	item := item.New(w, ss.ID, ss)
	w.objectDB.Set(item)
	return item
}

// Weapon

func (w *world) addWeapon(snapshot *protocol.ObjectSnapshot) common.Weapon {
	// logger.Debugf(nil, "add_weapon:%s", snapshot.ID)
	weapon := weapon.New(w, snapshot.ID, snapshot)
	w.objectDB.Set(weapon)
	return weapon
}

func (w *world) addBullet(snapshot *protocol.ObjectSnapshot) common.Bullet {
	// logger.Debugf(nil, "add_bullet:%s", snapshot.ID)
	bullet := weapon.NewBullet(w, snapshot.ID)
	w.objectDB.Set(bullet)
	return bullet
}

// Props

func (w *world) addTree(ss *protocol.ObjectSnapshot) common.Tree {
	// logger.Debugf(nil, "add_tree:%s", ss.ID)
	tree := entity.NewTree(w, ss.ID)
	w.objectDB.Set(tree)
	return tree
}

func (w *world) setupFields() {
	fields := []common.Field{}
	for i := 0; i < worldFieldHeight; i++ {
		for j := 0; j < worldFieldWidth; j++ {
			pos := pixel.V(
				float64(j)*worldFieldSize.W(),
				float64(i)*worldFieldSize.H(),
			)
			fields = append(fields, entity.NewField(pos))
		}
	}
	w.fields = fields
}

func (w *world) renderFields() {
	smooth := w.win.Smooth()
	w.win.SetSmooth(false)
	defer w.win.SetSmooth(smooth)
	if animation.FieldSheet != nil {
		batch := pixel.NewBatch(&pixel.TrianglesData{}, animation.FieldSheet)
		batch.Clear()
		for _, f := range w.fields {
			if w.isInScreen(f.GetShape()) {
				f.Render(batch, w.GetCameraViewPos())
			}
		}
		batch.Draw(w.win)
	}
}

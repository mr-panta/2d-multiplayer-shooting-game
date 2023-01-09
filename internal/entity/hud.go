package entity

import (
	"fmt"
	"image/color"
	"math"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/text"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/animation"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/common"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/config"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/protocol"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/ticktime"
	"golang.org/x/image/colornames"
)

var (
	// common
	hudZ = config.MinWindowRenderZ - 1
	// ammo
	hudAmmoMarginBottomRight = pixel.V(-24, 24)
	// armor and hp
	hudHPMarginBottomLeft    = pixel.V(24, 24)
	hudArmorMarginBottomLeft = pixel.V(188, 24)
	hudHPIconMargin          = pixel.V(24, 20)
	hudHPTextMrginLeft       = pixel.V(60, 0)
	hudArmorTextMrginLeft    = pixel.V(64, 0)
	// inventory
	hudInventoryMarginBottomLeft = pixel.V(24, 44)
	hudInventoryMargin           = pixel.V(48, 64)
	hudFirstItem                 = pixel.V(47, 118)
	hudSecondItem                = pixel.V(98, 108)
	// crosshair
	crosshairColor = colornames.Red
	// kill feed
	killFeedLimit           = 32
	killFeedLifeTime        = 10 * time.Second
	killFeedPadding         = 12.
	killFeedRowHeight       = 20.
	killFeedRowMargin       = 8.
	killFeedRowMarginTop    = 4.
	killFeedRowMarginBottom = 6.
	killFeedBGColor         = color.RGBA{0, 0, 0, 127}
	// game stats
	hudGameStatsOffset = pixel.V(8, 14)
)

type killFeedRow struct {
	createTime     time.Time
	killerPlayerID string
	victimPlayerID string
	weaponID       string
}

type Hud struct {
	world               common.World
	mag                 int
	ammo                int
	hp                  float64
	armor               float64
	respawnCountdown    int
	crosshair           *animation.Crosshair
	killFeedRowImds     []*imdraw.IMDraw
	killFeedRows        []*killFeedRow
	killFeedTxts        []*text.Text
	gameStatsText       *text.Text
	hpTxt               *text.Text
	armorTxt            *text.Text
	ammoTxt             *text.Text
	respawnCountdownTxt *text.Text
}

func NewHud(world common.World) common.Hud {
	killFeedRowImds := []*imdraw.IMDraw{}
	killFeedTxts := []*text.Text{}
	for i := 0; i < killFeedLimit; i++ {
		killFeedRowImds = append(killFeedRowImds, imdraw.New(nil))
		killFeedTxts = append(killFeedTxts, animation.NewText())
	}
	return &Hud{
		world:               world,
		crosshair:           animation.NewCrosshair(),
		killFeedRowImds:     killFeedRowImds,
		killFeedTxts:        killFeedTxts,
		gameStatsText:       animation.NewText(),
		hpTxt:               animation.NewText(),
		armorTxt:            animation.NewText(),
		ammoTxt:             animation.NewText(),
		respawnCountdownTxt: animation.NewText(),
	}
}

func (h *Hud) AddKillFeedRow(killerPlayerID, victimPlayerID, weaponID string) {
	h.killFeedRows = append(h.killFeedRows, &killFeedRow{
		createTime:     ticktime.GetServerTime(),
		killerPlayerID: killerPlayerID,
		victimPlayerID: victimPlayerID,
		weaponID:       weaponID,
	})
}

func (h *Hud) ClientUpdate() {
	h.updateAmmo()
	h.updateArmorHP()
	h.updateRespawnCountdown()
}

func (h *Hud) ServerUpdate() {
	h.updateKillFeed()
}

func (h *Hud) GetKillFeedSnapshot() *protocol.KillFeedSnapshot {
	ss := &protocol.KillFeedSnapshot{}
	for _, r := range h.killFeedRows {
		ss.Rows = append(ss.Rows, &protocol.KillFeedRow{
			CreateTime:     r.createTime.UnixNano(),
			KillerPlayerID: r.killerPlayerID,
			VictimPlayerID: r.victimPlayerID,
			WeaponID:       r.weaponID,
		})
	}
	return ss
}

func (h *Hud) SetKillFeedSnapshot(snapshot *protocol.KillFeedSnapshot) {
	rows := []*killFeedRow{}
	for _, r := range snapshot.Rows {
		rows = append(rows, &killFeedRow{
			createTime:     time.Unix(0, r.CreateTime),
			killerPlayerID: r.KillerPlayerID,
			victimPlayerID: r.VictimPlayerID,
			weaponID:       r.WeaponID,
		})
	}
	h.killFeedRows = rows
}

func (h *Hud) getPlayer() common.Player {
	if playerID := h.world.GetMainPlayerID(); playerID != "" {
		if o, exists := h.world.GetObjectDB().SelectOne(playerID); exists {
			return o.(common.Player)
		}
	}
	return nil
}

func (h *Hud) updateArmorHP() {
	var armor, hp float64
	if player := h.getPlayer(); player != nil {
		armor, hp = player.GetArmorHP()
	}
	h.hp = hp
	h.armor = armor
}

func (h *Hud) updateAmmo() {
	mag := 0
	ammo := 0
	if player := h.getPlayer(); player != nil {
		if weapon := player.GetWeapon(); weapon != nil {
			mag, ammo = weapon.GetAmmo()
		}
	}
	h.mag = mag
	h.ammo = ammo
}

func (h *Hud) updateRespawnCountdown() {
	countdown := 0
	if player := h.getPlayer(); player != nil {
		now := ticktime.GetServerTime()
		if d := player.GetRespawnTime().Sub(now); d > 0 {
			countdown = int(math.Ceil(d.Seconds()))
		}

	}
	h.respawnCountdown = countdown
}

func (h *Hud) updateKillFeed() {
	rows := []*killFeedRow{}
	now := ticktime.GetServerTime()
	for _, r := range h.killFeedRows {
		if now.Sub(r.createTime) <= killFeedLifeTime {
			rows = append(rows, r)
		}
	}
	offset := killFeedLimit
	if len(rows) < offset {
		offset = len(rows)
	}
	h.killFeedRows = rows[len(rows)-offset:]
}

func (h *Hud) Render(target pixel.Target) {
	win := h.world.GetWindow()
	smooth := win.Smooth()
	win.SetSmooth(false)
	defer win.SetSmooth(smooth)
	// render
	h.renderAmmo(target)
	h.renderHP(target)
	h.renderArmor(target)
	h.renderRespawnCountdown(target)
	h.renderCursor(target)
	h.renderKillFeed(target)
	h.renderFPS(target)
}

func (h *Hud) GetRenderObjects() []common.RenderObject {
	player := h.world.GetMainPlayer()
	if player == nil {
		return nil
	}
	p := player.GetPivot()
	shape := pixel.Rect{Min: p, Max: p}
	return []common.RenderObject{
		common.NewRenderObject(hudZ, shape, h.renderIcons),
	}
}

func (h *Hud) renderIcons(target pixel.Target, posView pixel.Vec) {
	// Render hp icon
	{
		pos := hudHPMarginBottomLeft
		icon := animation.NewIconHeart()
		icon.Pos = pos.Add(hudHPIconMargin)
		icon.Draw(target)
	}
	// Render armor icon
	{
		pos := hudArmorMarginBottomLeft
		icon := animation.NewIconShield()
		icon.Pos = pos.Add(hudHPIconMargin)
		icon.Draw(target)
	}
	// Render inventory
	{
		pos := hudInventoryMarginBottomLeft
		icon := animation.NewIconInventory()
		icon.Pos = pos.Add(hudInventoryMargin)
		icon.Draw(target)
	}
	// Render items
	h.renderItems(target)
}

func (h *Hud) renderItems(target pixel.Target) {
	player := h.getPlayer()
	items := player.GetItems()
	if item := items[0]; item != nil {
		if icon := item.GetIcon(); icon != nil {
			icon.Pos = hudFirstItem
			icon.Draw(target)
		}
	}
	if item := items[1]; item != nil {
		if icon := item.GetIcon(); icon != nil {
			icon.Pos = hudSecondItem
			icon.Draw(target)
		}
	}
}

func (h *Hud) renderHP(target pixel.Target) {
	pos := hudHPMarginBottomLeft.Add(hudHPTextMrginLeft)
	txt := h.hpTxt
	animation.DrawShadowTextLeft(
		txt,
		target,
		pos,
		fmt.Sprintf("%d", int(math.Ceil(h.hp))),
		4,
	)
}

func (h *Hud) renderArmor(target pixel.Target) {
	pos := hudArmorMarginBottomLeft.Add(hudArmorTextMrginLeft)
	txt := h.armorTxt
	animation.DrawShadowTextLeft(
		txt,
		target,
		pos,
		fmt.Sprintf("%d", int(math.Ceil(h.armor))),
		4,
	)
}

func (h *Hud) renderAmmo(target pixel.Target) {
	win := h.world.GetWindow()
	pos := pixel.V(win.Bounds().W(), 0).Add(hudAmmoMarginBottomRight)
	txt := h.ammoTxt
	animation.DrawShadowTextRight(
		txt,
		target,
		pos,
		fmt.Sprintf("%d / %d", h.mag, h.ammo),
		4,
	)
}

func (h *Hud) renderRespawnCountdown(target pixel.Target) {
	if h.respawnCountdown > 0 {
		pos := h.world.GetWindow().Bounds().Center()
		txt := h.respawnCountdownTxt
		animation.DrawStrokeTextCenter(
			txt,
			target,
			pos,
			fmt.Sprintf("%d", h.respawnCountdown),
			8, nil, nil,
		)
	}
}

func (h *Hud) renderCursor(target pixel.Target) {
	h.crosshair.Pos = h.world.GetWindow().MousePosition()
	h.crosshair.Color = crosshairColor
	h.crosshair.Draw(target)
}

func (h *Hud) renderKillFeed(target pixel.Target) {
	index := 0
	for _, row := range h.killFeedRows {
		if h.renderKillFeedRow(target, index, row) {
			index++
		}
	}
}

func (h *Hud) renderFPS(target pixel.Target) {
	fps := ticktime.GetFPS()
	ping := ticktime.GetPing() / 1000000
	win := h.world.GetWindow()
	animation.DrawShadowTextRight(
		h.gameStatsText,
		target,
		win.Bounds().Vertices()[2].Sub(hudGameStatsOffset),
		fmt.Sprintf("PING:%d FPS:%d", ping, fps),
		1,
	)
}

func (h *Hud) renderKillFeedRow(target pixel.Target, i int, row *killFeedRow) bool {
	// Prepare
	win := h.world.GetWindow()
	db := h.world.GetObjectDB()
	killerObj, exists := db.SelectOne(row.killerPlayerID)
	if !exists {
		return false
	}
	killer := killerObj.(common.Player)
	victimObj, exists := db.SelectOne(row.victimPlayerID)
	if !exists {
		return false
	}
	victim := victimObj.(common.Player)
	// weaponObj, exists := db.SelectOne(row.weaponID)
	// if !exists {
	// 	return
	// }
	// weapon := weaponObj.(common.Weapon)
	// Setup position
	message := fmt.Sprintf("%s > %s", killer.GetPlayerName(), victim.GetPlayerName())
	topRight := win.Bounds().Vertices()[2]
	pos := topRight.Sub(pixel.V(killFeedPadding, killFeedPadding))
	pos = pos.Sub(pixel.V(killFeedRowMargin, float64(i+1)*(killFeedRowHeight+killFeedRowMargin)))
	bounds := animation.GetTextRightBounds(pos, message, 1)
	bounds.Min = bounds.Min.Sub(pixel.V(killFeedRowMargin, killFeedRowMarginBottom))
	bounds.Max = bounds.Max.Add(pixel.V(killFeedRowMargin, killFeedRowMarginTop))
	// Render
	imd := h.killFeedRowImds[i]
	imd.Clear()
	imd.Color = killFeedBGColor
	imd.Push(bounds.Min, bounds.Max)
	imd.Rectangle(0)
	imd.Draw(target)
	animation.DrawShadowTextRight(h.killFeedTxts[i], target, pos, message, 1)
	return true
}

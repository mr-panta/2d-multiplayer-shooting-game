package entity

import (
	"fmt"
	"image/color"
	"math"
	"sort"
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
	"golang.org/x/image/font/basicfont"
)

var (
	// common
	hudZ         = 999
	shadowOffset = pixel.V(0.5, -0.5)
	// ammo
	hudAmmoMarginBottomRight = pixel.V(-24, 24)
	hudAmmoColor             = colornames.White
	// armor and hp
	hudHPMarginBottomLeft    = pixel.V(24, 24)
	hudArmorMarginBottomLeft = pixel.V(188, 24)
	hudHPIconMargin          = pixel.V(24, 20)
	hudHPTextMrginLeft       = pixel.V(60, 0)
	hudArmorTextMrginLeft    = pixel.V(64, 0)
	hudHPColor               = colornames.White
	// crosshair
	crosshairColor = colornames.Red
	// scoreboard
	scoreboardPadding          = 12.
	scoreboardWidth            = 160.
	scoreboardMarginLeft       = 12.
	scoreboardMarginTop        = 8.
	scoreboardMarginBottom     = 8.
	scoreboardLineHeight       = 16.
	scoreboardLimit            = 10
	scoreboardPlayerNameLength = 10
	scoreboardBGColor          = color.RGBA{0, 0, 0, 127}
	// kill feed
	killFeedLimit           = 32
	killFeedLifeTime        = 10 * time.Second
	killFeedPadding         = 12.
	killFeedRowHeight       = 20.
	killFeedRowMargin       = 8.
	killFeedRowMarginTop    = 4.
	killFeedRowMarginBottom = 6.
	killFeedBGColor         = color.RGBA{0, 0, 0, 127}
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
	scoreboardImd       *imdraw.IMDraw
	killFeedRowImds     []*imdraw.IMDraw
	killFeedRows        []*killFeedRow
	killFeedTxts        []*text.Text
	scoreboardPlayers   []common.Player
	scoreboardNameTxts  []*text.Text
	scoreboardScoreTxts []*text.Text
}

func NewHud(world common.World) common.Hud {
	killFeedRowImds := []*imdraw.IMDraw{}
	killFeedTxts := []*text.Text{}
	for i := 0; i < killFeedLimit; i++ {
		killFeedRowImds = append(killFeedRowImds, imdraw.New(nil))
		killFeedTxts = append(killFeedTxts, animation.NewText())
	}
	scoreboardNameTxts := []*text.Text{}
	scoreboardScoreTxts := []*text.Text{}
	for i := 0; i < scoreboardLimit*2; i++ {
		scoreboardNameTxts = append(scoreboardNameTxts, animation.NewText())
		scoreboardScoreTxts = append(scoreboardScoreTxts, animation.NewText())
	}
	return &Hud{
		world:               world,
		crosshair:           animation.NewCrosshair(),
		scoreboardImd:       imdraw.New(nil),
		killFeedRowImds:     killFeedRowImds,
		killFeedTxts:        killFeedTxts,
		scoreboardNameTxts:  scoreboardNameTxts,
		scoreboardScoreTxts: scoreboardScoreTxts,
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
	h.updateScoreboard()
}

func (h *Hud) ServerUpdate() {
	h.updateScoreboard()
	h.updateKillFeed()
}

func (h *Hud) GetScoreboardPlayers() []common.Player {
	return h.scoreboardPlayers
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

func (h *Hud) updateScoreboard() {
	players := []common.Player{}
	for _, obj := range h.world.GetObjectDB().SelectAll() {
		if obj.GetType() == config.PlayerObject {
			players = append(players, obj.(common.Player))
		}
	}
	sort.Slice(players, func(i, j int) bool {
		// i
		iName := players[i].GetPlayerName()
		iKill, iDeath, iStreak, iMaxStreak := players[i].GetStats()
		// j
		jName := players[j].GetPlayerName()
		jKill, jDeath, jStreak, jMaxStreak := players[j].GetStats()
		// compare
		if iStreak != jStreak {
			return iStreak > jStreak
		}
		if iMaxStreak != jMaxStreak {
			return iMaxStreak > jMaxStreak
		}
		if iKill != jKill {
			return iKill > jKill
		}
		if iDeath != jDeath {
			return iDeath < jDeath
		}
		return iName < jName
	})
	h.scoreboardPlayers = players
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
	h.renderScoreboard(target)
	h.renderKillFeed(target)
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
}

func (h *Hud) renderHP(target pixel.Target) {
	pos := hudHPMarginBottomLeft
	atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	txt := text.New(pos.Add(hudHPTextMrginLeft), atlas)
	txt.Clear()
	txt.LineHeight = atlas.LineHeight()
	txt.Color = color.Black
	fmt.Fprintf(txt, "%d", int(math.Ceil(h.hp)))
	m := pixel.IM.Scaled(txt.Orig, 4)
	txt.Draw(target, m.Moved(shadowOffset.Scaled(4)))
	txt.Color = hudHPColor
	fmt.Fprintf(txt, "\r%d", int(math.Ceil(h.hp)))
	txt.Draw(target, m)
}

func (h *Hud) renderArmor(target pixel.Target) {
	pos := hudArmorMarginBottomLeft
	atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	txt := text.New(pos.Add(hudArmorTextMrginLeft), atlas)
	txt.Clear()
	txt.LineHeight = atlas.LineHeight()
	txt.Color = color.Black
	fmt.Fprintf(txt, "%d", int(math.Ceil(h.armor)))
	m := pixel.IM.Scaled(txt.Orig, 4)
	txt.Draw(target, m.Moved(shadowOffset.Scaled(4)))
	txt.Color = hudHPColor
	fmt.Fprintf(txt, "\r%d", int(math.Ceil(h.armor)))
	txt.Draw(target, m)
}

func (h *Hud) renderAmmo(target pixel.Target) {
	win := h.world.GetWindow()
	pos := pixel.V(win.Bounds().W(), 0).Add(hudAmmoMarginBottomRight)
	atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	txt := text.New(pos, atlas)
	txt.Clear()
	txt.LineHeight = atlas.LineHeight()
	txt.Color = color.Black
	fmt.Fprintf(txt, "%d / %d", h.mag, h.ammo)
	m := pixel.IM.Moved(pixel.V(-txt.Bounds().W(), 0)).Scaled(txt.Orig, 4)
	txt.Draw(target, m.Moved(shadowOffset.Scaled(4)))
	txt.Color = hudAmmoColor
	fmt.Fprintf(txt, "\r%d / %d", h.mag, h.ammo)
	txt.Draw(target, m)
}

func (h *Hud) renderRespawnCountdown(target pixel.Target) {
	if h.respawnCountdown > 0 {
		pos := h.world.GetWindow().Bounds().Center()
		atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
		txt := text.New(pos, atlas)
		txt.Clear()
		txt.LineHeight = atlas.LineHeight()
		txt.Color = color.Black
		fmt.Fprintf(txt, "%d", h.respawnCountdown)
		m := pixel.IM.Moved(pixel.V(-txt.Bounds().W()/2, 0)).Scaled(txt.Bounds().Center(), 8)
		txt.Draw(target, m.Moved(shadowOffset.Scaled(8)))
		txt.Color = hudAmmoColor
		fmt.Fprintf(txt, "\r%d", h.respawnCountdown)
		txt.Draw(target, m)
	}
}

func (h *Hud) renderCursor(target pixel.Target) {
	h.crosshair.Pos = h.world.GetWindow().MousePosition()
	h.crosshair.Color = crosshairColor
	h.crosshair.Draw(target)
}

func (h *Hud) getScoreboard() (players []common.Player, mainPlayer common.Player, mainPlayerPlace int) {
	for i, player := range h.scoreboardPlayers {
		if i < scoreboardLimit {
			players = append(players, player)
		} else if player.GetID() == h.world.GetMainPlayerID() {
			mainPlayer = player
			mainPlayerPlace = i + 1
		}
	}
	return players, mainPlayer, mainPlayerPlace
}

func (h *Hud) renderScoreboard(target pixel.Target) {
	win := h.world.GetWindow()
	players, mainPlayer, mainPlayerPlace := h.getScoreboard()
	{
		pos := win.Bounds().Vertices()[1]
		pos = pos.Add(pixel.V(
			scoreboardPadding,
			-scoreboardPadding,
		))
		height := scoreboardLineHeight*float64(len(players)+1) + scoreboardMarginTop*2 + scoreboardMarginBottom
		if mainPlayer != nil {
			height += scoreboardLineHeight
		}
		h.scoreboardImd.Clear()
		h.scoreboardImd.Color = scoreboardBGColor
		h.scoreboardImd.EndShape = imdraw.RoundEndShape
		h.scoreboardImd.Push(pos, pos.Add(pixel.V(scoreboardWidth+scoreboardMarginLeft*2, -height)))
		h.scoreboardImd.Rectangle(0)
		h.scoreboardImd.Draw(win)
	}
	{
		//columns
		pos := win.Bounds().Vertices()[1]
		pos = pos.Add(pixel.V(
			scoreboardPadding+scoreboardMarginLeft,
			-scoreboardPadding-scoreboardMarginTop,
		))
		pos = pos.Add(pixel.V(0, -scoreboardLineHeight))
		animation.DrawShadowTextLeft(h.scoreboardNameTxts[0], target, pos, "PLAYER", 1)
		pos = pos.Add(pixel.V(scoreboardWidth, 0))
		animation.DrawShadowTextRight(h.scoreboardScoreTxts[0], target, pos, "STREAK", 1)
	}
	for i, player := range players {
		pos := win.Bounds().Vertices()[1]
		pos = pos.Add(pixel.V(
			scoreboardPadding+scoreboardMarginLeft,
			-scoreboardPadding-scoreboardMarginTop,
		))
		pos = pos.Add(pixel.V(0, -float64(i+2)*scoreboardLineHeight))
		playerName := player.GetPlayerName()
		if len(playerName) > scoreboardPlayerNameLength {
			playerName = playerName[:scoreboardPlayerNameLength] + "..."
		}
		animation.DrawShadowTextLeft(h.scoreboardNameTxts[i+1], target, pos, fmt.Sprintf("%d. %s", i+1, playerName), 1)
		_, _, streak, _ := player.GetStats()
		pos = pos.Add(pixel.V(scoreboardWidth, 0))
		animation.DrawShadowTextRight(h.scoreboardScoreTxts[i+1], target, pos, fmt.Sprint(streak), 1)
	}
	if mainPlayerPlace > 0 && mainPlayer != nil {
		playerLen := len(players)
		pos := win.Bounds().Vertices()[1]
		pos = pos.Add(pixel.V(
			scoreboardPadding+scoreboardMarginLeft,
			-scoreboardPadding-scoreboardMarginTop,
		))
		pos = pos.Add(pixel.V(0, -float64(playerLen+2)*scoreboardLineHeight))
		playerName := mainPlayer.GetPlayerName()
		if len(playerName) > scoreboardPlayerNameLength {
			playerName = playerName[:scoreboardPlayerNameLength] + "..."
		}
		animation.DrawShadowTextLeft(h.scoreboardNameTxts[scoreboardLimit+1], target, pos, fmt.Sprintf("%d. %s", mainPlayerPlace, playerName), 1)
		_, _, streak, _ := mainPlayer.GetStats()
		pos = pos.Add(pixel.V(scoreboardWidth, 0))
		animation.DrawShadowTextRight(h.scoreboardScoreTxts[scoreboardLimit+1], target, pos, fmt.Sprint(streak), 1)
	}
}

func (h *Hud) renderKillFeed(target pixel.Target) {
	index := 0
	for _, row := range h.killFeedRows {
		if h.renderKillFeedRow(target, index, row) {
			index++
		}
	}
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

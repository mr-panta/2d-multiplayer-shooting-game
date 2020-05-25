package entity

import (
	"fmt"
	"image/color"
	"math"
	"sort"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/text"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/animation"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/common"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/config"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/ticktime"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

var (
	// common
	shadowOffset = pixel.V(0.5, -0.5)
	// ammo
	hudAmmoMarginBottomRight = pixel.V(-24, 24)
	hudAmmoColor             = colornames.White
	// hp
	hudHPMarginBottomLeft = pixel.V(24, 24)
	hudHPColor            = colornames.White
	// crosshair
	crosshairColor = colornames.Red
	// scoreboard
	scoreboardPadding          = 12.
	scoreboardWidth            = 160.
	scoreboardMarginLeft       = 12.
	scoreboardMarginTop        = 8.
	scoreboardMarginBottom     = 8.
	scoreboardLineHeight       = 16.
	scoreboardSize             = 10
	scoreboardPlayerNameLength = 10
	scoreboardColor            = color.RGBA{0, 0, 0, 127}
)

type Hud struct {
	world             common.World
	mag               int
	ammo              int
	hp                float64
	respawnCountdown  int
	crosshair         *animation.Crosshair
	scoreboardPlayers []common.Player
	scoreboardImd     *imdraw.IMDraw
}

func NewHud(world common.World) common.Hud {
	return &Hud{
		world:         world,
		crosshair:     animation.NewCrosshair(),
		scoreboardImd: imdraw.New(nil),
	}
}

func (h *Hud) Update() {
	h.updateAmmo()
	h.updateHP()
	h.updateRespawnCountdown()
	h.updateScoreboard()
}

func (h *Hud) getPlayer() common.Player {
	if playerID := h.world.GetMainPlayerID(); playerID != "" {
		if o, exists := h.world.GetObjectDB().SelectOne(playerID); exists {
			return o.(common.Player)
		}
	}
	return nil
}

func (h *Hud) updateHP() {
	var hp float64
	if player := h.getPlayer(); player != nil {
		hp = player.GetHP()
	}
	h.hp = hp
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
		if iMaxStreak != jMaxStreak {
			return iMaxStreak > jMaxStreak
		}
		if iKill != jKill {
			return iKill > jKill
		}
		if iDeath != jDeath {
			return iDeath < jDeath
		}
		if iStreak != jStreak {
			return iStreak > jStreak
		}
		return iName < jName
	})
	h.scoreboardPlayers = players
}

func (h *Hud) Render(target pixel.Target) {
	win := h.world.GetWindow()
	smooth := win.Smooth()
	win.SetSmooth(false)
	defer win.SetSmooth(smooth)
	// render
	h.renderAmmo(target)
	h.renderHP(target)
	h.renderRespawnCountdown(target)
	h.renderCursor(target)
	h.renderScoreboard(target)
}

func (h *Hud) renderHP(target pixel.Target) {
	pos := hudHPMarginBottomLeft
	atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	txt := text.New(pos, atlas)
	txt.Clear()
	txt.LineHeight = atlas.LineHeight()
	txt.Color = color.Black
	fmt.Fprintf(txt, "HP: %d", int(math.Ceil(h.hp)))
	m := pixel.IM.Scaled(txt.Orig, 4)
	txt.Draw(target, m.Moved(shadowOffset.Scaled(4)))
	txt.Color = hudHPColor
	fmt.Fprintf(txt, "\rHP: %d", int(math.Ceil(h.hp)))
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
		if i < scoreboardSize {
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
		h.scoreboardImd.Color = scoreboardColor
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
		animation.DrawShadowTextLeft(target, pos, "#", 1)
		pos = pos.Add(pixel.V(scoreboardWidth, 0))
		animation.DrawShadowTextRight(target, pos, "K/D/S", 1)
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
		animation.DrawShadowTextLeft(target, pos, fmt.Sprintf("%d. %s", i+1, playerName), 1)
		kill, death, _, streak := player.GetStats()
		pos = pos.Add(pixel.V(scoreboardWidth, 0))
		animation.DrawShadowTextRight(target, pos, fmt.Sprintf("%d/%d/%d", kill, death, streak), 1)
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
		if len(playerName) > 10 {
			playerName = playerName[:10] + "..."
		}
		animation.DrawShadowTextLeft(target, pos, fmt.Sprintf("%d. %s", mainPlayerPlace, playerName), 1)
		kill, death, _, streak := mainPlayer.GetStats()
		pos = pos.Add(pixel.V(scoreboardWidth, 0))
		animation.DrawShadowTextRight(target, pos, fmt.Sprintf("%d/%d/%d", kill, death, streak), 1)
	}
}

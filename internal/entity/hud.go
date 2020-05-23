package entity

import (
	"fmt"
	"image/color"
	"math"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/animation"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/common"
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
)

type Hud struct {
	world            common.World
	mag              int
	ammo             int
	hp               float64
	respawnCountdown int
	crosshair        *animation.Crosshair
}

func NewHud(world common.World) common.Hud {
	return &Hud{
		world:     world,
		crosshair: animation.NewCrosshair(),
	}
}

func (h *Hud) Update() {
	h.updateAmmo()
	h.updateHP()
	h.updateRespawnCountdown()
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

func (h *Hud) Render(win *pixelgl.Window) {
	h.renderAmmo(win)
	h.renderHP(win)
	h.renderRespawnCountdown(win)
	h.renderCursor(win)
}

func (h *Hud) renderHP(win *pixelgl.Window) {
	smooth := win.Smooth()
	win.SetSmooth(false)
	defer win.SetSmooth(smooth)
	pos := hudHPMarginBottomLeft
	atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	txt := text.New(pos, atlas)
	txt.Clear()
	txt.LineHeight = atlas.LineHeight()
	txt.Color = color.Black
	fmt.Fprintf(txt, "HP: %d", int(math.Ceil(h.hp)))
	m := pixel.IM.Scaled(txt.Orig, 4)
	txt.Draw(win, m.Moved(shadowOffset.Scaled(4)))
	txt.Color = hudHPColor
	fmt.Fprintf(txt, "\rHP: %d", int(math.Ceil(h.hp)))
	txt.Draw(win, m)
}

func (h *Hud) renderAmmo(win *pixelgl.Window) {
	smooth := win.Smooth()
	win.SetSmooth(false)
	defer win.SetSmooth(smooth)
	pos := pixel.V(win.Bounds().W(), 0).Add(hudAmmoMarginBottomRight)
	atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	txt := text.New(pos, atlas)
	txt.Clear()
	txt.LineHeight = atlas.LineHeight()
	txt.Color = color.Black
	fmt.Fprintf(txt, "%d / %d", h.mag, h.ammo)
	m := pixel.IM.Moved(pixel.V(-txt.Bounds().W(), 0)).Scaled(txt.Orig, 4)
	txt.Draw(win, m.Moved(shadowOffset.Scaled(4)))
	txt.Color = hudAmmoColor
	fmt.Fprintf(txt, "\r%d / %d", h.mag, h.ammo)
	txt.Draw(win, m)
}

func (h *Hud) renderRespawnCountdown(win *pixelgl.Window) {
	smooth := win.Smooth()
	win.SetSmooth(false)
	defer win.SetSmooth(smooth)
	if h.respawnCountdown > 0 {
		pos := win.Bounds().Center()
		atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
		txt := text.New(pos, atlas)
		txt.Clear()
		txt.LineHeight = atlas.LineHeight()
		txt.Color = color.Black
		fmt.Fprintf(txt, "%d", h.respawnCountdown)
		m := pixel.IM.Moved(pixel.V(-txt.Bounds().W()/2, 0)).Scaled(txt.Bounds().Center(), 8)
		txt.Draw(win, m.Moved(shadowOffset.Scaled(8)))
		txt.Color = hudAmmoColor
		fmt.Fprintf(txt, "\r%d", h.respawnCountdown)
		txt.Draw(win, m)
	}
}

func (h *Hud) renderCursor(win *pixelgl.Window) {
	h.crosshair.Pos = win.MousePosition()
	h.crosshair.Color = crosshairColor
	h.crosshair.Draw(win)
}

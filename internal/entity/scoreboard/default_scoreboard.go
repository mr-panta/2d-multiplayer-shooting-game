package scoreboard

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
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/entity/item"
)

var (
	defaultScoreboardPadding          = 12.
	defaultScoreboardWidth            = 160.
	defaultScoreboardMarginLeft       = 12.
	defaultScoreboardMarginTop        = 8.
	defaultScoreboardMarginBottom     = 8.
	defaultScoreboardLineHeight       = 16.
	defaultScoreboardLimit            = 10
	defaultScoreboardPlayerNameLength = 10
	defaultScoreboardBGColor          = color.RGBA{0, 0, 0, 127}
)

type DefaultScoreboard struct {
	world               common.World
	skullID             string
	scoreboardImd       *imdraw.IMDraw
	scoreboardPlayers   []common.Player
	scoreboardNameTxts  []*text.Text
	scoreboardScoreTxts []*text.Text
}

func NewDefaultScoreboard(world common.World) *DefaultScoreboard {
	scoreboardNameTxts := []*text.Text{}
	scoreboardScoreTxts := []*text.Text{}
	for i := 0; i < defaultScoreboardLimit*2; i++ {
		scoreboardNameTxts = append(scoreboardNameTxts, animation.NewText())
		scoreboardScoreTxts = append(scoreboardScoreTxts, animation.NewText())
	}
	return &DefaultScoreboard{
		world:               world,
		scoreboardImd:       imdraw.New(nil),
		scoreboardNameTxts:  scoreboardNameTxts,
		scoreboardScoreTxts: scoreboardScoreTxts,
	}
}

func (s *DefaultScoreboard) getRemainingTimeMap() map[string]time.Duration {
	remainingTimeMap := make(map[string]time.Duration)
	if s.skullID == "" {
		return remainingTimeMap
	}
	obj, exists := s.world.GetObjectDB().SelectOne(s.skullID)
	if !exists {
		return remainingTimeMap
	}
	skull := obj.(*item.ItemSkull)
	remainingTimeMap = skull.GetRemainingTimeMap()
	return remainingTimeMap
}

func (s *DefaultScoreboard) SetSkullID(skullID string) {
	s.skullID = skullID
}

func (s *DefaultScoreboard) ClientUpdate() {
	players := []common.Player{}
	for _, obj := range s.world.GetObjectDB().SelectAll() {
		if obj.GetType() == config.PlayerObject {
			players = append(players, obj.(common.Player))
		}
	}
	remainingTimeMap := s.getRemainingTimeMap()
	sort.Slice(players, func(i, j int) bool {
		// i
		iName := players[i].GetPlayerName()
		iRemainingTime, exists := remainingTimeMap[players[i].GetID()]
		if !exists {
			iRemainingTime = config.DefaultWorldInitTime
		}
		// j
		jName := players[j].GetPlayerName()
		jRemainingTime, exists := remainingTimeMap[players[j].GetID()]
		if !exists {
			jRemainingTime = config.DefaultWorldInitTime
		}
		// compare
		if iRemainingTime != jRemainingTime {
			return iRemainingTime < jRemainingTime
		}
		return iName < jName
	})
	s.scoreboardPlayers = players
}

func (s *DefaultScoreboard) ServerUpdate() {
	// NOOP
}

func (s *DefaultScoreboard) Render(target pixel.Target) {
	win := s.world.GetWindow()
	smooth := win.Smooth()
	win.SetSmooth(false)
	defer win.SetSmooth(smooth)
	s.renderScoreboard(target)
}

func (s *DefaultScoreboard) getScoreboard() (players []common.Player, mainPlayer common.Player, mainPlayerPlace int) {
	for i, player := range s.scoreboardPlayers {
		if i < defaultScoreboardLimit {
			players = append(players, player)
		} else if player.GetID() == s.world.GetMainPlayerID() {
			mainPlayer = player
			mainPlayerPlace = i + 1
		}
	}
	return players, mainPlayer, mainPlayerPlace
}

func (s *DefaultScoreboard) renderScoreboard(target pixel.Target) {
	remainingTimeMap := s.getRemainingTimeMap()
	win := s.world.GetWindow()
	players, mainPlayer, mainPlayerPlace := s.getScoreboard()
	{
		pos := win.Bounds().Vertices()[1]
		pos = pos.Add(pixel.V(
			defaultScoreboardPadding,
			-defaultScoreboardPadding,
		))
		height := defaultScoreboardLineHeight*float64(len(players)+1) + defaultScoreboardMarginTop*2 + defaultScoreboardMarginBottom
		if mainPlayer != nil {
			height += defaultScoreboardLineHeight
		}
		s.scoreboardImd.Clear()
		s.scoreboardImd.Color = defaultScoreboardBGColor
		s.scoreboardImd.EndShape = imdraw.RoundEndShape
		s.scoreboardImd.Push(pos, pos.Add(pixel.V(defaultScoreboardWidth+defaultScoreboardMarginLeft*2, -height)))
		s.scoreboardImd.Rectangle(0)
		s.scoreboardImd.Draw(win)
	}
	{
		//columns
		pos := win.Bounds().Vertices()[1]
		pos = pos.Add(pixel.V(
			defaultScoreboardPadding+defaultScoreboardMarginLeft,
			-defaultScoreboardPadding-defaultScoreboardMarginTop,
		))
		pos = pos.Add(pixel.V(0, -defaultScoreboardLineHeight))
		animation.DrawShadowTextLeft(s.scoreboardNameTxts[0], target, pos, "PLAYER", 1)
		pos = pos.Add(pixel.V(defaultScoreboardWidth, 0))
		animation.DrawShadowTextRight(s.scoreboardScoreTxts[0], target, pos, "REMAINING", 1)
	}
	for i, player := range players {
		pos := win.Bounds().Vertices()[1]
		pos = pos.Add(pixel.V(
			defaultScoreboardPadding+defaultScoreboardMarginLeft,
			-defaultScoreboardPadding-defaultScoreboardMarginTop,
		))
		pos = pos.Add(pixel.V(0, -float64(i+2)*defaultScoreboardLineHeight))
		playerName := player.GetPlayerName()
		if len(playerName) > defaultScoreboardPlayerNameLength {
			playerName = playerName[:defaultScoreboardPlayerNameLength] + "..."
		}
		animation.DrawShadowTextLeft(s.scoreboardNameTxts[i+1], target, pos, fmt.Sprintf("%d. %s", i+1, playerName), 1)
		remainingTime, exists := remainingTimeMap[player.GetID()]
		if !exists {
			remainingTime = config.DefaultWorldInitTime
		}
		t := int(math.Ceil(remainingTime.Seconds()))
		pos = pos.Add(pixel.V(defaultScoreboardWidth, 0))
		animation.DrawShadowTextRight(s.scoreboardScoreTxts[i+1], target, pos, fmt.Sprint(t), 1)
	}
	if mainPlayerPlace > 0 && mainPlayer != nil {
		playerLen := len(players)
		pos := win.Bounds().Vertices()[1]
		pos = pos.Add(pixel.V(
			defaultScoreboardPadding+defaultScoreboardMarginLeft,
			-defaultScoreboardPadding-defaultScoreboardMarginTop,
		))
		pos = pos.Add(pixel.V(0, -float64(playerLen+2)*defaultScoreboardLineHeight))
		playerName := mainPlayer.GetPlayerName()
		if len(playerName) > defaultScoreboardPlayerNameLength {
			playerName = playerName[:defaultScoreboardPlayerNameLength] + "..."
		}
		animation.DrawShadowTextLeft(s.scoreboardNameTxts[defaultScoreboardLimit+1], target, pos, fmt.Sprintf("%d. %s", mainPlayerPlace, playerName), 1)
		remainingTime, exists := remainingTimeMap[mainPlayer.GetID()]
		if !exists {
			remainingTime = config.DefaultWorldInitTime
		}
		t := int(math.Ceil(remainingTime.Seconds()))
		pos = pos.Add(pixel.V(defaultScoreboardWidth, 0))
		animation.DrawShadowTextRight(s.scoreboardScoreTxts[defaultScoreboardLimit+1], target, pos, fmt.Sprint(t), 1)
	}
}

package client

import (
	"errors"
	"time"

	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/config"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/protocol"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/ticktime"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/world"
	"github.com/mr-panta/go-logger"
)

func (c *clientProcessor) registerPlayer(playerName string) error {
	now := time.Now()
	r, err := c.client.Send(
		protocol.CmdRegisterPlayer,
		&protocol.RegisterPlayerRequest{
			PlayerName: playerName,
			Version:    config.Version,
		},
	)
	ping := time.Since(now)
	if err != nil {
		logger.Debugf(nil, err.Error())
		return errors.New("CAN'T REGISTER PLAYER")
	}
	resp := r.(*protocol.RegisterPlayerResponse)
	if !resp.OK {
		return errors.New(resp.DebugMessage)
	}
	c.worldID = resp.WorldSnapshot.ID
	serverTime := time.Unix(0, resp.ServerTime)
	startTime := time.Unix(0, resp.StartTime)
	ticktime.SetServerTime(serverTime, ping)
	ticktime.SetServerStartTime(startTime)
	// Set world
	switch resp.WorldSnapshot.Type {
	case config.DefaultWorld:
		c.world = world.NewDefaultWorld(c, c.worldID)
	default:
		return errors.New("UNKNOWN WORLD TYPE")
	}
	c.world.SetMainPlayerID(resp.PlayerID)
	c.world.SetSnapshot(resp.Tick, resp.WorldSnapshot)
	return nil
}

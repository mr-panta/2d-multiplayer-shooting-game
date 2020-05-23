package client

import (
	"time"

	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/protocol"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/ticktime"
)

func (c *clientProcessor) registerPlayer() error {
	now := time.Now()
	r, err := c.client.Send(protocol.CmdRegisterPlayer, &protocol.RegisterPlayerRequest{})
	ping := time.Since(now)
	if err != nil {
		return err
	}
	resp := r.(*protocol.RegisterPlayerResponse)
	serverTime := time.Unix(0, resp.ServerTime)
	startTime := time.Unix(0, resp.StartTime)
	ticktime.SetServerTime(serverTime, ping)
	ticktime.SetServerStartTime(startTime)
	// Set world
	c.world.SetMainPlayerID(resp.PlayerID)
	c.world.SetSnapshot(resp.Tick, resp.WorldSnapshot)
	return nil
}

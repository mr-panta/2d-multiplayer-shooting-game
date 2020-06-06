package client

import (
	"context"
	"time"

	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/config"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/protocol"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/ticktime"
	"github.com/mr-panta/go-logger"
	"golang.org/x/time/rate"
)

func (p *clientProcessor) produceInputSnapshot() {
	ctx := context.Background()
	limiter := rate.NewLimiter(rate.Limit(config.ClientSyncRate), 1)
	worldID := p.worldID
	for worldID == p.worldID {
		_ = limiter.Wait(ctx)
		now := time.Now()
		_, err := p.client.Send(protocol.CmdSetPlayerInput, &protocol.SetPlayerInputRequest{
			PlayerID:      p.world.GetMainPlayerID(),
			InputSnapshot: p.world.GetInputSnapshot(),
		})
		ticktime.SetPing(time.Since(now))
		if err != nil {
			logger.Errorf(ctx, err.Error())
		}
	}
}

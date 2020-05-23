package server

import (
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/protocol"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/ticktime"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/util"
)

func (p *serverProcessor) processRegisterPlayer(request interface{}) (resp *protocol.RegisterPlayerResponse) {
	playerID := util.GenerateID()
	for {
		if _, exists := p.world.GetObjectDB().SelectOne(playerID); !exists {
			break
		}
		playerID = util.GenerateID()
	}
	p.world.SpawnPlayer(playerID)
	tick, worldSnapshot := p.world.GetSnapshot(true)
	return &protocol.RegisterPlayerResponse{
		PlayerID:      playerID,
		ServerTime:    ticktime.GetServerTime().UnixNano(),
		StartTime:     ticktime.GetServerStartTime().UnixNano(),
		Tick:          tick,
		WorldSnapshot: worldSnapshot,
	}
}

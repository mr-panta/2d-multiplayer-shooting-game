package server

import (
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/protocol"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/ticktime"
)

func (p *serverProcessor) processRegisterPlayer(request interface{}) (resp *protocol.RegisterPlayerResponse) {
	req := request.(*protocol.RegisterPlayerRequest)
	playerID := p.world.GetObjectDB().GetAvailableID()
	p.world.SpawnPlayer(playerID, req.PlayerName)
	tick, worldSnapshot := p.world.GetSnapshot(true)
	return &protocol.RegisterPlayerResponse{
		PlayerID:      playerID,
		ServerTime:    ticktime.GetServerTime().UnixNano(),
		StartTime:     ticktime.GetServerStartTime().UnixNano(),
		Tick:          tick,
		WorldSnapshot: worldSnapshot,
	}
}

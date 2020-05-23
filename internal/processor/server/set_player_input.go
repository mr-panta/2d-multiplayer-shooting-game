package server

import (
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/protocol"
)

func (p *serverProcessor) processSetPlayerInput(request interface{}) (resp *protocol.SetPlayerInputResponse) {
	req := request.(*protocol.SetPlayerInputRequest)
	p.world.SetInputSnapshot(req.PlayerID, req.InputSnapshot)
	p.markActiveTime(req.PlayerID)
	return &protocol.SetPlayerInputResponse{}
}

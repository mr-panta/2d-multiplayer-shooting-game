package server

import (
	"context"
	"encoding/json"

	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/config"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/network"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/protocol"
	"github.com/mr-panta/go-logger"
)

type GameProcess func(cmd int, req interface{}) (resp interface{})

type ServerNetwork interface {
	Start() error
	Wait()
	Close() error
	Broadcast(cmd int, data interface{}) error
}

type serverNetwork struct {
	server network.Server
}

func NewServerNetwork(gameProcess GameProcess) ServerNetwork {
	server := network.NewServer(
		config.TCPPortA,
		config.TCPPortB,
		translateProcess(gameProcess),
	)
	return &serverNetwork{
		server: server,
	}
}

func (s *serverNetwork) Start() error {
	return s.server.Start()
}

func (s *serverNetwork) Wait() {
	s.server.Wait()
}

func (s *serverNetwork) Close() error {
	return s.server.Close()
}

func translateProcess(gameProcess GameProcess) (process network.Process) {
	return func(reqBytes []byte) (respBytes []byte) {
		// Prepare
		ctx := context.Background()
		wrappedData := &protocol.WrappedData{}
		// Read req
		if err := json.Unmarshal(reqBytes, wrappedData); err != nil {
			logger.Errorf(ctx, err.Error())
			return []byte{}
		}
		// Command routing
		var resp interface{}
		switch wrappedData.Cmd {
		case protocol.CmdRegisterPlayer:
			resp = gameProcess(wrappedData.Cmd, wrappedData.RegisterPlayer)
		case protocol.CmdSetPlayerInput:
			resp = gameProcess(wrappedData.Cmd, wrappedData.SetPlayerInput)
		default:
			return []byte{}
		}
		// Write resp
		var err error
		respBytes, err = json.Marshal(resp)
		if err != nil {
			logger.Errorf(ctx, err.Error())
			return []byte{}
		}
		return respBytes
	}
}

func (s *serverNetwork) Broadcast(cmd int, data interface{}) error {
	wrappedData := &protocol.WrappedData{
		Cmd: cmd,
	}
	// Command routing
	switch cmd {
	case protocol.CmdAddWorldSnapshot:
		wrappedData.AddWorldSnapshot = data.(*protocol.AddWorldSnapshotRequest)
	}
	respBytes, err := json.Marshal(wrappedData)
	if err != nil {
		return err
	}
	s.server.Broadcast(respBytes)
	return nil
}

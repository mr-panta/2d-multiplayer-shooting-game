package client

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/config"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/network"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/protocol"
	"github.com/mr-panta/go-logger"
)

type ClientNetwork interface {
	Start() error
	Wait()
	Close() error
	Send(cmd int, req interface{}) (resp interface{}, err error)
	Listen() <-chan *protocol.CmdData
}

type clientNetwork struct {
	buffer   chan *protocol.CmdData
	client   network.Client
	isClosed bool
	lock     sync.RWMutex
}

func NewClientNetwork(hostIP string) ClientNetwork {
	client := network.NewClient(hostIP+config.TCPPortA, hostIP+config.TCPPortB)
	return &clientNetwork{
		buffer: make(chan *protocol.CmdData),
		client: client,
	}
}

func (c *clientNetwork) Start() error {
	go c.translateCmdData()
	return c.client.Start()
}

func (c *clientNetwork) Wait() {
	c.client.Wait()
}

func (c *clientNetwork) Close() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.isClosed = true
	close(c.buffer)
	return c.client.Close()
}

func (c *clientNetwork) Listen() <-chan *protocol.CmdData {
	return c.buffer
}

func (c *clientNetwork) translateCmdData() {
	for reqBytes := range c.client.Listen() {
		c.lock.RLock()
		if c.isClosed {
			break
		}
		// Prepare
		ctx := context.Background()
		wrappedData := &protocol.WrappedData{}
		// Read req
		if err := json.Unmarshal(reqBytes, wrappedData); err != nil {
			logger.Errorf(ctx, err.Error())
			continue
		}
		// Command routing
		var data interface{}
		switch wrappedData.Cmd {
		case protocol.CmdAddWorldSnapshot:
			data = wrappedData.AddWorldSnapshot
		default:
			continue
		}
		// Passing
		c.buffer <- &protocol.CmdData{
			Cmd:  wrappedData.Cmd,
			Data: data,
		}
		c.lock.RUnlock()
	}
}

func (c *clientNetwork) Send(cmd int, req interface{}) (resp interface{}, err error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if c.isClosed {
		return nil, fmt.Errorf("processor network client is closed")
	}
	wrappedData := &protocol.WrappedData{
		Cmd: cmd,
	}
	// Command routing
	switch cmd {
	case protocol.CmdRegisterPlayer:
		wrappedData.RegisterPlayer = req.(*protocol.RegisterPlayerRequest)
		resp = &protocol.RegisterPlayerResponse{}
	case protocol.CmdSetPlayerInput:
		wrappedData.SetPlayerInput = req.(*protocol.SetPlayerInputRequest)
		resp = &protocol.SetPlayerInputResponse{}
	}
	// Send and receive data
	reqBytes, err := json.Marshal(wrappedData)
	if err != nil {
		return nil, err
	}
	respBytes, err := c.client.Send(reqBytes)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(respBytes, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

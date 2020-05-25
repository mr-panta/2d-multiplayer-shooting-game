package server

import (
	"context"
	"sync"
	"time"

	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/common"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/config"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/protocol"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/ticktime"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/world"
	"github.com/mr-panta/go-logger"
)

type serverProcessor struct {
	server             ServerNetwork
	world              common.World
	lastActiveTimeMap  map[string]time.Time
	lastActiveTimeLock sync.RWMutex
}

func NewServerProcessor() (common.ServerProcessor, error) {
	p := &serverProcessor{
		lastActiveTimeMap: make(map[string]time.Time),
	}
	p.server = NewServerNetwork(p.process)
	if err := p.server.Start(); err != nil {
		return nil, err
	}
	p.world = world.New(nil)
	return p, nil
}

func (p *serverProcessor) Wait() {
	p.server.Wait()
}

func (p *serverProcessor) UpdateWorld() {
	ticktime.SetServerStartTime(time.Now())
	for tick := int64(0); ; tick++ {
		tickTime := ticktime.GetTickTime(tick)
		waitTime := time.Until(tickTime)
		<-time.NewTimer(waitTime).C
		p.world.ServerUpdate(tick)
	}
}

func (p *serverProcessor) BroadcastSnapshot() {
	ticker := time.NewTicker(time.Second / config.ServerSyncRate)
	ctx := context.Background()
	for range ticker.C {
		tick, snapshot := p.world.GetSnapshot(false)
		req := &protocol.AddWorldSnapshotRequest{
			Tick:          tick,
			WorldSnapshot: snapshot,
		}
		if err := p.server.Broadcast(protocol.CmdAddWorldSnapshot, req); err != nil {
			logger.Errorf(ctx, err.Error())
		}
	}
}

func (p *serverProcessor) CleanWorld() {
	ticker := time.NewTicker(time.Second / config.ServerSyncRate)
	for range ticker.C {
		p.lastActiveTimeLock.RLock()
		now := ticktime.GetServerTime()
		for playerID, lastActiveTime := range p.lastActiveTimeMap {
			if now.Sub(lastActiveTime) > config.PlayerTimeOut {
				if o, exists := p.world.GetObjectDB().SelectOne(playerID); exists {
					player := o.(common.Player)
					player.DropWeapon()
				}
				p.world.GetObjectDB().Delete(playerID)
			}
		}
		p.lastActiveTimeLock.RUnlock()
	}
}

func (p *serverProcessor) markActiveTime(playerID string) {
	p.lastActiveTimeLock.Lock()
	defer p.lastActiveTimeLock.Unlock()
	p.lastActiveTimeMap[playerID] = ticktime.GetServerTime()
}

func (p *serverProcessor) process(cmd int, req interface{}) (resp interface{}) {
	switch cmd {
	case protocol.CmdRegisterPlayer:
		resp = p.processRegisterPlayer(req)
	case protocol.CmdSetPlayerInput:
		resp = p.processSetPlayerInput(req)
	}
	return resp
}

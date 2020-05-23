package client

import (
	"context"
	"fmt"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/animation"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/common"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/config"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/protocol"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/world"
	"golang.org/x/time/rate"
)

type clientProcessor struct {
	win          *pixelgl.Window
	restartSig   chan bool
	restartCount int
	world        common.World
	client       ClientNetwork
}

func NewClientProcessor() (processor common.ClientProcessor, err error) {
	p := &clientProcessor{
		restartSig: make(chan bool),
	}
	cfg := config.GetConfig()
	// Create win
	winCfg := pixelgl.WindowConfig{
		Title:     config.Title,
		Bounds:    pixel.R(0, 0, cfg.WindowWidth, cfg.WindowHeight),
		Resizable: true,
	}
	if p.win, err = pixelgl.NewWindow(winCfg); err != nil {
		return nil, err
	}
	p.win.SetSmooth(true)
	p.win.SetCursorVisible(false)
	// Load sprite
	if err := animation.LoadAllSprite(); err != nil {
		return nil, err
	}
	// Create network
	p.client = NewClientNetwork()
	if err = p.client.Start(); err != nil {
		return nil, err
	}
	// Create world
	p.world = world.New(p)
	// Register player
	if err := p.registerPlayer(); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *clientProcessor) Restart() {
	fmt.Println(config.GetConfig())
	p.restartSig <- false
}

func (p *clientProcessor) Close() {
	p.restartSig <- true
}

func (p *clientProcessor) GetWindow() *pixelgl.Window {
	return p.win
}

func (p *clientProcessor) Run() {
	for p.restartCount = 0; ; p.restartCount++ {
		time.Sleep(time.Millisecond * 10)
		go p.startUpdateLoop(p.restartCount)
		go p.startRenderLoop(p.restartCount)
		go p.consumeWorldSnapshot()
		go p.produceInputSnapshot()
		if <-p.restartSig {
			return
		}
	}
}

func (p *clientProcessor) startUpdateLoop(restartCount int) {
	ctx := context.Background()
	limiter := rate.NewLimiter(rate.Limit(config.ClientInputRate), 1)
	for {
		_ = limiter.Wait(ctx)
		if p.win.Closed() {
			p.Close()
		}
		if restartCount != p.restartCount {
			return
		}
		p.world.ClientUpdate()
		p.win.UpdateInput()
	}
}

func (p *clientProcessor) startRenderLoop(restartCount int) {
	cfg := config.GetConfig()
	ctx := context.Background()
	limiter := rate.NewLimiter(rate.Limit(cfg.RefreshRate), 1)
	for {
		_ = limiter.Wait(ctx)
		if restartCount != p.restartCount {
			return
		}
		p.world.Render()
		p.win.Update()
	}
}

func (p *clientProcessor) consumeWorldSnapshot() {
	for cmdData := range p.client.Listen() {
		switch cmdData.Cmd {
		case protocol.CmdAddWorldSnapshot:
			data := cmdData.Data.(*protocol.AddWorldSnapshotRequest)
			p.world.SetSnapshot(data.Tick, data.WorldSnapshot)
		}
	}
}

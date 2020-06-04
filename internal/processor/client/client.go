package client

import (
	"errors"
	"fmt"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/animation"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/common"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/config"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/menu"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/protocol"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/sound"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/world"
	"github.com/mr-panta/go-logger"
	"golang.org/x/image/colornames"
)

type clientProcessor struct {
	win          *pixelgl.Window
	restartSig   chan bool
	restartCount int
	menu         common.Menu
	world        common.World
	client       ClientNetwork
	started      bool
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
	// Load sprite
	if err := animation.LoadAllSprites(); err != nil {
		return nil, err
	}
	// Load sound
	if err := sound.LoadAllSounds(); err != nil {
		return nil, err
	}
	// Create menu
	p.menu = menu.New(p)
	return p, nil
}

func (p *clientProcessor) ToggleFPSLimit() {
	cfg := config.GetConfig()
	if cfg.RefreshRate == config.MaxRefreshRate {
		cfg.RefreshRate = config.DefaultRefreshRate
	} else {
		cfg.RefreshRate = config.MaxRefreshRate
	}
	p.Restart()
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
		if <-p.restartSig {
			return
		}
	}
}

func (p *clientProcessor) StartWorld(hostIP, playerName string) (err error) {
	// Create network
	p.client = NewClientNetwork(hostIP)
	if err = p.client.Start(); err != nil {
		logger.Debugf(nil, err.Error())
		return errors.New("CAN'T CONNECT TO HOST")
	}
	// Create world
	p.world = world.New(p)
	// Register player
	if err := p.registerPlayer(playerName); err != nil {
		return err
	}
	p.started = true
	p.win.SetSmooth(true)
	p.win.SetCursorVisible(false)
	go p.consumeWorldSnapshot()
	go p.produceInputSnapshot()
	return nil
}

func (p *clientProcessor) startUpdateLoop(restartCount int) {
	ticker := time.NewTicker(time.Second / config.ClientInputRate)
	for range ticker.C {
		if p.win.Closed() || restartCount != p.restartCount {
			return
		}
		if p.started && p.world != nil {
			p.world.ClientUpdate()
			p.win.UpdateInput()
		}
	}
}

func (p *clientProcessor) startRenderLoop(restartCount int) {
	cfg := config.GetConfig()
	ticker := time.NewTicker(time.Second / time.Duration(cfg.RefreshRate))
	for range ticker.C {
		if p.win.Closed() {
			p.Close()
			return
		}
		if restartCount != p.restartCount {
			return
		}
		p.win.Clear(colornames.Black)
		if p.menu != nil {
			p.menu.UpdateAndRender()
		}
		if p.started && p.world != nil {
			p.world.Render()
		}
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

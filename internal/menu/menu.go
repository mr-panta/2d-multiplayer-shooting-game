package menu

import (
	"fmt"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/common"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/config"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

type Menu struct {
	win             *pixelgl.Window
	clientProcessor common.ClientProcessor
	hostAddrInput   *Input
	playerNameInput *Input
	playButton      *Button
	message         string
}

func New(clientProcessor common.ClientProcessor) *Menu {
	win := clientProcessor.GetWindow()
	return &Menu{
		win:             win,
		clientProcessor: clientProcessor,
		hostAddrInput:   NewInput(win, config.TCPIP),
		playerNameInput: NewInput(win, ""),
		playButton:      NewButton(win),
	}
}

func (m *Menu) UpdateAndRender() {
	m.update()
	m.render()
}

func (m *Menu) render() {
	m.hostAddrInput.Render()
	m.playerNameInput.Render()
	m.playButton.Render()
	m.renderMessage()
}

func (m *Menu) update() {
	m.updateHostAddrInput()
	m.updatePlayerNameInput()
	m.updatePlayerButton()
}

func (m *Menu) updateHostAddrInput() {
	m.hostAddrInput.Width = 360
	m.hostAddrInput.Height = 36
	m.hostAddrInput.Thickness = 2
	m.hostAddrInput.Size = 2
	m.hostAddrInput.Color = colornames.White
	m.hostAddrInput.HoverColor = colornames.Yellow
	m.hostAddrInput.FocusColor = colornames.Red
	m.hostAddrInput.Label = "HOST ADDRESS"
	m.hostAddrInput.Pos = m.win.Bounds().Center().Sub(pixel.V(m.hostAddrInput.Width/2, -60))
	m.hostAddrInput.Update()
}

func (m *Menu) updatePlayerNameInput() {
	m.playerNameInput.Width = 360
	m.playerNameInput.Height = 36
	m.playerNameInput.Thickness = 2
	m.playerNameInput.Size = 2
	m.playerNameInput.Color = colornames.White
	m.playerNameInput.HoverColor = colornames.Yellow
	m.playerNameInput.FocusColor = colornames.Red
	m.playerNameInput.Label = "PLAYER NAME"
	m.playerNameInput.Pos = m.win.Bounds().Center().Sub(pixel.V(m.playerNameInput.Width/2, 0))
	m.playerNameInput.Update()
}

func (m *Menu) updatePlayerButton() {
	m.playButton.Width = 120
	m.playButton.Height = 40
	m.playButton.Thickness = 2
	m.playButton.Size = 2
	m.playButton.Color = colornames.White
	m.playButton.HoverColor = colornames.Yellow
	m.playButton.FocusColor = colornames.Red
	m.playButton.Label = "PLAY"
	m.playButton.Pos = m.win.Bounds().Center().Sub(pixel.V(m.playButton.Width/2, 64))
	m.playButton.Update()
	if m.playButton.Actived() {
		err := m.clientProcessor.StartWorld(
			m.hostAddrInput.GetValue(),
			m.playerNameInput.GetValue(),
		)
		if err != nil {
			m.message = err.Error()
		}
	}
}

func (m *Menu) renderMessage() {
	atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	txt := text.New(pixel.ZV, atlas)
	txt.Clear()
	txt.LineHeight = atlas.LineHeight()
	txt.Color = colornames.Red
	fmt.Fprintf(txt, m.message)
	txt.Draw(m.win, pixel.IM.
		Moved(m.win.Bounds().Center().Sub(pixel.V(txt.Bounds().W()/2, 88))),
	)
}

package config

import (
	"image/color"
	"time"

	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

const (
	DefaultRefreshRate  = float64(60)
	DefaultWindowWidth  = float64(768)
	DefaultWindowHeight = float64(512)
)

const (
	Timestep        = 15 * time.Millisecond
	LerpPeriod      = 67 * time.Millisecond
	ServerSyncRate  = 30
	ClientSyncRate  = 30
	ClientInputRate = 256
	Title           = "Sirimongkol Project V2"
	IDLength        = 8
	RespawnTime     = 3 * time.Second
	PlayerTimeOut   = 10 * time.Second
	BufferSize      = 1000
)

// network
const (
	TCPIP = "34.87.23.164"
	// TCPIP    = ""
	TCPPortA = ":4999"
	TCPPortB = ":4998"
)

// key
const (
	FireKey   = pixelgl.MouseButton1
	FocusKey  = pixelgl.MouseButton2
	UpKey     = pixelgl.KeyW
	LeftKey   = pixelgl.KeyA
	DownKey   = pixelgl.KeyS
	RightKey  = pixelgl.KeyD
	ReloadKey = pixelgl.KeyR
	DropKey   = pixelgl.KeyG
)

// color
var (
	LerpColor         = color.RGBA{0x00, 0x00, 0xff, 72}
	LashSnapshotColor = color.RGBA{0x00, 0x80, 0x00, 72}
	ColliderColor     = colornames.Red
	ShapeColor        = colornames.Blue
)

// object type
const (
	PlayerObject = 1
	ItemObject   = 2
	WeaponObject = 3
	BulletObject = 4
)
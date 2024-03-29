package config

import (
	"image/color"
	"time"

	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

const Version = "0.2.0"

// default
const (
	DefaultRefreshRate  = 60
	MaxRefreshRate      = 300
	DefaultWindowWidth  = float64(768)
	DefaultWindowHeight = float64(512)
)

// settings
const (
	Timestep        = 15 * time.Millisecond
	LerpPeriod      = 100 * time.Millisecond
	ServerSyncRate  = 30
	ClientSyncRate  = 30
	ClientInputRate = 256
	Title           = "Sirimongkol Project V2"
	IDLength        = 8
	RespawnTime     = 3 * time.Second
	PlayerTimeOut   = 10 * time.Second
	BufferSize      = 1000
	LogFile         = "data.log"
)

// world
const (
	MinWindowRenderZ     = 1000
	DefaultWorldInitTime = 60 * time.Second
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
	FireKey          = pixelgl.MouseButton1
	MeleeKey         = pixelgl.MouseButton2
	UpKey            = pixelgl.KeyW
	LeftKey          = pixelgl.KeyA
	DownKey          = pixelgl.KeyS
	RightKey         = pixelgl.KeyD
	ReloadKey        = pixelgl.KeyR
	DropKey          = pixelgl.KeyG
	Use1stItemKey    = pixelgl.Key1
	Use2ndItemKey    = pixelgl.Key2
	Use3rdItemKey    = pixelgl.Key3
	ToggleMuteKey    = pixelgl.KeyM
	VolumeUpKey      = pixelgl.KeyUp
	VolumeDownKey    = pixelgl.KeyDown
	ToggleFullScreen = pixelgl.KeyF10
	ToggleFPSLimit   = pixelgl.KeyF9
)

// color
var (
	LerpColor         = color.RGBA{0x00, 0x00, 0xff, 72}
	LashSnapshotColor = color.RGBA{0x00, 0x80, 0x00, 72}
	ColliderColor     = colornames.Red
	ShapeColor        = colornames.Blue
)

// world type
const (
	DefaultWorld = 1
)

// object type
const (
	PlayerObject   = 1
	ItemObject     = 2
	WeaponObject   = 3
	BulletObject   = 4
	TreeObject     = 5
	TerrainObject  = 6
	BoundaryObject = 7
)

// weapon type
const (
	M4Weapon      = 1
	ShotgunWeapon = 2
	SniperWeapon  = 3
	PistolWeapon  = 4
	SMGWeapon     = 5
	KnifeWeapon   = 6
)

// tree type
const (
	TreeTypeA = "A"
	TreeTypeB = "B"
	TreeTypeC = "C"
	TreeTypeD = "D"
	TreeTypeE = "E"
)

var TreeTypes = []string{
	TreeTypeA,
	TreeTypeB,
	TreeTypeC,
	TreeTypeD,
	TreeTypeE,
}

// terrain
const TerrainTypeAmount = 5

// item type
const (
	InstanceUsedItem = 1
	CollectibleItem  = 2
)

package weapon

import (
	"image/color"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/common"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/config"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/protocol"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/ticktime"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/util"
	"golang.org/x/image/colornames"
)

const (
	bulletThickness = 6.
	bulletZ         = 10
)

var (
	bulletAColor = color.RGBA{0xff, 0xba, 0x00, 0xff}
	bulletBColor = colornames.White
)

type Bullet struct {
	world      common.World
	imd        *imdraw.IMDraw
	initPosImd *imdraw.IMDraw
	// state fields
	id         string
	playerID   string
	weaponID   string
	initPos    pixel.Vec
	dir        pixel.Vec
	speed      float64
	maxRange   float64
	damage     float64
	length     float64
	fireTime   time.Time
	deleteTime time.Time
	// calculated fields
	pos         pixel.Vec
	prevPos     pixel.Vec
	isDestroyed bool
}

func NewBullet(world common.World, id string) *Bullet {
	return &Bullet{
		world:       world,
		imd:         imdraw.New(nil),
		initPosImd:  imdraw.New(nil),
		id:          id,
		isDestroyed: true,
	}
}

func (o *Bullet) GetID() string {
	return o.id
}

func (o *Bullet) GetType() int {
	return config.BulletObject
}

func (o *Bullet) Destroy() {
	o.isDestroyed = true
}

func (o *Bullet) Exists() bool {
	return !o.isDestroyed
}

func (o *Bullet) GetShape() pixel.Rect {
	size := bulletThickness
	if o.length > size {
		size = o.length
	}
	return pixel.R(
		o.pos.X-size/2,
		o.pos.Y-size/2,
		o.pos.X+size/2,
		o.pos.Y-size/2,
	)
}

func (o *Bullet) GetCollider() (pixel.Rect, bool) {
	return pixel.ZR, false
}

func (o *Bullet) GetRenderObjects() []common.RenderObject {
	return []common.RenderObject{common.NewRenderObject(bulletZ, o.GetShape(), o.render)}
}

func (o *Bullet) GetSnapshot(tick int64) *protocol.ObjectSnapshot {
	return &protocol.ObjectSnapshot{
		ID:   o.GetID(),
		Type: o.GetType(),
		Bullet: &protocol.BulletSnapshot{
			PlayerID:   o.playerID,
			WeaponID:   o.weaponID,
			InitPos:    util.ConvertVec(o.initPos),
			Dir:        util.ConvertVec(o.dir),
			Speed:      o.speed,
			MaxRange:   o.maxRange,
			Damage:     o.damage,
			Length:     o.length,
			FireTime:   o.fireTime.UnixNano(),
			DeleteTime: o.deleteTime.UnixNano(),
		},
	}
}

func (o *Bullet) SetSnapshot(tick int64, snapshot *protocol.ObjectSnapshot) {
	if snapshot != nil && snapshot.Bullet != nil {
		ss := snapshot.Bullet
		o.playerID = ss.PlayerID
		o.weaponID = ss.WeaponID
		o.initPos = ss.InitPos.Convert()
		o.dir = ss.Dir.Convert()
		o.speed = ss.Speed
		o.maxRange = ss.MaxRange
		o.damage = ss.Damage
		o.length = ss.Length
		o.fireTime = time.Unix(0, ss.FireTime)
		o.deleteTime = time.Unix(0, ss.DeleteTime)
		o.isDestroyed = false
	}
}

func (o *Bullet) ServerUpdate(tick int64) {
	now := ticktime.GetServerTime()
	if !ticktime.IsZeroTime(o.deleteTime) {
		if now.Sub(o.deleteTime) > config.LerpPeriod*2 {
			o.world.GetObjectDB().Delete(o.id)
		}
	} else {
		o.calculatePosByTime(ticktime.GetServerTime())
		if obj := o.checkObjectCollision(); obj != nil && obj.GetID() != o.playerID {
			o.deleteTime = now
			if obj.GetType() == config.PlayerObject {
				player := obj.(common.Player)
				player.AddDamage(o.playerID, o.weaponID, o.damage)
			}
		}
	}
}

func (o *Bullet) ClientUpdate() {
	isDestroyed := o.isDestroyed
	if o.playerID == o.world.GetMainPlayerID() {
		if !ticktime.IsZeroTime(o.deleteTime) {
			isDestroyed = true
		} else {
			o.calculatePosByTime(ticktime.GetServerTime())
		}
	} else if ticktime.GetLerpTime().After(o.fireTime) {
		if !ticktime.IsZeroTime(o.deleteTime) &&
			ticktime.GetLerpTime().After(o.deleteTime) {
			isDestroyed = true
		} else {
			o.calculatePosByTime(ticktime.GetLerpTime())
			isDestroyed = false
		}
	} else {
		isDestroyed = true
	}
	o.isDestroyed = isDestroyed
}

func (o *Bullet) Fire(playerID, weaponID string, initPos, dir pixel.Vec, speed, maxRange, damage, length float64) {
	o.playerID = playerID
	o.weaponID = weaponID
	o.initPos = initPos
	o.dir = dir
	o.speed = speed
	o.maxRange = maxRange
	o.damage = damage
	o.length = length
	o.fireTime = ticktime.GetServerTime()
	o.isDestroyed = false
}

func (o *Bullet) render(target pixel.Target, viewPos pixel.Vec) {
	a := pixel.ZV.Add(pixel.V(o.length/2, 0))
	b := pixel.ZV.Sub(pixel.V(o.length/2, 0))
	matrix := pixel.IM.Rotated(pixel.ZV, o.dir.Angle())
	matrix = matrix.Moved(o.pos)
	matrix = matrix.Moved(pixel.ZV.Sub(viewPos))
	o.imd.Clear()
	o.imd.Color = bulletAColor
	o.imd.Push(a)
	o.imd.Color = bulletBColor
	o.imd.Push(b)
	o.imd.Line(bulletThickness)
	o.imd.SetMatrix(matrix)
	o.imd.Draw(target)
	// debug
	if config.EnvDebug() {
		o.renderInitPos(target, viewPos)
	}
}

func (o *Bullet) renderInitPos(target pixel.Target, viewPos pixel.Vec) { // For debugging
	o.initPosImd.Clear()
	o.initPosImd.Color = colornames.Red
	o.initPosImd.Push(o.initPos)
	o.initPosImd.Circle(2, 1)
	o.initPosImd.SetMatrix(pixel.IM.Moved(pixel.ZV.Sub(viewPos)))
	o.initPosImd.Draw(target)
}

func (o *Bullet) getColliderByPos(pos pixel.Vec) pixel.Rect {
	return pixel.Rect{
		Min: pos.Sub(pixel.V(1, 1)),
		Max: pos.Add(pixel.V(1, 1)),
	}
}

func (o *Bullet) calculatePosByTime(t time.Time) {
	prevPos := o.pos
	diff := t.Sub(o.fireTime)
	dist := o.speed * diff.Seconds()
	if dist > o.maxRange {
		o.deleteTime = ticktime.GetServerTime()
	} else {
		o.pos = o.initPos.Add(o.dir.Unit().Scaled(dist))
	}
	o.prevPos = prevPos
}

func (o *Bullet) checkObjectCollision() common.Object {
	prevCollider := o.getColliderByPos(o.prevPos)
	currCollider := o.getColliderByPos(o.pos)
	for _, obj := range o.world.GetObjectDB().SelectAll() {
		if !obj.Exists() || obj.GetID() == o.id {
			continue
		}
		if obj.GetType() == config.PlayerObject {
			staticAdjust, _ := util.CheckCollision(
				obj.GetShape(),
				prevCollider,
				currCollider,
			)
			if staticAdjust.Len() > 0 {
				return obj
			}
		}
		if collider, exists := obj.GetCollider(); exists {
			staticAdjust, _ := util.CheckCollision(
				collider,
				prevCollider,
				currCollider,
			)
			if staticAdjust.Len() > 0 {
				return obj
			}
		}
	}
	return nil
}

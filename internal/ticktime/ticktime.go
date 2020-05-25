package ticktime

import (
	"time"

	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/config"
)

var (
	startTime time.Time
	diff      time.Duration
	ping      time.Duration
)

func SetServerTime(t time.Time, p time.Duration) {
	diff = time.Until(t) + p/2
}

func SetServerStartTime(t time.Time) {
	startTime = t
}

func SetPing(d time.Duration) {
	ping = d
}

func GetServerTime() time.Time {
	return time.Now().Add(diff)
}

func GetLerpTime() time.Time {
	return GetServerTime().Add(-config.LerpPeriod)
}

func GetServerStartTime() time.Time {
	return startTime
}

func GetTick(t time.Time) int64 {
	d := t.Sub(startTime)
	return d.Nanoseconds() / config.Timestep.Nanoseconds()
}

func GetTickTime(tick int64) time.Time {
	d := time.Duration((tick * config.Timestep.Nanoseconds()))
	return startTime.Add(d)
}

func GetPing() time.Duration {
	return ping
}

func IsZeroTime(t time.Time) bool {
	return t.Before(startTime)
}

func GetServerTimeMS() int64 {
	return GetServerTime().UnixNano() / 1000000
}

package protocol

import (
	"time"

	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/config"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/ticktime"
)

type TickSnapshot struct {
	Tick     int64
	Snapshot *ObjectSnapshot
}

func GetSnapshotByTime(t time.Time, tickSnapshots []*TickSnapshot) (
	ssA, ssB *ObjectSnapshot, d float64) {
	var tickA, tickB int64
	tick := ticktime.GetTick(t)
	if len(tickSnapshots) == 0 {
		return nil, nil, 0
	} else if ts := tickSnapshots[len(tickSnapshots)-1]; ts.Tick <= tick {
		tickA = ts.Tick
		tickB = ts.Tick
		ssA = ts.Snapshot
		ssB = ts.Snapshot
	} else {
		for i := len(tickSnapshots) - 1; i > 0; i-- {
			tsA := tickSnapshots[i-1]
			tsB := tickSnapshots[i]
			if tsA.Tick <= tick && tick < tsB.Tick {
				tickA = tsA.Tick
				tickB = tsB.Tick
				ssA = tsA.Snapshot
				ssB = tsB.Snapshot
				break
			}
		}
	}
	if ssA == nil || ssB == nil {
		return nil, nil, 0
	}
	if tickA != tickB {
		x := float64(t.Sub(ticktime.GetTickTime(tickA)).Nanoseconds())
		y := float64((tickB - tickA) * config.Timestep.Nanoseconds())
		d = x / y
	}
	return ssA, ssB, d
}

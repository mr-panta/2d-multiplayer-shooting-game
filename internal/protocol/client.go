package protocol

// AddWorldSnapshot

type AddWorldSnapshotRequest struct {
	Tick          int64          `json:"tick,omitempty"`
	WorldSnapshot *WorldSnapshot `json:"world_snapshot,omitempty"`
}

type AddWorldSnapshotResponse struct{}

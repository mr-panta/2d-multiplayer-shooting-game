package protocol

// RegisterPlayer

type RegisterPlayerRequest struct {
	PlayerName string `json:"player_name,omitempty"`
	Version    string `json:"version,omitempty"`
}

type RegisterPlayerResponse struct {
	// Debug
	OK           bool   `json:"ok,omitempty"`
	DebugMessage string `json:"debug_message,omitempty"`
	// Info
	PlayerID      string         `json:"player_id,omitempty"`
	ServerTime    int64          `json:"server_time,omitempty"`
	StartTime     int64          `json:"start_time,omitempty"`
	Tick          int64          `json:"tick,omitempty"`
	WorldSnapshot *WorldSnapshot `json:"world_snapshot,omitempty"`
}

// SetPlayerInput

type SetPlayerInputRequest struct {
	PlayerID      string         `json:"player_id,omitempty"`
	InputSnapshot *InputSnapshot `json:"input_snapshot,omitempty"`
}

type SetPlayerInputResponse struct{}

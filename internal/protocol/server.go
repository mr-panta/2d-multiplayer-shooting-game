package protocol

// RegisterPlayer

type RegisterPlayerRequest struct{}

type RegisterPlayerResponse struct {
	PlayerID   string `json:"player_id,omitempty"`
	ServerTime int64  `json:"server_time,omitempty"`
	StartTime  int64  `json:"start_time,omitempty"`
}

// SetPlayerInput

type SetPlayerInputRequest struct {
	PlayerID      string         `json:"player_id,omitempty"`
	InputSnapshot *InputSnapshot `json:"input_snapshot,omitempty"`
}

type SetPlayerInputResponse struct{}

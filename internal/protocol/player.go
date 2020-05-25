package protocol

type PlayerSnapshot struct {
	PlayerName   string  `json:"player_name,omitempty"`
	WeaponID     string  `json:"weapon_id,omitempty"`
	CursorDir    *Vec    `json:"cursor_dir,omitempty"`
	Pos          *Vec    `json:"pos,omitempty"`
	MoveDir      *Vec    `json:"move_dir,omitempty"`
	MoveSpeed    float64 `json:"move_speed,omitempty"`
	MaxMoveSpeed float64 `json:"max_move_speed,omitempty"`
	HP           float64 `json:"hp,omitempty"`
	RespawnTime  int64   `json:"respawn_time,omitempty"`
	HitTime      int64   `json:"hit_time,omitempty"`
	TriggerTime  int64   `json:"trigger_time,omitempty"`
}

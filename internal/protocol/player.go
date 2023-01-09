package protocol

type PlayerSnapshot struct {
	PlayerName       string   `json:"player_name,omitempty"`
	MeleeWeaponID    string   `json:"melee_weapon_id,omitempty"`
	WeaponID         string   `json:"weapon_id,omitempty"`
	ItemIDs          []string `json:"item_ids,omitempty"`
	Kill             int      `json:"kill,omitempty"`
	Death            int      `json:"death,omitempty"`
	Streak           int      `json:"streak,omitempty"`
	MaxStreak        int      `json:"max_streak,omitempty"`
	CursorDir        *Vec     `json:"cursor_dir,omitempty"`
	Pos              *Vec     `json:"pos,omitempty"`
	MoveDir          *Vec     `json:"move_dir,omitempty"`
	MoveSpeed        float64  `json:"move_speed,omitempty"`
	MaxMoveSpeed     float64  `json:"max_move_speed,omitempty"`
	HP               float64  `json:"hp,omitempty"`
	Armor            float64  `json:"armor,omitempty"`
	RespawnTime      int64    `json:"respawn_time,omitempty"`
	HitTime          int64    `json:"hit_time,omitempty"`
	MeleeTime        int64    `json:"melee_time,omitempty"`
	TriggerTime      int64    `json:"trigger_time,omitempty"`
	PickupTime       int64    `json:"pickup_time,omitempty"`
	HitVisibleMS     int      `json:"hit_visible_ms,omitempty"`
	TriggerVisibleMS int      `json:"trigger_visible_ms,omitempty"`
	IsInvulnerable   bool     `json:"is_invulnerable,omitempty"`
	IsVisible        bool     `json:"is_visible,omitempty"`
}

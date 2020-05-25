package protocol

// Bullet

type BulletSnapshot struct {
	PlayerID   string  `json:"player_id,omitempty"`
	WeaponID   string  `json:"weapon_id,omitempty"`
	InitPos    *Vec    `json:"init_pos,omitempty"`
	Dir        *Vec    `json:"dir,omitempty"`
	Speed      float64 `json:"speed,omitempty"`
	MaxRange   float64 `json:"max_range,omitempty"`
	Damage     float64 `json:"damage,omitempty"`
	Length     float64 `json:"length,omitempty"`
	FireTime   int64   `json:"fire_time,omitemptye"`
	DeleteTime int64   `json:"delete_time,omitemptye"`
}

// Weapon

type WeaponSnapshot struct {
	M4 *WeaponM4Snapshot `json:"m4,omitempty"`
}

type WeaponM4Snapshot struct {
	PlayerID    string `json:"player_id,omitempty"`
	Mag         int    `json:"mag,omitempty"`
	Ammo        int    `json:"ammo,omitempty"`
	TriggerTime int64  `json:"trigger_time,omitempty"`
	ReloadTime  int64  `json:"reload_time,omitempty"`
}

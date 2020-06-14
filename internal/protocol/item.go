package protocol

type ItemSnapshot struct {
	Weapon   *ItemWeaponSnapshot   `json:"weapon,omitempty"`
	Ammo     *ItemAmmoSnapshot     `json:"ammo,omitempty"`
	AmmoSM   *ItemAmmoSMSnapshot   `json:"ammo_sm,omitempty"`
	Armor    *ItemArmorSnapshot    `json:"armor,omitempty"`
	Skull    *ItemSkullSnapshot    `json:"skull,omitempty"`
	LandMine *ItemLandMineSnapshot `json:"land_mine,omitempty"`
}

type ItemWeaponSnapshot struct {
	Pos      *Vec   `json:"pos,omitempty"`
	WeaponID string `json:"weapon_id,omitempty"`
}

type ItemAmmoSnapshot struct {
	Pos *Vec `json:"pos,omitempty"`
}

type ItemAmmoSMSnapshot struct {
	Pos *Vec `json:"pos,omitempty"`
}

type ItemArmorSnapshot struct {
	Pos   *Vec    `json:"pos,omitempty"`
	Armor float64 `json:"armor,omitempty"`
}

type ItemSkullRecord struct {
	RemainingMS int   `json:"remaining_ms,omitempty"`
	PickupTime  int64 `json:"pickup_time,omitempty"`
	DropTime    int64 `json:"drop_time,omitempty"`
}

type ItemSkullSnapshot struct {
	Pos       *Vec                        `json:"pos,omitempty"`
	PlayerID  string                      `json:"player_id,omitempty"`
	RecordMap map[string]*ItemSkullRecord `json:"item_skull_record,omitempty"`
}

type ItemLandMineSnapshot struct {
	Pos       *Vec   `json:"pos,omitempty"`
	PlayerID  string `json:"player_id,omitempty"`
	SlotIndex int    `json:"slot_index,omitempty"`
	IsActive  bool   `json:"is_active,omitempty"`
}

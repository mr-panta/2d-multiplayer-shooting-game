package protocol

type ItemSnapshot struct {
	Weapon *ItemWeaponSnapshot `json:"weapon,omitempty"`
	Ammo   *ItemAmmoSnapshot   `json:"ammo,omitempty"`
	AmmoSM *ItemAmmoSMSnapshot `json:"ammo_sm,omitempty"`
	Armor  *ItemArmorSnapshot  `json:"armor,omitempty"`
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

package protocol

type TreeSnapshot struct {
	Pos      *Vec   `json:"pos,omitempty"`
	TreeType string `json:"tree_type,omitempty"`
	Right    bool   `json:"right,omitempty"`
}

type TerrainSnapshot struct {
	Pos         *Vec `json:"pos,omitempty"`
	TerrainType int  `json:"terrain_type,omitempty"`
}

type BoundarySnapshot struct {
	Collider *Rect `json:"collider,omitempty"`
}

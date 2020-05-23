package protocol

type TreeSnapshot struct {
	Pos      *Vec   `json:"pos,omitempty"`
	TreeType string `json:"tree_type,omitempty"`
	Right    bool   `json:"right,omitempty"`
}

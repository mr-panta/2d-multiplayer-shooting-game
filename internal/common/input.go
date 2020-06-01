package common

import "github.com/faiface/pixel"

type RawInput struct {
	MousePos                pixel.Vec
	PressedFireKey          bool
	PressedFocusKey         bool
	PressedUpKey            bool
	PressedLeftKey          bool
	PressedDownKey          bool
	PressedRightKey         bool
	PressedReloadKey        bool
	PressedDropKey          bool
	PressedToggleMuteKey    bool
	PressedVolumeUpKey      bool
	PressedVolumeDownKey    bool
	PressedToggleFullScreen bool
}

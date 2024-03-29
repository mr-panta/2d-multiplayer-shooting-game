package common

import "github.com/faiface/pixel"

type RawInput struct {
	MousePos                pixel.Vec
	PressedFireKey          bool
	PressedMeleeKey         bool
	PressedUpKey            bool
	PressedLeftKey          bool
	PressedDownKey          bool
	PressedRightKey         bool
	PressedReloadKey        bool
	PressedDropKey          bool
	PressedUse1stItemKey    bool
	PressedUse2ndItemKey    bool
	PressedUse3rdItemKey    bool
	PressedToggleMuteKey    bool
	PressedVolumeUpKey      bool
	PressedVolumeDownKey    bool
	PressedToggleFullScreen bool
	PressedToggleFPSLimit   bool
}

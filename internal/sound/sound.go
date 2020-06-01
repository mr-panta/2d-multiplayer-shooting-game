package sound

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/config"
)

const (
	maxVolume  = 1.0
	minVolume  = -5.0
	volumeDiff = 0.5
)

var (
	sampleRate = beep.SampleRate(44100)
	mute       = false
	volume     = -2.0
)

type loadSoundFunc func(assetPath string) error

func LoadAllSounds() (err error) {
	err = speaker.Init(sampleRate, sampleRate.N(time.Second/20))
	if err != nil {
		return err
	}
	path := "./"
	if !config.EnvGorun() {
		if path, err = filepath.Abs(filepath.Dir(os.Args[0])); err != nil {
			return err
		}
	}
	assetPath := fmt.Sprintf("%s/asset/sound", path)
	fnList := []loadSoundFunc{
		loadCommonSounds,
		loadWeaponM4Sounds,
		loadWeaponShotgunSounds,
		loadWeaponSniperSounds,
		loadWeaponSMGSounds,
		loadWeaponPistolSounds,
	}
	for _, fn := range fnList {
		if err = fn(assetPath); err != nil {
			return err
		}
	}
	return nil
}

func ToggleMute() {
	mute = !mute
}

func VolumeUp() {
	v := volume + volumeDiff
	if v > maxVolume {
		v = maxVolume
	}
	volume = v
}

func VolumeDown() {
	v := volume - volumeDiff
	if v < minVolume {
		v = minVolume
	}
	volume = v
}

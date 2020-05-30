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

var (
	sampleRate = beep.SampleRate(44100)
)

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
	if err = loadCommonSounds(assetPath); err != nil {
		return err
	}
	if err = loadWeaponM4Sounds(assetPath); err != nil {
		return err
	}
	return nil
}
package sound

import (
	"os"

	"github.com/faiface/beep/effects"

	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

// buffer
var (
	weaponM4FireBuffer   *beep.Buffer
	weaponM4ReloadBuffer *beep.Buffer
)

func loadWeaponM4Sounds(assetPath string) (err error) {
	if err = loadWeaponM4FireSound(assetPath); err != nil {
		return err
	}
	if err = loadWeaponM4ReloadSound(assetPath); err != nil {
		return err
	}
	return nil
}

func loadWeaponM4FireSound(assetPath string) (err error) {
	file, err := os.Open(assetPath + "/weapon_m4/fire.mp3")
	if err != nil {
		return err
	}
	streamer, format, err := mp3.Decode(file)
	if err != nil {
		return err
	}
	defer streamer.Close()
	resampled := beep.Resample(4, format.SampleRate, sampleRate, streamer)
	weaponM4FireBuffer = beep.NewBuffer(format)
	weaponM4FireBuffer.Append(resampled)
	return nil
}

func loadWeaponM4ReloadSound(assetPath string) (err error) {
	file, err := os.Open(assetPath + "/weapon_m4/reload.mp3")
	if err != nil {
		return err
	}
	streamer, format, err := mp3.Decode(file)
	if err != nil {
		return err
	}
	defer streamer.Close()
	resampled := beep.Resample(4, format.SampleRate, sampleRate, streamer)
	weaponM4ReloadBuffer = beep.NewBuffer(format)
	weaponM4ReloadBuffer.Append(resampled)
	return nil
}

func PlayWeaponM4Fire(dist float64) {
	k := 1.0 / 500.0
	streamer := weaponM4FireBuffer.Streamer(0, weaponM4FireBuffer.Len())
	speaker.Play(&effects.Volume{
		Streamer: streamer,
		Base:     2,
		Volume:   1 - dist*k,
	})
}

func PlayWeaponM4Reload(dist float64) {
	k := 1.0 / 100.0
	streamer := weaponM4ReloadBuffer.Streamer(0, weaponM4ReloadBuffer.Len())
	speaker.Play(&effects.Volume{
		Streamer: streamer,
		Base:     2,
		Volume:   dist * k,
	})
}

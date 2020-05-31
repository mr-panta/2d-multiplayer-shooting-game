package sound

import (
	"os"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

// buffer
var (
	weaponSMGFireBuffer   *beep.Buffer
	weaponSMGReloadBuffer *beep.Buffer
)

func loadWeaponSMGSounds(assetPath string) (err error) {
	if err = loadWeaponSMGFireSound(assetPath); err != nil {
		return err
	}
	if err = loadWeaponSMGReloadSound(assetPath); err != nil {
		return err
	}
	return nil
}

func loadWeaponSMGFireSound(assetPath string) (err error) {
	file, err := os.Open(assetPath + "/weapon_smg/fire.mp3")
	if err != nil {
		return err
	}
	streamer, format, err := mp3.Decode(file)
	if err != nil {
		return err
	}
	defer streamer.Close()
	resampled := beep.Resample(4, format.SampleRate, sampleRate, streamer)
	weaponSMGFireBuffer = beep.NewBuffer(format)
	weaponSMGFireBuffer.Append(resampled)
	return nil
}

func loadWeaponSMGReloadSound(assetPath string) (err error) {
	file, err := os.Open(assetPath + "/weapon_smg/reload.mp3")
	if err != nil {
		return err
	}
	streamer, format, err := mp3.Decode(file)
	if err != nil {
		return err
	}
	defer streamer.Close()
	resampled := beep.Resample(4, format.SampleRate, sampleRate, streamer)
	weaponSMGReloadBuffer = beep.NewBuffer(format)
	weaponSMGReloadBuffer.Append(resampled)
	return nil
}

func PlayWeaponSMGFire(dist float64) {
	k := 1.0 / 500.0
	streamer := weaponSMGFireBuffer.Streamer(0, weaponSMGFireBuffer.Len())
	speaker.Play(&effects.Volume{
		Streamer: streamer,
		Base:     2,
		Volume:   -dist * k,
	})
}

func PlayWeaponSMGReload(dist float64) {
	k := 1.0 / 100.0
	streamer := weaponSMGReloadBuffer.Streamer(0, weaponSMGReloadBuffer.Len())
	speaker.Play(&effects.Volume{
		Streamer: streamer,
		Base:     2,
		Volume:   -dist * k,
	})
}

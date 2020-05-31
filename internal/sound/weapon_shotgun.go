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
	weaponShotgunFireBuffer   *beep.Buffer
	weaponShotgunReloadBuffer *beep.Buffer
)

func loadWeaponShotgunSounds(assetPath string) (err error) {
	if err = loadWeaponShotgunFireSound(assetPath); err != nil {
		return err
	}
	if err = loadWeaponShotgunReloadSound(assetPath); err != nil {
		return err
	}
	return nil
}

func loadWeaponShotgunFireSound(assetPath string) (err error) {
	file, err := os.Open(assetPath + "/weapon_shotgun/fire.mp3")
	if err != nil {
		return err
	}
	streamer, format, err := mp3.Decode(file)
	if err != nil {
		return err
	}
	defer streamer.Close()
	resampled := beep.Resample(4, format.SampleRate, sampleRate, streamer)
	weaponShotgunFireBuffer = beep.NewBuffer(format)
	weaponShotgunFireBuffer.Append(resampled)
	return nil
}

func loadWeaponShotgunReloadSound(assetPath string) (err error) {
	file, err := os.Open(assetPath + "/weapon_shotgun/reload.mp3")
	if err != nil {
		return err
	}
	streamer, format, err := mp3.Decode(file)
	if err != nil {
		return err
	}
	defer streamer.Close()
	resampled := beep.Resample(4, format.SampleRate, sampleRate, streamer)
	weaponShotgunReloadBuffer = beep.NewBuffer(format)
	weaponShotgunReloadBuffer.Append(resampled)
	return nil
}

func PlayWeaponShotgunFire(dist float64) {
	k := 1.0 / 500.0
	streamer := weaponShotgunFireBuffer.Streamer(0, weaponShotgunFireBuffer.Len())
	speaker.Play(&effects.Volume{
		Streamer: streamer,
		Base:     2,
		Volume:   1 - dist*k,
	})
}

func PlayWeaponShotgunReload(dist float64) {
	k := 1.0 / 100.0
	streamer := weaponShotgunReloadBuffer.Streamer(0, weaponShotgunReloadBuffer.Len())
	speaker.Play(&effects.Volume{
		Streamer: streamer,
		Base:     2,
		Volume:   dist * k,
	})
}

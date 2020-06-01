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
	weaponPistolFireBuffer   *beep.Buffer
	weaponPistolReloadBuffer *beep.Buffer
)

func loadWeaponPistolSounds(assetPath string) (err error) {
	if err = loadWeaponPistolFireSound(assetPath); err != nil {
		return err
	}
	if err = loadWeaponPistolReloadSound(assetPath); err != nil {
		return err
	}
	return nil
}

func loadWeaponPistolFireSound(assetPath string) (err error) {
	file, err := os.Open(assetPath + "/weapon_pistol/fire.mp3")
	if err != nil {
		return err
	}
	streamer, format, err := mp3.Decode(file)
	if err != nil {
		return err
	}
	defer streamer.Close()
	resampled := beep.Resample(4, format.SampleRate, sampleRate, streamer)
	weaponPistolFireBuffer = beep.NewBuffer(format)
	weaponPistolFireBuffer.Append(resampled)
	return nil
}

func loadWeaponPistolReloadSound(assetPath string) (err error) {
	file, err := os.Open(assetPath + "/weapon_pistol/reload.mp3")
	if err != nil {
		return err
	}
	streamer, format, err := mp3.Decode(file)
	if err != nil {
		return err
	}
	defer streamer.Close()
	resampled := beep.Resample(4, format.SampleRate, sampleRate, streamer)
	weaponPistolReloadBuffer = beep.NewBuffer(format)
	weaponPistolReloadBuffer.Append(resampled)
	return nil
}

func PlayWeaponPistolFire(dist float64) {
	k := 1.0 / 500.0
	streamer := weaponPistolFireBuffer.Streamer(0, weaponPistolFireBuffer.Len())
	speaker.Play(&effects.Volume{
		Silent:   mute,
		Streamer: streamer,
		Base:     2,
		Volume:   volume - dist*k,
	})
}

func PlayWeaponPistolReload(dist float64) {
	k := 1.0 / 100.0
	streamer := weaponPistolReloadBuffer.Streamer(0, weaponPistolReloadBuffer.Len())
	speaker.Play(&effects.Volume{
		Silent:   mute,
		Streamer: streamer,
		Base:     2,
		Volume:   volume - dist*k,
	})
}

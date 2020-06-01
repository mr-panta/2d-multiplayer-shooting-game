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
	weaponSniperFireBuffer   *beep.Buffer
	weaponSniperReloadBuffer *beep.Buffer
)

func loadWeaponSniperSounds(assetPath string) (err error) {
	if err = loadWeaponSniperFireSound(assetPath); err != nil {
		return err
	}
	if err = loadWeaponSniperReloadSound(assetPath); err != nil {
		return err
	}
	return nil
}

func loadWeaponSniperFireSound(assetPath string) (err error) {
	file, err := os.Open(assetPath + "/weapon_sniper/fire.mp3")
	if err != nil {
		return err
	}
	streamer, format, err := mp3.Decode(file)
	if err != nil {
		return err
	}
	defer streamer.Close()
	resampled := beep.Resample(4, format.SampleRate, sampleRate, streamer)
	weaponSniperFireBuffer = beep.NewBuffer(format)
	weaponSniperFireBuffer.Append(resampled)
	return nil
}

func loadWeaponSniperReloadSound(assetPath string) (err error) {
	file, err := os.Open(assetPath + "/weapon_sniper/reload.mp3")
	if err != nil {
		return err
	}
	streamer, format, err := mp3.Decode(file)
	if err != nil {
		return err
	}
	defer streamer.Close()
	resampled := beep.Resample(4, format.SampleRate, sampleRate, streamer)
	weaponSniperReloadBuffer = beep.NewBuffer(format)
	weaponSniperReloadBuffer.Append(resampled)
	return nil
}

func PlayWeaponSniperFire(dist float64) {
	k := 1.0 / 500.0
	streamer := weaponSniperFireBuffer.Streamer(0, weaponSniperFireBuffer.Len())
	speaker.Play(&effects.Volume{
		Silent:   mute,
		Streamer: streamer,
		Base:     2,
		Volume:   volume - dist*k,
	})
}

func PlayWeaponSniperReload(dist float64) {
	k := 1.0 / 100.0
	streamer := weaponSniperReloadBuffer.Streamer(0, weaponSniperReloadBuffer.Len())
	speaker.Play(&effects.Volume{
		Silent:   mute,
		Streamer: streamer,
		Base:     2,
		Volume:   volume - dist*k,
	})
}

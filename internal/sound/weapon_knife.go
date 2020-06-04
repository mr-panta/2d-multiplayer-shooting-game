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
	weaponKnifeStabBuffer *beep.Buffer
)

func loadWeaponKnifeSounds(assetPath string) (err error) {
	if err = loadWeaponKnifeStabSound(assetPath); err != nil {
		return err
	}
	return nil
}

func loadWeaponKnifeStabSound(assetPath string) (err error) {
	file, err := os.Open(assetPath + "/weapon_knife/stab.mp3")
	if err != nil {
		return err
	}
	streamer, format, err := mp3.Decode(file)
	if err != nil {
		return err
	}
	defer streamer.Close()
	resampled := beep.Resample(4, format.SampleRate, sampleRate, streamer)
	weaponKnifeStabBuffer = beep.NewBuffer(format)
	weaponKnifeStabBuffer.Append(resampled)
	return nil
}

func PlayWeaponKnifeStab(dist float64) {
	k := 1.0 / 500.0
	streamer := weaponKnifeStabBuffer.Streamer(0, weaponKnifeStabBuffer.Len())
	speaker.Play(&effects.Volume{
		Silent:   mute,
		Streamer: streamer,
		Base:     2,
		Volume:   volume - dist*k,
	})
}

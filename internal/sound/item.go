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
	itemExplosionBuffer *beep.Buffer
)

func loadItemSounds(assetPath string) (err error) {
	if err = loadItemExplosionSound(assetPath); err != nil {
		return err
	}
	return nil
}

func loadItemExplosionSound(assetPath string) (err error) {
	file, err := os.Open(assetPath + "/item/explosion.mp3")
	if err != nil {
		return err
	}
	streamer, format, err := mp3.Decode(file)
	if err != nil {
		return err
	}
	defer streamer.Close()
	resampled := beep.Resample(4, format.SampleRate, sampleRate, streamer)
	itemExplosionBuffer = beep.NewBuffer(format)
	itemExplosionBuffer.Append(resampled)
	return nil
}

func PlayItemExplosion(dist float64) {
	k := 1.0 / 500.0
	streamer := itemExplosionBuffer.Streamer(0, itemExplosionBuffer.Len())
	speaker.Play(&effects.Volume{
		Silent:   mute,
		Streamer: streamer,
		Base:     2,
		Volume:   volume - dist*k,
	})
}

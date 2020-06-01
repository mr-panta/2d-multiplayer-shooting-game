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
	commonKillBuffer   *beep.Buffer
	commonPickupBuffer *beep.Buffer
)

func loadCommonSounds(assetPath string) (err error) {
	if err = loadCommonKillSound(assetPath); err != nil {
		return err
	}
	if err = loadCommonPickupSound(assetPath); err != nil {
		return err
	}
	return nil
}

func loadCommonKillSound(assetPath string) (err error) {
	file, err := os.Open(assetPath + "/common/kill.mp3")
	if err != nil {
		return err
	}
	streamer, format, err := mp3.Decode(file)
	if err != nil {
		return err
	}
	defer streamer.Close()
	resampled := beep.Resample(4, format.SampleRate, sampleRate, streamer)
	commonKillBuffer = beep.NewBuffer(format)
	commonKillBuffer.Append(resampled)
	return nil
}

func loadCommonPickupSound(assetPath string) (err error) {
	file, err := os.Open(assetPath + "/common/pickup.mp3")
	if err != nil {
		return err
	}
	streamer, format, err := mp3.Decode(file)
	if err != nil {
		return err
	}
	defer streamer.Close()
	resampled := beep.Resample(4, format.SampleRate, sampleRate, streamer)
	commonPickupBuffer = beep.NewBuffer(format)
	commonPickupBuffer.Append(resampled)
	return nil
}

func PlayCommonKill() {
	streamer := commonKillBuffer.Streamer(0, commonKillBuffer.Len())
	speaker.Play(&effects.Volume{
		Silent:   mute,
		Streamer: streamer,
		Base:     2,
		Volume:   volume + 0,
	})
}

func PlayCommonPickup() {
	streamer := commonPickupBuffer.Streamer(0, commonPickupBuffer.Len())
	speaker.Play(&effects.Volume{
		Silent:   mute,
		Streamer: streamer,
		Base:     2,
		Volume:   volume - 1,
	})
}

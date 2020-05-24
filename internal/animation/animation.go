package animation

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"time"

	"github.com/faiface/pixel"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/config"
)

var (
	objectSheet *pixel.PictureData
)

func LoadAllSprite() (err error) {
	path := "./"
	if !config.EnvGorun() {
		if path, err = filepath.Abs(filepath.Dir(os.Args[0])); err != nil {
			return err
		}
	}
	assetPath := fmt.Sprintf("%s/asset/sprite", path)
	if err := loadObjectSprite(assetPath); err != nil {
		return err
	}
	return nil
}

func timeMS() int64 {
	return time.Now().UnixNano() / 1000000
}

func loadObjectSprite(assetPath string) error {
	file, err := os.Open(fmt.Sprintf("%s/object.png", assetPath))
	if err != nil {
		return err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}
	objectSheet = pixel.PictureDataFromImage(img)
	return nil
}

func GetObjectSheet() *pixel.PictureData {
	return objectSheet
}

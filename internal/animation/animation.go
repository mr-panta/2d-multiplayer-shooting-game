package animation

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"

	"github.com/faiface/pixel"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/config"
)

var (
	objectSheet *pixel.PictureData
	shadowColor = color.RGBA{0, 0, 0, 88}
)

func LoadAllSprites() (err error) {
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

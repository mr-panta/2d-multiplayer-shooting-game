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
	weaponSheet    *pixel.PictureData
	characterSheet *pixel.PictureData
	itemSheet      *pixel.PictureData
	treeSheet      *pixel.PictureData
)

var FieldSheet *pixel.PictureData

func LoadAllSprite() (err error) {
	path := "./"
	if !config.EnvGorun() {
		if path, err = filepath.Abs(filepath.Dir(os.Args[0])); err != nil {
			return err
		}
	}
	assetPath := fmt.Sprintf("%s/asset/sprite", path)
	if err := loadCharacterSprite(assetPath); err != nil {
		return err
	}
	if err := loadWeaponSprite(assetPath); err != nil {
		return err
	}
	if err := loadItemSprite(assetPath); err != nil {
		return err
	}
	if err := loadTreeSprite(assetPath); err != nil {
		return err
	}
	if err := loadFieldSprite(assetPath); err != nil {
		return err
	}
	return nil
}

func timeMS() int64 {
	return time.Now().UnixNano() / 1000000
}

func loadWeaponSprite(assetPath string) error {
	file, err := os.Open(fmt.Sprintf("%s/weapon.png", assetPath))
	if err != nil {
		return err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}
	weaponSheet = pixel.PictureDataFromImage(img)
	return nil
}

func loadCharacterSprite(assetPath string) error {
	file, err := os.Open(fmt.Sprintf("%s/character.png", assetPath))
	if err != nil {
		return err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}
	characterSheet = pixel.PictureDataFromImage(img)
	return nil
}

func loadItemSprite(assetPath string) error {
	file, err := os.Open(fmt.Sprintf("%s/item.png", assetPath))
	if err != nil {
		return err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}
	itemSheet = pixel.PictureDataFromImage(img)
	return nil
}

func loadTreeSprite(assetPath string) error {
	file, err := os.Open(fmt.Sprintf("%s/tree.png", assetPath))
	if err != nil {
		return err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}
	treeSheet = pixel.PictureDataFromImage(img)
	return nil
}

func loadFieldSprite(assetPath string) error {
	file, err := os.Open(fmt.Sprintf("%s/field.png", assetPath))
	if err != nil {
		return err
	}
	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return err
	}
	FieldSheet = pixel.PictureDataFromImage(img)
	return nil
}

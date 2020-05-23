package util

import (
	"math/rand"
	"time"

	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/config"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func GenerateID() string {
	return RandString(config.IDLength)
}

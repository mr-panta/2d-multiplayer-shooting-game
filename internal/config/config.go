package config

import "os"

type Config struct {
	WindowWidth  float64
	WindowHeight float64
	RefreshRate  int64
}

var config *Config

func GetConfig() *Config {
	if config == nil {
		config = &Config{
			WindowWidth:  DefaultWindowWidth,
			WindowHeight: DefaultWindowHeight,
			RefreshRate:  DefaultRefreshRate,
		}
	}
	return config
}

func SetWindowWidth(width float64) {
	cfg := GetConfig()
	cfg.WindowWidth /= 2
}

func EnvGorun() bool {
	return os.Getenv("GORUN") != ""
}

func EnvDebug() bool {
	return os.Getenv("DEBUG") != ""
}

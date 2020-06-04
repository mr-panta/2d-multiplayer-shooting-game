package main

import (
	"context"

	"github.com/faiface/pixel/pixelgl"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/config"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/processor/client"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/util"
	"github.com/mr-panta/go-logger"
)

func run() {
	ctx := logger.GetContextWithLogID(context.Background(), "client_main")
	logPrinter, err := util.NewLogPrinter(config.LogFile)
	if err != nil {
		logger.Fatalf(ctx, err.Error())
	}
	logger.SetupLogger(logPrinter.Printf)
	p, err := client.NewClientProcessor()
	if err != nil {
		logger.Fatalf(ctx, err.Error())
	}
	p.Run()
}

func main() {
	pixelgl.Run(run)
}

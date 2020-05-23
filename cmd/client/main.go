package main

import (
	"context"

	"github.com/faiface/pixel/pixelgl"
	"github.com/mr-panta/go-logger"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/processor/client"
)

func run() {
	ctx := logger.GetContextWithLogID(context.Background(), "client_main")
	p, err := client.NewClientProcessor()
	if err != nil {
		logger.Fatalf(ctx, err.Error())
	}
	p.Run()
}

func main() {
	pixelgl.Run(run)
}

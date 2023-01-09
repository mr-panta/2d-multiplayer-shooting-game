package main

import (
	"context"

	"github.com/mr-panta/go-logger"
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/processor/server"
)

func main() {
	ctx := logger.GetContextWithLogID(context.Background(), "server_main")
	p, err := server.NewServerProcessor()
	if err != nil {
		logger.Fatalf(ctx, err.Error())
	}
	go p.CleanWorld()
	go p.UpdateWorld()
	go p.BroadcastSnapshot()
	p.Wait()
}

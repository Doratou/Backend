package main

import (
	"context"
	"log"

	"github.com/Doratou/Backend/bootstrap"
	runner "github.com/Doratou/Backend/cmd/worker"
	brokerInfra "github.com/Doratou/Backend/internal/infrastructure/broker/rabbitmq"
)

func main() {
	runner.Run("WS", func(ctx context.Context, consumer *brokerInfra.Consumer, cfg *bootstrap.Config) error {
		log.Print("[WS WORKER] WebSocket worker initialized and awaiting future hub integration.")
		<-ctx.Done()
		return nil
	})
}

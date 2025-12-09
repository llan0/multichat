package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/llan0/multichat/internal/adapters/kick"
	"github.com/llan0/multichat/internal/adapters/twitch"
	"github.com/llan0/multichat/internal/logger"
	"github.com/llan0/multichat/internal/models"
	"github.com/llan0/multichat/internal/service"
	"github.com/llan0/multichat/internal/ui"
	"go.uber.org/zap"
)

func main() {
	logger.Log.Info("starting multichat application")
	defer logger.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		logger.Log.Info("shutdown signal received")
		cancel()
	}()

	// mock producers
	var twitchProducer service.Producer = &twitchProducer{}
	var kickProducer service.Producer = &kickProducer{}

	logger.Log.Info("initializing producers", zap.Int("producer_count", 2))

	// mocks -> service -> UI
	ui.Start(ctx, twitchProducer, kickProducer)

	logger.Log.Info("application shutdown complete")
}

// twitch producer
type twitchProducer struct{}

func (t *twitchProducer) Stream(ctx context.Context) <-chan models.ChatMessage {
	return twitch.Stream(ctx)
}

// kick producer
type kickProducer struct{}

func (k *kickProducer) Stream(ctx context.Context) <-chan models.ChatMessage {
	return kick.Stream(ctx)
}

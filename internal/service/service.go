package service

import (
	"context"

	"github.com/llan0/go-chat/internal/logger"
	"github.com/llan0/go-chat/internal/models"
	"go.uber.org/zap"
)

// any chat producer that can stream messages
type Producer interface {
	Stream(ctx context.Context) <-chan models.ChatMessage
}

// merge twitch and kick chats using fanin pattern
func Merge(ctx context.Context, producers ...Producer) <-chan models.ChatMessage {
	merged := make(chan models.ChatMessage, 100)
	logger.Log.Info("merging producer streams", zap.Int("producer_count", len(producers)))

	for i, producer := range producers {
		go func(idx int, p Producer) {
			defer func() {
				logger.Log.Debug("producer goroutine exiting", zap.Int("producer_index", idx))
			}()
			for {
				select {
				case <-ctx.Done():
					logger.Log.Debug("producer context cancelled", zap.Int("producer_index", idx))
					return
				case msg, ok := <-p.Stream(ctx):
					if !ok {
						logger.Log.Info("producer channel closed", zap.Int("producer_index", idx))
						return
					}
					select {
					case merged <- msg:
						logger.Log.Debug("message merged",
							zap.String("platform", msg.Platform),
							zap.String("username", msg.Username),
							zap.Int("producer_index", idx),
						)
					case <-ctx.Done():
						return
					}
				}
			}
		}(i, producer)
	}

	// close merged channel when context is cancelled
	go func() {
		<-ctx.Done()
		logger.Log.Debug("closing merged channel")
		close(merged)
	}()

	return merged
}

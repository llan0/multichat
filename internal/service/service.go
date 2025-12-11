package service

import (
	"context"
	"sync"

	"github.com/llan0/multichat/internal/logger"
	"github.com/llan0/multichat/internal/models"
	"go.uber.org/zap"
)

// chat source that can stream messages
type Producer interface {
	Stream(ctx context.Context) <-chan models.ChatMessage
}

// fan in
func Merge(ctx context.Context, log logger.Logger, producers ...Producer) <-chan models.ChatMessage {
	merged := make(chan models.ChatMessage, 100)
	log.Info("merging producer streams", zap.Int("producer_count", len(producers)))

	var wg sync.WaitGroup

	for i, producer := range producers {
		wg.Add(1)

		// get the stream channel ONCE per producer
		stream := producer.Stream(ctx)

		go func(idx int, ch <-chan models.ChatMessage) {
			defer wg.Done()
			defer func() {
				log.Debug("producer goroutine exiting", zap.Int("producer_index", idx))
			}()

			for {
				select {
				case <-ctx.Done():
					log.Debug("producer context cancelled", zap.Int("producer_index", idx))
					return
				case msg, ok := <-ch:
					if !ok {
						log.Info("producer stream closed", zap.Int("producer_index", idx))
						return
					}
					select {
					case merged <- msg:
					case <-ctx.Done():
						return
					}
				}
			}
		}(i, stream)
	}

	// close merged channel when all producers are done
	go func() {
		wg.Wait()
		close(merged)
		log.Debug("all producers finished, merged channel closed")
	}()

	return merged
}

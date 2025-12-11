package adapters

import (
	"context"

	"github.com/llan0/multichat/internal/models"
)

type ChatAdapter interface {
	// returns a channel of chat messages. The channel is closed when the context is cancelled or an unrecoverable error occurs
	Stream(ctx context.Context) <-chan models.ChatMessage

	// gracefully shuts down the adapter
	Close() error

	// return platform name
	Platform() string
}

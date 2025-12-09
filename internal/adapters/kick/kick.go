package kick

import (
	"context"
	"math/rand"
	"time"

	"github.com/llan0/go-chat/internal/logger"
	"github.com/llan0/go-chat/internal/models"
	"go.uber.org/zap"
)

var (
	kickMessages = []string{
		"Let's go!",
		"W stream",
		"Fire emoji",
		"GOAT",
		"Absolute unit",
		"Banger",
		"Insane",
		"Respect",
		"Legend",
		"Based",
		"Facts",
		"Real",
		"Big W",
		"Clutch",
		"GG",
	}

	kickUsernames = []string{
		"KickUser1",
		"StreamWatcher",
		"ChatHero",
		"KickFan",
		"ViewerPro",
		"StreamEnjoyer",
		"KickLover",
		"ChatMaster",
		"KickWarrior",
		"StreamKing",
	}
)

// simulates kick web socket traffic for now
func Stream(ctx context.Context) <-chan models.ChatMessage {
	logger.Log.Info("starting Kick mock producer")
	out := make(chan models.ChatMessage, 100)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	go func() {
		defer func() {
			close(out)
			logger.Log.Info("Kick mock producer stopped")
		}()
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				logger.Log.Debug("Kick producer context cancelled")
				return
			case <-ticker.C:
				msg := models.ChatMessage{
					Platform:  "Kick",
					Username:  kickUsernames[rng.Intn(len(kickUsernames))],
					Content:   kickMessages[rng.Intn(len(kickMessages))],
					Color:     "#53FC18",
					Timestamp: time.Now(),
				}
				select {
				case out <- msg:
					logger.Log.Debug("Kick message emitted",
						zap.String("username", msg.Username),
						zap.String("content", msg.Content),
					)
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return out
}

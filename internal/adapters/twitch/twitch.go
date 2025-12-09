package twitch

import (
	"context"
	"math/rand"
	"time"

	"github.com/llan0/go-chat/internal/logger"
	"github.com/llan0/go-chat/internal/models"
	"go.uber.org/zap"
)

var (
	twitchMessages = []string{
		"PogChamp!",
		"KEKW",
		"monkaS",
		"LULW",
		"PepeHands",
		"OMEGALUL",
		"EZ Clap",
		"Pepega",
		"FeelsGoodMan",
		"TriHard",
		"HeyGuys",
		"Kappa",
		"ResidentSleeper",
		"BibleThump",
		"DansGame",
	}

	twitchUsernames = []string{
		"xQcFan123",
		"TwitchViewer",
		"ChatWarrior",
		"StreamLover",
		"PogChampUser",
		"KEKWMaster",
		"LULWKing",
		"PepeHandsUser",
		"EZClapPro",
		"TriHardGamer",
	}
)

// simulates twitch irc traffic for now
func Stream(ctx context.Context) <-chan models.ChatMessage {
	logger.Log.Info("starting Twitch mock producer")
	out := make(chan models.ChatMessage, 100)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	go func() {
		defer func() {
			close(out)
			logger.Log.Info("Twitch mock producer stopped")
		}()
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				logger.Log.Debug("Twitch producer context cancelled")
				return
			case <-ticker.C:
				msg := models.ChatMessage{
					Platform:  "Twitch",
					Username:  twitchUsernames[rng.Intn(len(twitchUsernames))],
					Content:   twitchMessages[rng.Intn(len(twitchMessages))],
					Color:     "#9146FF",
					Timestamp: time.Now(),
				}
				select {
				case out <- msg:
					logger.Log.Debug("Twitch message emitted",
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

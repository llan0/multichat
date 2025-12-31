package emotes

import (
	"fyne.io/fyne/v2"
	"github.com/llan0/multichat/internal/logger"
	"go.uber.org/zap"
)

// manage emote image caching
type Service struct {
	log   logger.Logger
	cache *ImageCache
}

func NewService(log logger.Logger) *Service {
	return &Service{
		log:   log.With(zap.String("component", "emotes")),
		cache: NewImageCache(),
	}
}

// GetImage returns a cached image resource for an emote, loading async if needed
// The onLoaded callback is called when an async load completes
func (s *Service) GetImage(emoteURL string, onLoaded func()) fyne.Resource {
	resource, loading := s.cache.GetOrLoad(emoteURL, emoteURL, func(r fyne.Resource) {
		if onLoaded != nil {
			onLoaded()
		}
	})

	if loading {
		return nil
	}

	return resource
}

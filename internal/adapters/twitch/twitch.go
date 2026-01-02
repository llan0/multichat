package twitch

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/gempir/go-twitch-irc/v4"
	"github.com/llan0/multichat/internal/emotes"
	"github.com/llan0/multichat/internal/logger"
	"github.com/llan0/multichat/internal/models"
	"go.uber.org/zap"
)

const (
	defaultColor   = "#9146FF"
	channelBufSize = 100
)

type Client struct {
	config Config
	log    logger.Logger

	mu     sync.Mutex
	conn   *twitch.Client
	closed bool
}

func NewClient(channel string, log logger.Logger) (*Client, error) {
	return NewClientWithConfig(DefaultConfig(channel), log)
}

func NewClientWithConfig(cfg Config, log logger.Logger) (*Client, error) {
	if cfg.Channel == "" {
		return nil, errors.New("channel is required")
	}

	return &Client{
		config: cfg,
		log:    log.With(zap.String("adapter", "twitch"), zap.String("channel", cfg.Channel)),
	}, nil
}

func (c *Client) Stream(ctx context.Context) <-chan models.ChatMessage {
	out := make(chan models.ChatMessage, channelBufSize)
	go c.streamLoop(ctx, out)
	return out
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.closed = true
	if c.conn != nil {
		err := c.conn.Disconnect()
		c.conn = nil
		return err
	}
	return nil
}

func (c *Client) Platform() string {
	return "Twitch"
}

func (c *Client) streamLoop(ctx context.Context, out chan<- models.ChatMessage) {
	defer close(out)

	delay := c.config.RetryConfig.InitialDelay

	for {
		if c.shouldStop(ctx) {
			return
		}

		if err := c.runSession(ctx, out); err != nil {
			c.log.Warn("session ended", zap.Error(err))
		}

		if c.shouldStop(ctx) {
			return
		}

		delay = c.backoff(ctx, delay)
		if delay == 0 {
			return
		}
	}
}

func (c *Client) shouldStop(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		c.log.Debug("context cancelled, stopping stream")
		return true
	default:
	}

	c.mu.Lock()
	closed := c.closed
	c.mu.Unlock()
	return closed
}

func (c *Client) runSession(ctx context.Context, out chan<- models.ChatMessage) error {
	conn := twitch.NewAnonymousClient()
	msgChan := make(chan models.ChatMessage, channelBufSize)
	disconnected := make(chan struct{})
	var connectErr error
	var wg sync.WaitGroup

	c.setConn(conn)
	c.registerHandlers(conn, msgChan, &connectErr)

	conn.Join(c.config.Channel)

	wg.Go(func() {
		c.connect(conn, disconnected)
	})

	c.forwardMessages(ctx, msgChan, out, disconnected)
	c.disconnect(msgChan)

	// wait for connect goroutine to finish
	wg.Wait()

	return connectErr
}

func (c *Client) setConn(conn *twitch.Client) {
	c.mu.Lock()
	c.conn = conn
	c.mu.Unlock()
}

// set up the IRC event handlers
func (c *Client) registerHandlers(conn *twitch.Client, msgChan chan<- models.ChatMessage, connectErr *error) {
	conn.OnPrivateMessage(func(msg twitch.PrivateMessage) {
		chatMsg := c.parseMessage(msg)
		select {
		case msgChan <- chatMsg:
		default:
			// drop message if buffer full??
		}
	})

	conn.OnConnect(func() {
		c.log.Info("connected")
	})

	conn.OnSelfJoinMessage(func(msg twitch.UserJoinMessage) {
		c.log.Info("joined channel", zap.String("channel", msg.Channel))
	})
}

func (c *Client) parseMessage(msg twitch.PrivateMessage) models.ChatMessage {
	color := msg.User.Color
	if color == "" {
		color = defaultColor
	}

	segments := c.buildSegments(msg.Message, msg.Emotes)

	return models.ChatMessage{
		Platform:  "Twitch",
		Username:  msg.User.DisplayName,
		Content:   msg.Message,
		Segments:  segments,
		Color:     color,
		Timestamp: msg.Time,
	}
}

// emotePosition represents a single emote occurrence in the message.
type emotePosition struct {
	id    string
	name  string
	start int
	end   int
}

// buildSegments parses the message content and Twitch emotes into segments.
func (c *Client) buildSegments(content string, ircEmotes []*twitch.Emote) []models.MessageSegment {
	if len(ircEmotes) == 0 {
		return []models.MessageSegment{{Type: models.SegmentText, Text: content}}
	}

	// Collect all emote positions
	var positions []emotePosition
	for _, e := range ircEmotes {
		for _, pos := range e.Positions {
			positions = append(positions, emotePosition{
				id:    e.ID,
				name:  e.Name,
				start: pos.Start,
				end:   pos.End,
			})
		}
	}

	// Sort by start position
	sort.Slice(positions, func(i, j int) bool {
		return positions[i].start < positions[j].start
	})

	// Build segments
	var segments []models.MessageSegment
	lastEnd := 0
	contentRunes := []rune(content)

	for _, pos := range positions {
		// Add text before this emote
		if pos.start > lastEnd {
			text := string(contentRunes[lastEnd:pos.start])
			if text != "" {
				segments = append(segments, models.MessageSegment{
					Type: models.SegmentText,
					Text: text,
				})
			}
		}

		// Add emote segment
		segments = append(segments, models.MessageSegment{
			Type:      models.SegmentEmote,
			EmoteID:   pos.id,
			EmoteURL:  emotes.BuildTwitchEmoteURL(pos.id),
			EmoteName: pos.name,
		})

		lastEnd = pos.end + 1
	}

	// Add remaining text after last emote
	if lastEnd < len(contentRunes) {
		text := string(contentRunes[lastEnd:])
		if text != "" {
			segments = append(segments, models.MessageSegment{
				Type: models.SegmentText,
				Text: text,
			})
		}
	}

	return segments
}

func (c *Client) connect(conn *twitch.Client, disconnected chan<- struct{}) {
	err := conn.Connect()
	if err != nil {
		c.log.Warn("disconnected", zap.Error(err))
	}
	close(disconnected)
}

// pipes messages from the internal channel to the output channel
func (c *Client) forwardMessages(ctx context.Context, in <-chan models.ChatMessage, out chan<- models.ChatMessage, disconnected <-chan struct{}) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-disconnected:
			return
		case msg, ok := <-in:
			if !ok {
				return
			}
			if !c.trySend(ctx, msg, out, disconnected) {
				return
			}
		}
	}
}

// attempt to send a message to the output channel
func (c *Client) trySend(ctx context.Context, msg models.ChatMessage, out chan<- models.ChatMessage, disconnected <-chan struct{}) bool {
	select {
	case out <- msg:
		c.log.Debug("message received", zap.String("username", msg.Username))
		return true
	case <-ctx.Done():
		return false
	case <-disconnected:
		return false
	}
}

func (c *Client) disconnect(msgChan chan models.ChatMessage) {
	c.mu.Lock()
	if c.conn != nil {
		c.conn.Disconnect()
		c.conn = nil
	}
	c.mu.Unlock()
	close(msgChan)
	c.log.Debug("disconnected")
}

// wait with exponential backoff before reconnecting
func (c *Client) backoff(ctx context.Context, delay time.Duration) time.Duration {
	c.log.Info("reconnecting", zap.Duration("delay", delay))

	select {
	case <-ctx.Done():
		return 0
	case <-time.After(delay):
	}

	next := time.Duration(float64(delay) * c.config.RetryConfig.Multiplier)
	next = max(next, c.config.RetryConfig.MaxDelay)
	return next
}

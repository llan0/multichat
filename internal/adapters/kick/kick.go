package kick

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/llan0/multichat/internal/adapters/common"
	"github.com/llan0/multichat/internal/logger"
	"github.com/llan0/multichat/internal/models"
	"go.uber.org/zap"
)

const (
	defaultColor   = "#53FC18"
	channelBufSize = 100
	readTimeout    = 180 * time.Second
	connectTimeout = 10 * time.Second
	readLimit      = 1 << 20 // 1MB
)

type apiResponse struct {
	Chatroom struct {
		ID int `json:"id"`
	} `json:"chatroom"`
}

// wsEvent represents a Pusher WebSocket event
type wsEvent struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}

// chat message payload
type chatMessageData struct {
	Content string `json:"content"`
	Sender  struct {
		Username string `json:"username"`
	} `json:"sender"`
}

// Kick chat client
type Client struct {
	config     Config
	chatroomID int
	log        logger.Logger

	mu     sync.RWMutex
	conn   *websocket.Conn
	closed bool
}

func NewClient(channel string, log logger.Logger) (*Client, error) {
	return NewClientWithConfig(DefaultConfig(channel), log)
}

func NewClientWithConfig(cfg Config, log logger.Logger) (*Client, error) {
	if cfg.Channel == "" {
		return nil, errors.New("channel is required")
	}

	l := log.With(zap.String("adapter", "kick"), zap.String("channel", cfg.Channel))

	chatroomID, err := fetchChatroomID(cfg.Channel, cfg.HTTPClient, cfg.APIURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get chatroom ID for channel %s: %w", cfg.Channel, err)
	}

	l.Info("client created", zap.Int("chatroom_id", chatroomID))

	return &Client{
		config:     cfg,
		chatroomID: chatroomID,
		log:        l,
	}, nil
}

// returns a channel of chat messages.
func (c *Client) Stream(ctx context.Context) <-chan models.ChatMessage {
	out := make(chan models.ChatMessage, channelBufSize)
	go c.streamLoop(ctx, out)
	return out
}

// gracefully shuts down the client.
func (c *Client) Close() error {
	c.mu.Lock()
	c.closed = true
	conn := c.conn
	c.conn = nil
	c.mu.Unlock()

	if conn != nil {
		return conn.Close(websocket.StatusNormalClosure, "closing")
	}
	return nil
}

func (c *Client) Platform() string {
	return "Kick"
}

// manage connection lifecycle with automatic reconnection
func (c *Client) streamLoop(ctx context.Context, out chan<- models.ChatMessage) {
	defer close(out)

	delay := c.config.RetryConfig.InitialDelay

	for {
		if c.shouldStop(ctx) {
			return
		}

		if err := c.runSession(ctx, out); err != nil {
			c.log.Warn("session ended", zap.Error(err), zap.Duration("retry_in", delay))
			delay = c.backoff(ctx, delay)
			if delay == 0 {
				return
			}
			continue
		}

		// reset backoff on clean disconnect
		delay = c.config.RetryConfig.InitialDelay
	}
}

// checks if the client should stop streaming
func (c *Client) shouldStop(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		c.log.Debug("context cancelled, stopping stream")
		return true
	default:
	}

	c.mu.RLock()
	closed := c.closed
	c.mu.RUnlock()
	return closed
}

// run a single connection session until disconnection
func (c *Client) runSession(ctx context.Context, out chan<- models.ChatMessage) error {
	if err := c.connect(ctx); err != nil {
		return err
	}
	defer c.disconnect()

	if err := c.subscribe(ctx); err != nil {
		return err
	}

	return c.readLoop(ctx, out)
}

// connect establishes the websocket connection
func (c *Client) connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	connectCtx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()

	conn, _, err := websocket.Dial(connectCtx, c.config.WebSocketURL, nil)
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}

	conn.SetReadLimit(readLimit)
	c.conn = conn
	c.log.Info("connected")
	return nil
}

// subscribe to chatroom channel
func (c *Client) subscribe(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return errors.New("not connected")
	}

	payload := map[string]any{
		"event": "pusher:subscribe",
		"data": map[string]string{
			"auth":    "",
			"channel": fmt.Sprintf("chatrooms.%d.v2", c.chatroomID),
		},
	}

	if err := wsjson.Write(ctx, c.conn, payload); err != nil {
		return fmt.Errorf("subscribe: %w", err)
	}

	c.log.Info("joined channel", zap.Int("chatroom_id", c.chatroomID))
	return nil
}

// read and process websocket events
func (c *Client) readLoop(ctx context.Context, out chan<- models.ChatMessage) error {
	for {
		if c.shouldStop(ctx) {
			return nil
		}

		event, err := c.readEvent(ctx)
		if err != nil {
			if !c.isClosed() {
				return err
			}
			return nil
		}

		c.handleEvent(ctx, event, out)
	}
}

// read a single ws event
func (c *Client) readEvent(ctx context.Context) (wsEvent, error) {
	readCtx, cancel := context.WithTimeout(ctx, readTimeout)
	defer cancel()

	c.mu.RLock()
	conn := c.conn
	closed := c.closed
	c.mu.RUnlock()

	if conn == nil || closed {
		return wsEvent{}, errors.New("connection closed")
	}

	var event wsEvent
	if err := wsjson.Read(readCtx, conn, &event); err != nil {
		return wsEvent{}, err
	}

	return event, nil
}

func (c *Client) isClosed() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.closed
}

// route ws events to appropriate handlers
func (c *Client) handleEvent(ctx context.Context, event wsEvent, out chan<- models.ChatMessage) {
	switch event.Event {
	case "pusher:ping":
		c.sendPong(ctx)
	case "App\\Events\\ChatMessageEvent":
		c.handleChatMessage(ctx, event.Data, out)
	}
}

func (c *Client) sendPong(ctx context.Context) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		_ = wsjson.Write(ctx, c.conn, map[string]string{"event": "pusher:pong", "data": "{}"})
	}
}

// processes a chat message event
func (c *Client) handleChatMessage(ctx context.Context, data string, out chan<- models.ChatMessage) {
	msg, err := c.parseMessage(data)
	if err != nil {
		c.log.Debug("failed to parse message", zap.Error(err))
		return
	}

	select {
	case out <- msg:
		c.log.Debug("message received", zap.String("username", msg.Username))
	case <-ctx.Done():
	}
}

// converts raw json data to model.ChatMessage
func (c *Client) parseMessage(data string) (models.ChatMessage, error) {
	var msg chatMessageData
	if err := json.Unmarshal([]byte(data), &msg); err != nil {
		return models.ChatMessage{}, err
	}

	return models.ChatMessage{
		Platform:  "Kick",
		Username:  msg.Sender.Username,
		Content:   msg.Content,
		Color:     defaultColor,
		Timestamp: time.Now(),
	}, nil
}

func (c *Client) disconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		c.conn.Close(websocket.StatusNormalClosure, "")
		c.conn = nil
		c.log.Debug("disconnected")
	}
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
	if next > c.config.RetryConfig.MaxDelay {
		next = c.config.RetryConfig.MaxDelay
	}
	return next
}

func fetchChatroomID(slug string, httpClient common.HTTPClient, apiURL string) (int, error) {
	req, err := http.NewRequest(http.MethodGet, apiURL+slug, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; KickChat/1.0)")

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var data apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}

	if data.Chatroom.ID == 0 {
		return 0, fmt.Errorf("chatroom not found: %s", slug)
	}

	return data.Chatroom.ID, nil
}

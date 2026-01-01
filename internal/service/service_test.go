package service

import (
	"context"
	"testing"
	"time"

	"github.com/llan0/multichat/internal/logger"
	"github.com/llan0/multichat/internal/models"
)

type mockProducer struct {
	messages []models.ChatMessage
}

func (m *mockProducer) Stream(ctx context.Context) <-chan models.ChatMessage {
	ch := make(chan models.ChatMessage)
	go func() {
		defer close(ch)
		for _, msg := range m.messages {
			select {
			case <-ctx.Done():
				return
			case ch <- msg:
			}
		}
	}()
	return ch
}

func TestMerge_SingleProducer(t *testing.T) {
	log := logger.NewNop()
	producer := &mockProducer{
		messages: []models.ChatMessage{
			{Platform: "test", Username: "user1", Content: "hello"},
			{Platform: "test", Username: "user2", Content: "world"},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	merged := Merge(ctx, log, producer)

	var received []models.ChatMessage
	for msg := range merged {
		received = append(received, msg)
	}

	if len(received) != 2 {
		t.Errorf("received %d messages, want 2", len(received))
	}
}

func TestMerge_MultipleProducers(t *testing.T) {
	log := logger.NewNop()
	p1 := &mockProducer{messages: []models.ChatMessage{{Platform: "p1", Content: "a"}}}
	p2 := &mockProducer{messages: []models.ChatMessage{{Platform: "p2", Content: "b"}}}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	merged := Merge(ctx, log, p1, p2)

	var received []models.ChatMessage
	for msg := range merged {
		received = append(received, msg)
	}

	if len(received) != 2 {
		t.Errorf("received %d messages, want 2", len(received))
	}
}

func TestMerge_NoProducers(t *testing.T) {
	log := logger.NewNop()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	merged := Merge(ctx, log)

	var received []models.ChatMessage
	for msg := range merged {
		received = append(received, msg)
	}

	if len(received) != 0 {
		t.Errorf("received %d messages, want 0", len(received))
	}
}

package models

import "time"

type ChatMessage struct {
	Platform  string
	Username  string
	Content   string
	Color     string // hex code
	Timestamp time.Time
}

package models

import "time"

type SegmentType int

const (
	SegmentText SegmentType = iota
	SegmentEmote
)

type MessageSegment struct {
	Type      SegmentType
	Text      string // for text segments
	EmoteID   string // for emote segments
	EmoteURL  string // CDN URL for the emote
	EmoteName string // display name of the emote
}

type ChatMessage struct {
	Platform  string
	Username  string
	Content   string           // original raw content (kept for fallback)
	Segments  []MessageSegment // parsed segments (text + emotes)
	Color     string           // hex code
	Timestamp time.Time
}

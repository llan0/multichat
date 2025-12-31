package emotes

import "fmt"

// construct CDN URL for a twitch emote
func BuildTwitchEmoteURL(id string) string {
	return fmt.Sprintf("https://static-cdn.jtvnw.net/emoticons/v2/%s/default/dark/1.0", id)
}

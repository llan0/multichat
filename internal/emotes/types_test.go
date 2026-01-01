package emotes

import "testing"

func TestBuildTwitchEmoteURL(t *testing.T) {
	tests := []struct {
		id   string
		want string
	}{
		{"25", "https://static-cdn.jtvnw.net/emoticons/v2/25/default/dark/1.0"},
		{"emotesv2_abc123", "https://static-cdn.jtvnw.net/emoticons/v2/emotesv2_abc123/default/dark/1.0"},
	}

	for _, tt := range tests {
		got := BuildTwitchEmoteURL(tt.id)
		if got != tt.want {
			t.Errorf("BuildTwitchEmoteURL(%q) = %q, want %q", tt.id, got, tt.want)
		}
	}
}

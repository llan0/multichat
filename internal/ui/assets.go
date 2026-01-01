package ui

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed assets/icon.png
var appIconData []byte

//go:embed assets/twitch.png
var twitchIconData []byte

//go:embed assets/kick.png
var kickIconData []byte

var (
	appIcon    = fyne.NewStaticResource("icon.png", appIconData)
	twitchIcon = fyne.NewStaticResource("twitch.png", twitchIconData)
	kickIcon   = fyne.NewStaticResource("kick.png", kickIconData)
)

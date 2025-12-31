package ui

import (
	"context"
	"fmt"
	"image/color"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/llan0/multichat/internal/adapters/kick"
	"github.com/llan0/multichat/internal/adapters/twitch"
	"github.com/llan0/multichat/internal/emotes"
	"github.com/llan0/multichat/internal/logger"
	"github.com/llan0/multichat/internal/models"
	"github.com/llan0/multichat/internal/service"
	"go.uber.org/zap"
)

const (
	appVersion   = "0.0.1"
	windowWidth  = 380
	windowHeight = 700
)

type chatMessage struct {
	Platform string
	Username string
	Content  string
	Segments []models.MessageSegment
	Color    color.Color
}

type App struct {
	log    logger.Logger
	window fyne.Window

	ctx       context.Context
	cancelCtx context.CancelFunc

	channelEntry *widget.Entry
	showTwitch   bool
	showKick     bool

	emoteService *emotes.Service

	mu       sync.Mutex
	messages []chatMessage
	chatList *widget.List
}

func Run(log logger.Logger, defaultChannel string) {
	a := &App{
		log:          log,
		messages:     make([]chatMessage, 0, 1000),
		showTwitch:   true,
		showKick:     true,
		emoteService: emotes.NewService(log),
	}

	a.createWindow()
	a.setupLayout(defaultChannel)
	a.connectToChannel(defaultChannel)

	a.window.ShowAndRun()

	if a.cancelCtx != nil {
		a.cancelCtx()
	}
}

func (a *App) createWindow() {
	fyneApp := app.New()
	fyneApp.Settings().SetTheme(theme.DefaultTheme())
	a.window = fyneApp.NewWindow(fmt.Sprintf("multichat %s", appVersion))
	a.window.Resize(fyne.NewSize(windowWidth, windowHeight))
	a.window.CenterOnScreen()
}

func (a *App) setupLayout(defaultChannel string) {
	topBar := a.createTopBar(defaultChannel)
	a.chatList = a.createChatList()
	chatContainer := container.NewPadded(a.chatList)

	content := container.NewBorder(topBar, nil, nil, nil, chatContainer)
	a.window.SetContent(content)
}

func (a *App) createTopBar(defaultChannel string) fyne.CanvasObject {
	a.channelEntry = widget.NewEntry()
	a.channelEntry.SetPlaceHolder("Enter channel name...")
	a.channelEntry.SetText(defaultChannel)
	a.channelEntry.OnSubmitted = func(channel string) {
		if channel != "" {
			a.connectToChannel(channel)
		}
	}

	twitchCheck := widget.NewCheck("Twitch", func(checked bool) {
		a.showTwitch = checked
		a.chatList.Refresh()
	})
	twitchCheck.Checked = true

	kickCheck := widget.NewCheck("Kick", func(checked bool) {
		a.showKick = checked
		a.chatList.Refresh()
	})
	kickCheck.Checked = true

	filterRow := container.NewHBox(twitchCheck, kickCheck)
	topRow := container.NewBorder(nil, nil, nil, filterRow, a.channelEntry)

	return container.NewVBox(
		container.NewPadded(topRow),
		widget.NewSeparator(),
	)
}

func (a *App) connectToChannel(channel string) {
	a.log.Info("connecting to channel", zap.String("channel", channel))

	if a.cancelCtx != nil {
		a.cancelCtx()
	}

	a.mu.Lock()
	a.messages = a.messages[:0]
	a.mu.Unlock()

	fyne.Do(func() {
		a.chatList.Refresh()
	})

	a.ctx, a.cancelCtx = context.WithCancel(context.Background())

	go func() {
		twitchClient, err := twitch.NewClient(channel, a.log)
		if err != nil {
			a.log.Error("failed to create Twitch client", zap.Error(err))
			return
		}

		kickClient, err := kick.NewClient(channel, a.log)
		if err != nil {
			a.log.Error("failed to create Kick client", zap.Error(err))
			twitchClient.Close()
			return
		}

		mergedStream := service.Merge(a.ctx, a.log, twitchClient, kickClient)
		a.log.Info("connected to channel", zap.String("channel", channel))

		for msg := range mergedStream {
			a.addMessage(msg)
		}

		twitchClient.Close()
		kickClient.Close()
	}()
}

func (a *App) filteredMessages() []chatMessage {
	result := make([]chatMessage, 0, len(a.messages))
	for _, msg := range a.messages {
		if msg.Platform == "Twitch" && a.showTwitch {
			result = append(result, msg)
		} else if msg.Platform == "Kick" && a.showKick {
			result = append(result, msg)
		}
	}
	return result
}

func (a *App) createChatList() *widget.List {
	return widget.NewList(
		func() int {
			a.mu.Lock()
			defer a.mu.Unlock()
			return len(a.filteredMessages())
		},
		func() fyne.CanvasObject {
			return a.createMessageRow()
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			a.updateMessageRow(id, obj)
		},
	)
}

func (a *App) createMessageRow() fyne.CanvasObject {
	platformIcon := canvas.NewImageFromFile("internal/ui/assets/twitch.png")
	platformIcon.FillMode = canvas.ImageFillContain
	platformIcon.SetMinSize(fyne.NewSize(16, 16))

	username := canvas.NewText("username:", color.White)
	username.TextStyle = fyne.TextStyle{Bold: true}
	username.TextSize = 13

	// Use hbox for content to support mixed text/images
	content := container.NewHBox()

	left := container.NewHBox(platformIcon, username)
	return container.NewBorder(nil, nil, left, nil, content)
}

func (a *App) updateMessageRow(id widget.ListItemID, obj fyne.CanvasObject) {
	a.mu.Lock()
	filtered := a.filteredMessages()
	if id >= len(filtered) {
		a.mu.Unlock()
		return
	}
	msg := filtered[id]
	a.mu.Unlock()

	row := obj.(*fyne.Container)
	content := row.Objects[0].(*fyne.Container)
	left := row.Objects[1].(*fyne.Container)
	platformIcon := left.Objects[0].(*canvas.Image)
	username := left.Objects[1].(*canvas.Text)

	if msg.Platform == "Twitch" {
		platformIcon.File = "internal/ui/assets/twitch.png"
	} else {
		platformIcon.File = "internal/ui/assets/kick.png"
	}
	platformIcon.Refresh()

	username.Text = msg.Username + ":"
	username.Color = msg.Color
	username.Refresh()

	// Build content with text and emote images
	a.buildMessageContent(content, msg)
}

func (a *App) buildMessageContent(container *fyne.Container, msg chatMessage) {
	container.Objects = nil

	// If no segments, fall back to plain text
	if len(msg.Segments) == 0 {
		label := widget.NewLabel(msg.Content)
		label.Truncation = fyne.TextTruncateEllipsis
		container.Objects = append(container.Objects, label)
		container.Refresh()
		return
	}

	for _, seg := range msg.Segments {
		switch seg.Type {
		case models.SegmentText:
			if seg.Text != "" {
				text := canvas.NewText(seg.Text, color.White)
				text.TextSize = 13
				container.Objects = append(container.Objects, text)
			}
		case models.SegmentEmote:
			img := a.createEmoteImage(seg.EmoteURL)
			container.Objects = append(container.Objects, img)
		}
	}

	container.Refresh()
}

func (a *App) createEmoteImage(url string) fyne.CanvasObject {
	resource := a.emoteService.GetImage(url, func() {
		// Callback when image loads - refresh the list
		fyne.Do(func() {
			a.chatList.Refresh()
		})
	})

	img := canvas.NewImageFromResource(resource)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(24, 24))

	return img
}

func (a *App) addMessage(msg models.ChatMessage) {
	chatMsg := chatMessage{
		Platform: msg.Platform,
		Username: msg.Username,
		Content:  msg.Content,
		Segments: msg.Segments,
		Color:    parseHexColor(msg.Color),
	}

	a.mu.Lock()
	a.messages = append(a.messages, chatMsg)
	a.mu.Unlock()

	fyne.Do(func() {
		a.chatList.Refresh()
		a.chatList.ScrollToBottom()
	})
}

func parseHexColor(hex string) color.Color {
	if len(hex) == 0 {
		return color.White
	}

	if hex[0] == '#' {
		hex = hex[1:]
	}

	if len(hex) != 6 {
		return color.White
	}

	var r, g, b uint8
	_, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	if err != nil {
		return color.White
	}

	return color.RGBA{R: r, G: g, B: b, A: 255}
}

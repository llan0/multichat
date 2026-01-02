package ui

import (
	"context"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/llan0/multichat/internal/adapters/kick"
	"github.com/llan0/multichat/internal/adapters/twitch"
	"github.com/llan0/multichat/internal/models"
	"github.com/llan0/multichat/internal/service"
	"go.uber.org/zap"
)

type chatMessage struct {
	Platform string
	Username string
	Content  string
	Segments []models.MessageSegment
	Color    color.Color
}

func (a *App) createChatTab(defaultChannel string) *container.TabItem {
	topBar := a.createTopBar(defaultChannel)
	a.chatList = a.createChatList()
	chatContainer := container.NewPadded(a.chatList)

	content := container.NewBorder(topBar, nil, nil, nil, chatContainer)
	return container.NewTabItem(defaultChannel, content)
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

	settingsBtn := widget.NewButtonWithIcon("", theme.SettingsIcon(), func() {
		a.openSettings()
	})

	filterRow := container.NewHBox(twitchCheck, kickCheck)
	topRow := container.NewBorder(nil, nil, settingsBtn, filterRow, a.channelEntry)

	return container.NewVBox(
		container.NewPadded(topRow),
		widget.NewSeparator(),
	)
}

func (a *App) connectToChannel(channel string) {
	a.log.Info("connecting to channel", zap.String("channel", channel))

	// cancel previous connection and wait for cleanup
	if a.cancelCtx != nil {
		a.cancelCtx()
		a.connWg.Wait() // wait for previous goroutine to finish
	}

	a.mu.Lock()
	a.messages = a.messages[:0]
	a.mu.Unlock()

	fyne.Do(func() {
		a.chatList.Refresh()
	})

	a.ctx, a.cancelCtx = context.WithCancel(context.Background())

	a.connWg.Go(func() {
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
	})
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
	platformIcon := canvas.NewImageFromResource(twitchIcon)
	platformIcon.FillMode = canvas.ImageFillContain
	platformIcon.SetMinSize(fyne.NewSize(16, 16))

	username := canvas.NewText("username:", theme.ForegroundColor())
	username.TextStyle = fyne.TextStyle{Bold: true}
	username.TextSize = 13

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
		platformIcon.Resource = twitchIcon
	} else {
		platformIcon.Resource = kickIcon
	}
	platformIcon.Refresh()

	username.Text = msg.Username + ":"
	username.Color = msg.Color
	username.Refresh()

	a.buildMessageContent(content, msg)
}

func (a *App) buildMessageContent(container *fyne.Container, msg chatMessage) {
	container.Objects = nil

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
				text := canvas.NewText(seg.Text, theme.ForegroundColor())
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
	// evict old messages to prevent unbounded memory growth
	if len(a.messages) > maxMessages {
		// remove oldest messages, keep last maxMessages
		copy(a.messages, a.messages[len(a.messages)-maxMessages:])
		a.messages = a.messages[:maxMessages]
	}
	a.mu.Unlock()

	fyne.Do(func() {
		a.chatList.Refresh()
		a.chatList.ScrollToBottom()
	})
}

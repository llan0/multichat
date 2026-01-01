package ui

import (
	"context"
	"fmt"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/llan0/multichat/internal/emotes"
	"github.com/llan0/multichat/internal/logger"
)

const (
	appVersion   = "0.0.1"
	windowWidth  = 380
	windowHeight = 700
)

type App struct {
	log     logger.Logger
	fyneApp fyne.App
	window  fyne.Window

	ctx       context.Context
	cancelCtx context.CancelFunc

	// chat state
	channelEntry *widget.Entry
	showTwitch   bool
	showKick     bool
	emoteService *emotes.Service
	mu           sync.Mutex
	messages     []chatMessage
	chatList     *widget.List

	// tabs
	tabs        *container.DocTabs
	settingsTab *container.TabItem
	chatTab     *container.TabItem
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
	a.fyneApp = app.New()
	// a.fyneApp.SetIcon(appIcon)
	a.fyneApp.Settings().SetTheme(theme.DefaultTheme())
	a.window = a.fyneApp.NewWindow(fmt.Sprintf("multichat %s", appVersion))
	a.window.Resize(fyne.NewSize(windowWidth, windowHeight))
	a.window.CenterOnScreen()
}

func (a *App) setupLayout(defaultChannel string) {
	a.chatTab = a.createChatTab(defaultChannel)
	a.settingsTab = a.createSettingsTab()

	a.tabs = container.NewDocTabs(a.chatTab)
	a.tabs.SetTabLocation(container.TabLocationTop)

	// TODO: prevent closing the main chat tab for now
	a.tabs.OnClosed = func(tab *container.TabItem) {
		if tab == a.chatTab {
			a.tabs.Append(a.chatTab)
			a.tabs.Select(a.chatTab)
		}
	}

	a.window.SetContent(a.tabs)
}

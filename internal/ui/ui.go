package ui

import (
	"context"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"

	"github.com/llan0/multichat/internal/logger"
	"github.com/llan0/multichat/internal/models"
	"github.com/llan0/multichat/internal/service"
	"go.uber.org/zap"
)

const (
	appVersion  = "0.0.1"
	channelName = "xQc"
)

func ShowUI(ctx context.Context, mergedStream <-chan models.ChatMessage) {
	logger.Log.Info("initializing UI")
	a := app.New()
	w := a.NewWindow(fmt.Sprintf("multichat %s", appVersion))
	w.Resize(fyne.NewSize(400, 900))
	w.CenterOnScreen()

	// status line
	statusText := widget.NewLabel(fmt.Sprintf("%s ", channelName))
	statusContainer := container.NewHBox(statusText)

	// chat display area with binding for threadsafe updates
	messages := binding.NewStringList()
	chatList := widget.NewListWithData(
		messages,
		func() fyne.CanvasObject {
			label := widget.NewLabel("template")
			label.Wrapping = fyne.TextWrapWord
			return label
		},
		func(i binding.DataItem, o fyne.CanvasObject) {
			val, _ := i.(binding.String).Get()
			o.(*widget.Label).SetText(val)
		},
	)

	// autoscroll on new messages
	messages.AddListener(binding.NewDataListener(func() {
		if messages.Length() > 0 {
			chatList.ScrollTo(messages.Length() - 1)
		}
	}))

	// input field at bottom
	inputEntry := widget.NewEntry()
	inputEntry.SetPlaceHolder("Send a message")
	inputEntry.OnSubmitted = func(text string) {
		inputEntry.SetText("")
	}

	topBar := container.NewVBox(
		statusContainer,
		widget.NewSeparator(),
	)

	content := container.NewBorder(
		topBar,
		inputEntry, // fixed at bottom
		nil, nil,
		chatList, // scrollable chat area
	)

	w.SetContent(content)

	// consume from merged stream and update ui
	go func() {
		logger.Log.Info("starting message consumer")
		messageCount := 0
		for msg := range mergedStream {
			formatted := fmt.Sprintf("%s %s: %s", platformIcon(msg.Platform), msg.Username, msg.Content)

			fyne.Do(func() {
				messages.Append(formatted)
			})

			messageCount++
			logger.Log.Debug("message displayed",
				zap.String("platform", msg.Platform),
				zap.String("username", msg.Username),
				zap.Int("total_messages", messageCount),
			)
		}
		logger.Log.Info("message stream closed", zap.Int("total_messages", messageCount))
	}()

	logger.Log.Info("showing UI window")
	w.ShowAndRun()
	logger.Log.Info("UI window closed")
}

// start the ui with mock producers for now
func Start(ctx context.Context, producers ...service.Producer) {
	logger.Log.Info("starting UI with producers", zap.Int("producer_count", len(producers)))
	mergedStream := service.Merge(ctx, producers...)
	ShowUI(ctx, mergedStream)
}

// TODO: replace with actual icons
func platformIcon(platform string) string {
	return "[" + platform + "]"
}

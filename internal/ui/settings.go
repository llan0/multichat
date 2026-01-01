package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (a *App) createSettingsTab() *container.TabItem {
	return container.NewTabItem("Settings", a.createSettingsPanel())
}

func (a *App) openSettings() {
	// check if settings tab is already open
	for _, tab := range a.tabs.Items {
		if tab == a.settingsTab {
			a.tabs.Select(a.settingsTab)
			return
		}
	}
	// add and select settings tab
	a.tabs.Append(a.settingsTab)
	a.tabs.Select(a.settingsTab)
}

func (a *App) createSettingsPanel() fyne.CanvasObject {
	header := widget.NewLabelWithStyle("General", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	themeRadio := widget.NewRadioGroup([]string{"Dark", "Light"}, func(selected string) {
		switch selected {
		case "Dark":
			a.fyneApp.Settings().SetTheme(theme.DarkTheme())
		case "Light":
			a.fyneApp.Settings().SetTheme(theme.LightTheme())
		}
	})
	themeRadio.Horizontal = true
	themeRadio.SetSelected("Dark")

	themeRow := container.NewHBox(widget.NewLabel("Theme:"), themeRadio)

	content := container.NewVBox(
		header,
		widget.NewSeparator(),
		themeRow,
	)

	return container.NewPadded(content)
}

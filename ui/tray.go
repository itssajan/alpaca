package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/itssajan/alpaca/mover"
)

// SetupTray configures the system tray menu and icon.
// Must be called after the Fyne app has started.
func SetupTray(app fyne.App, m *mover.Mover) {
	if desk, ok := app.(desktop.App); ok {
		statusItem := fyne.NewMenuItem("● Alpaca — Active", nil)
		statusItem.Disabled = true

		toggleItem := fyne.NewMenuItem("Pause", nil)
		toggleItem.Action = func() {
			if m.Enabled() {
				m.SetEnabled(false)
				statusItem.Label = "○ Alpaca — Paused"
				toggleItem.Label = "Enable"
			} else {
				m.SetEnabled(true)
				statusItem.Label = "● Alpaca — Active"
				toggleItem.Label = "Pause"
			}
		}

		moveNowItem := fyne.NewMenuItem("Move Now", func() {
			m.MoveNow()
		})

		settingsItem := fyne.NewMenuItem("Settings", func() {
			ShowSettings(app, m)
		})

		quitItem := fyne.NewMenuItem("Quit", func() {
			m.Stop()
			app.Quit()
		})

		menu := fyne.NewMenu("Alpaca",
			statusItem,
			fyne.NewMenuItemSeparator(),
			settingsItem,
			toggleItem,
			moveNowItem,
			fyne.NewMenuItemSeparator(),
			quitItem,
		)

		desk.SetSystemTrayMenu(menu)
		desk.SetSystemTrayIcon(trayIcon())
	}
}

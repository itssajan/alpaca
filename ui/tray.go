package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/itssajan/alpaca/mover"
)

func SetupTray(app fyne.App, m *mover.Mover) {
	desk, ok := app.(desktop.App)
	if !ok {
		return
	}

	statusItem := fyne.NewMenuItem("● Alpaca — starting…", nil)
	statusItem.Disabled = true

	toggleItem := fyne.NewMenuItem("Pause", nil)

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

	// flushMenu re-renders the tray — required after any label mutation in Fyne.
	flushMenu := func() {
		desk.SetSystemTrayMenu(menu)
	}

	toggleItem.Action = func() {
		if m.Enabled() {
			m.SetEnabled(false)
			toggleItem.Label = "Enable"
		} else {
			m.SetEnabled(true)
			toggleItem.Label = "Pause"
		}
		flushMenu()
	}

	desk.SetSystemTrayMenu(menu)
	desk.SetSystemTrayIcon(trayIcon())

	// Status polling goroutine — updates the tray label every second.
	go func() {
		var lastLabel string
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for range ticker.C {
			label := statusLabel(m)
			if label != lastLabel {
				statusItem.Label = label
				flushMenu()
				lastLabel = label
			}
		}
	}()
}

func statusLabel(m *mover.Mover) string {
	if !m.Enabled() {
		return "○ Alpaca — Paused"
	}
	if m.IsUserActive() {
		return "⏸ Alpaca — Waiting (you're active)"
	}
	d := m.NextMoveIn()
	if d <= 0 {
		return "● Alpaca — Moving…"
	}
	secs := int(d.Seconds())
	if secs >= 60 {
		return fmt.Sprintf("● Alpaca — Next move in %dm%ds", secs/60, secs%60)
	}
	return fmt.Sprintf("● Alpaca — Next move in %ds", secs)
}

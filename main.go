package main

import (
	"fyne.io/fyne/v2/app"
	"github.com/itssajan/alpaca/config"
	"github.com/itssajan/alpaca/mover"
	alpacaUI "github.com/itssajan/alpaca/ui"
)

func main() {
	if _, err := config.Load(); err != nil {
		_ = err
	}

	m := mover.New()
	m.Start()

	a := app.NewWithID("com.itssajan.alpaca")

	alpacaUI.SetupTray(a, m)

	// Run Fyne event loop — blocks until Quit.
	a.Run()
}

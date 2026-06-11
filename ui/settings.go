package ui

import (
	"fmt"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/itssajan/alpaca/config"
	"github.com/itssajan/alpaca/mover"
)

func ShowSettings(app fyne.App, m *mover.Mover) {
	w := app.NewWindow("Alpaca Settings")
	w.Resize(fyne.NewSize(420, 380))
	w.SetFixedSize(true)

	cfg := config.Get()

	// --- Interval ---
	minLabel := widget.NewLabel(fmt.Sprintf("%ds", cfg.Interval.MinSeconds))
	maxLabel := widget.NewLabel(fmt.Sprintf("%ds", cfg.Interval.MaxSeconds))

	minSlider := widget.NewSlider(5, 300)
	minSlider.Step = 5
	minSlider.Value = float64(cfg.Interval.MinSeconds)

	maxSlider := widget.NewSlider(5, 300)
	maxSlider.Step = 5
	maxSlider.Value = float64(cfg.Interval.MaxSeconds)

	minSlider.OnChanged = func(v float64) {
		minLabel.SetText(fmt.Sprintf("%ds", int(v)))
		if v >= maxSlider.Value {
			maxSlider.Value = v + 5
			maxSlider.Refresh()
			maxLabel.SetText(fmt.Sprintf("%ds", int(maxSlider.Value)))
		}
	}
	maxSlider.OnChanged = func(v float64) {
		maxLabel.SetText(fmt.Sprintf("%ds", int(v)))
		if v <= minSlider.Value {
			minSlider.Value = v - 5
			minSlider.Refresh()
			minLabel.SetText(fmt.Sprintf("%ds", int(minSlider.Value)))
		}
	}

	// --- Quiet period ---
	quietLabel := widget.NewLabel(fmt.Sprintf("%ds", cfg.Idle.QuietPeriodSeconds))
	quietSlider := widget.NewSlider(1, 30)
	quietSlider.Step = 1
	quietSlider.Value = float64(cfg.Idle.QuietPeriodSeconds)
	quietSlider.OnChanged = func(v float64) {
		quietLabel.SetText(fmt.Sprintf("%ds", int(v)))
	}

	// --- Movement weights ---
	jLabel := widget.NewLabel(fmt.Sprintf("%d%%", cfg.Movement.JitterWeight))
	dLabel := widget.NewLabel(fmt.Sprintf("%d%%", cfg.Movement.DriftWeight))
	wLabel := widget.NewLabel(fmt.Sprintf("%d%%", cfg.Movement.WanderWeight))

	jSlider := widget.NewSlider(0, 100)
	jSlider.Step = 5
	jSlider.Value = float64(cfg.Movement.JitterWeight)

	dSlider := widget.NewSlider(0, 100)
	dSlider.Step = 5
	dSlider.Value = float64(cfg.Movement.DriftWeight)

	wSlider := widget.NewSlider(0, 100)
	wSlider.Step = 5
	wSlider.Value = float64(cfg.Movement.WanderWeight)

	// Normalize weights so they always sum to 100 when one changes.
	updateWeights := func(changed *widget.Slider, changedLabel *widget.Label) func(float64) {
		return func(v float64) {
			changedLabel.SetText(fmt.Sprintf("%d%%", int(v)))
			// distribute remainder equally between the other two
			rest := 100 - v
			others := []*widget.Slider{jSlider, dSlider, wSlider}
			otherLabels := []*widget.Label{jLabel, dLabel, wLabel}
			sum := 0.0
			for _, s := range others {
				if s != changed {
					sum += s.Value
				}
			}
			for i, s := range others {
				if s == changed {
					continue
				}
				var share float64
				if sum == 0 {
					share = rest / 2
				} else {
					share = math.Round((s.Value/sum)*rest/5) * 5
				}
				s.Value = share
				s.Refresh()
				otherLabels[i].SetText(fmt.Sprintf("%d%%", int(share)))
			}
		}
	}

	jSlider.OnChanged = updateWeights(jSlider, jLabel)
	dSlider.OnChanged = updateWeights(dSlider, dLabel)
	wSlider.OnChanged = updateWeights(wSlider, wLabel)

	// --- Buttons ---
	saveBtn := widget.NewButton("Save", func() {
		newCfg := config.Config{
			Interval: config.Interval{
				MinSeconds: int(minSlider.Value),
				MaxSeconds: int(maxSlider.Value),
			},
			Idle: config.Idle{
				QuietPeriodSeconds: int(quietSlider.Value),
			},
			Movement: config.Movement{
				JitterWeight: int(jSlider.Value),
				DriftWeight:  int(dSlider.Value),
				WanderWeight: int(wSlider.Value),
			},
		}
		_ = config.Save(newCfg)
		m.ReloadConfig(newCfg)
		w.Close()
	})

	cancelBtn := widget.NewButton("Cancel", func() { w.Close() })

	form := container.NewVBox(
		widget.NewLabel("Interval"),
		container.NewBorder(nil, nil, widget.NewLabel("Min"), minLabel, minSlider),
		container.NewBorder(nil, nil, widget.NewLabel("Max"), maxLabel, maxSlider),

		widget.NewSeparator(),
		widget.NewLabel("Idle detection"),
		container.NewBorder(nil, nil, widget.NewLabel("Quiet period"), quietLabel, quietSlider),

		widget.NewSeparator(),
		widget.NewLabel("Movement style"),
		container.NewBorder(nil, nil, widget.NewLabel("Jitter "), jLabel, jSlider),
		container.NewBorder(nil, nil, widget.NewLabel("Drift  "), dLabel, dSlider),
		container.NewBorder(nil, nil, widget.NewLabel("Wander "), wLabel, wSlider),

		widget.NewSeparator(),
		container.NewHBox(layout.NewSpacer(), cancelBtn, saveBtn),
	)

	w.SetContent(container.NewPadded(form))
	w.Show()
}

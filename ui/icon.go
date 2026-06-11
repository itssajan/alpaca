package ui

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	"image/png"

	"fyne.io/fyne/v2"
)

//go:embed icon.png
var iconBytes []byte

func trayIcon() fyne.Resource {
	return fyne.NewStaticResource("icon.png", iconBytes)
}

// GenerateIcon creates a simple 32x32 circle PNG for the tray icon.
// Call this once during build/init if icon.png doesn't exist yet.
func GenerateIcon() ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	cx, cy, r := 16.0, 16.0, 13.0
	teal := color.RGBA{R: 0x2d, G: 0xb2, B: 0x82, A: 0xff}
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			if dx*dx+dy*dy <= r*r {
				img.Set(x, y, teal)
			}
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

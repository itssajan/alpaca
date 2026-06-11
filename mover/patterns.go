package mover

import (
	"math"
	"math/rand"
	"time"

	"github.com/go-vgo/robotgo"
)

// jitter performs small random nudges around the current position.
func jitter() {
	x, y := robotgo.Location()
	fx, fy := float64(x), float64(y)

	nudges := 5 + rand.Intn(4) // 5–8 nudges
	for i := 0; i < nudges; i++ {
		dx := float64(rand.Intn(31)-15) // ±15 px
		dy := float64(rand.Intn(31)-15)
		nx := math.Round(fx + dx)
		ny := math.Round(fy + dy)
		robotgo.Move(int(nx), int(ny))
		time.Sleep(time.Duration(20+rand.Intn(41)) * time.Millisecond)
	}
	// return to origin
	robotgo.Move(x, y)
}

// drift moves the cursor to a nearby point (50–150 px away) along a bezier.
func drift() {
	x, y := robotgo.Location()
	sw, sh := robotgo.GetScreenSize()

	angle := rand.Float64() * 2 * math.Pi
	d := 50 + rand.Float64()*100
	tx := clamp(float64(x)+math.Cos(angle)*d, 0, float64(sw-1))
	ty := clamp(float64(y)+math.Sin(angle)*d, 0, float64(sh-1))

	moveBezier(float64(x), float64(y), tx, ty, 6.0) // moderate speed
}

// wander moves the cursor to a random location anywhere on screen.
func wander() {
	x, y := robotgo.Location()
	sw, sh := robotgo.GetScreenSize()

	tx := float64(rand.Intn(sw))
	ty := float64(rand.Intn(sh))

	moveBezier(float64(x), float64(y), tx, ty, 4.0) // slower speed
}

// moveBezier moves from (x0,y0) to (x1,y1) along a quadratic bezier curve.
// speedFactor: higher = fewer steps = faster movement.
func moveBezier(x0, y0, x1, y1, speedFactor float64) {
	d := dist(x0, y0, x1, y1)
	if d < 1 {
		return
	}

	steps := int(d / speedFactor)
	if steps < 10 {
		steps = 10
	}
	if steps > 300 {
		steps = 300
	}

	// control point offset: random perpendicular push up to 20% of distance
	maxOffset := d * 0.20
	offset := point{
		x: (rand.Float64()*2 - 1) * maxOffset,
		y: (rand.Float64()*2 - 1) * maxOffset,
	}

	pts := quadraticBezier(x0, y0, x1, y1, offset, steps)
	stepDelay := time.Duration(1+rand.Intn(3)) * time.Millisecond

	for _, p := range pts {
		robotgo.Move(int(math.Round(p.x)), int(math.Round(p.y)))
		time.Sleep(stepDelay)
	}
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

package mover

import "math"

type point struct{ x, y float64 }

// smoothstep easing: ease-in-out
func ease(t float64) float64 {
	return t * t * (3 - 2*t)
}

// quadraticBezier returns evenly-spaced points along a quadratic bezier curve
// from (x0,y0) to (x1,y1) with a random control point for natural curvature.
func quadraticBezier(x0, y0, x1, y1 float64, controlOffset point, steps int) []point {
	if steps < 2 {
		steps = 2
	}
	// midpoint + perpendicular offset as control point
	cx := (x0+x1)/2 + controlOffset.x
	cy := (y0+y1)/2 + controlOffset.y

	pts := make([]point, steps)
	for i := 0; i < steps; i++ {
		t := ease(float64(i) / float64(steps-1))
		inv := 1 - t
		pts[i] = point{
			x: inv*inv*x0 + 2*inv*t*cx + t*t*x1,
			y: inv*inv*y0 + 2*inv*t*cy + t*t*y1,
		}
	}
	return pts
}

func dist(x0, y0, x1, y1 float64) float64 {
	dx, dy := x1-x0, y1-y0
	return math.Sqrt(dx*dx + dy*dy)
}

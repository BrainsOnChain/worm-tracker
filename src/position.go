package src

import (
	"math"
	"time"
)

type Position struct {
	ID        string    `json:"id"`
	X         float64   `json:"x"`
	Y         float64   `json:"y"`
	Direction float64   `json:"direction"`
	Price     float64   `json:"price"`
	Timestamp time.Time `json:"timestamp"`
}

func (p *Position) update(angle, magnitude float64) (float64, float64) {
	p.Direction += angle
	if p.Direction < 0 {
		p.Direction += 360
	} else if p.Direction >= 360 {
		p.Direction -= 360
	}

	dX, dY := magnitude*math.Cos(p.Direction*math.Pi/180), magnitude*math.Sin(p.Direction*math.Pi/180)

	// Update the position based on the Direction
	p.X += dX
	p.Y += dY

	return dX, dY
}

// movement outputs the movement in the form of angle and magnitude based on the
// left and right muscle activity.
func movement(left, right float64) (float64, float64) {
	// Calculate the angle and magnitude
	angle := (right - left) / 2
	magnitude := (right + left) / 2

	return angle, magnitude
}

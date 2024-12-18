package src

import (
	"math"
	"time"
)

type position struct {
	ID        int       `json:"id"` // set by the DB
	block     int       `json:"-"`  // the associated block number that contained the muscle movements
	X         float64   `json:"x"`
	Y         float64   `json:"y"`
	Direction float64   `json:"direction"`
	Price     float64   `json:"price"`
	Timestamp time.Time `json:"timestamp"`
}

// updatePosition takes the contract data and the current position to create a
// new position object.
func updatePosition(c contractData, cp position) position {

	angle := float64(c.rightMuscle-c.leftMuscle) / 2
	magnitude := float64(c.rightMuscle+c.leftMuscle) / 2

	newDirection := cp.Direction + angle
	if newDirection < 0 {
		newDirection += 360
	} else if newDirection >= 360 {
		newDirection -= 360
	}

	dX := magnitude * math.Cos(cp.Direction*math.Pi/180)
	dY := magnitude * math.Sin(cp.Direction*math.Pi/180)

	newX := cp.X + dX
	newY := cp.Y + dY

	np := position{
		block:     c.block,
		X:         newX,
		Y:         newY,
		Direction: newDirection,
		Price:     c.price,
		Timestamp: c.ts,
	}

	return np
}

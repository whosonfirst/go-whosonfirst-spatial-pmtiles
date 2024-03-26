package geo

import (
	"errors"

	"github.com/paulmach/orb"
)

func NewCoordinate(x float64, y float64) (*orb.Point, error) {

	if !IsValidLatitude(y) {
		return nil, errors.New("Invalid latitude")
	}

	if !IsValidLongitude(y) {
		return nil, errors.New("Invalid longitude")
	}

	coord := &orb.Point{x, y}
	return coord, nil
}

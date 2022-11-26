package geo

import (
	"errors"
	"github.com/paulmach/orb"
)

func NewBoundingBox(minx float64, miny float64, maxx float64, maxy float64) (*orb.Bound, error) {

	if !IsValidLongitude(minx) {
		return nil, errors.New("Invalid min longitude")
	}

	if !IsValidLatitude(miny) {
		return nil, errors.New("Invalid min latitude")
	}

	if !IsValidLongitude(maxx) {
		return nil, errors.New("Invalid max longitude")
	}

	if !IsValidLatitude(maxy) {
		return nil, errors.New("Invalid max latitude")
	}

	if minx > maxx {
		return nil, errors.New("Min lon is greater than max lon")
	}

	if minx > maxx {
		return nil, errors.New("Min latitude is greater than max latitude")
	}

	min_coord, err := NewCoordinate(minx, miny)

	if err != nil {
		return nil, err
	}

	max_coord, err := NewCoordinate(maxx, maxy)

	if err != nil {
		return nil, err
	}

	rect := &orb.Bound{
		Min: *min_coord,
		Max: *max_coord,
	}

	return rect, nil
}

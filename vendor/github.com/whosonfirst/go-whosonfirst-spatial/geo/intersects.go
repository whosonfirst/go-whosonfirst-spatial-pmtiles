package geo

import (
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/encoding/wkt"
	"github.com/peterstace/simplefeatures/geom"
)

// https://pkg.go.dev/github.com/peterstace/simplefeatures/geom#Intersects

func Intersects(g1 orb.Geometry, g2 orb.Geometry) (bool, error) {

	g1_wkt := wkt.MarshalString(g1)
	g2_wkt := wkt.MarshalString(g2)

	simple_g1, err := geom.UnmarshalWKT(g1_wkt)

	if err != nil {
		return false, err
	}

	simple_g2, err := geom.UnmarshalWKT(g2_wkt)

	if err != nil {
		return false, err
	}

	return geom.Intersects(simple_g1, simple_g2), nil
}

package geometry

import (
	"github.com/whosonfirst/go-whosonfirst-flags"
)

type NullAlternateGeometryFlag struct {
	flags.AlternateGeometryFlag
}

func NewNullAlternateGeometryFlag() (flags.AlternateGeometryFlag, error) {

	n := NullAlternateGeometryFlag{}
	return &n, nil
}

func (f *NullAlternateGeometryFlag) MatchesAny(others ...flags.AlternateGeometryFlag) bool {
	return true
}

func (f *NullAlternateGeometryFlag) MatchesAll(others ...flags.AlternateGeometryFlag) bool {
	return true
}

func (f *NullAlternateGeometryFlag) IsAlternateGeometry() bool {
	return false
}

func (f *NullAlternateGeometryFlag) Label() string {
	return ""
}

func (f *NullAlternateGeometryFlag) String() string {
	return "NULL"
}

package wkttoorb

import (
	"strings"

	"github.com/paulmach/orb"
)

func Scan(s string) (orb.Geometry, error) {
	p := Parser{NewLexer(strings.NewReader(s))}
	return p.Parse()
}

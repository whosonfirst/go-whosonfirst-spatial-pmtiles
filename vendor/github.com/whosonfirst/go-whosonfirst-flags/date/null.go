package date

import (
	"github.com/whosonfirst/go-whosonfirst-flags"
)

type NullDateFlag struct {
	flags.DateFlag
}

func NewNullDateFlag() (flags.DateFlag, error) {
	fl := NullDateFlag{}
	return &fl, nil
}

func (fl *NullDateFlag) InnerRange() (*int64, *int64) {
	return nil, nil
}

func (fl *NullDateFlag) OuterRange() (*int64, *int64) {
	return nil, nil
}

func (fl *NullDateFlag) MatchesAny(others ...flags.DateFlag) bool {
	return true
}

func (fl *NullDateFlag) MatchesAll(others ...flags.DateFlag) bool {
	return true
}

func (fl *NullDateFlag) String() string {
	return ""
}

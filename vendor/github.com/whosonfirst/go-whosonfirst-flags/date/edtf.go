package date

/*

This package is not stable yet. Specifically:

* There is no way to define custom ranges (inner, outer) or operations (contains, intersects)
* As of this writing everything is (outer, intersects)
* Open-ended queries (things like 2020/ or 2020/.. or /2019) don't work - this is caused by
  the describeRange method returning nil (if a edtf.Timestamp is nil) and the various methods
  for operating on dates returning nil in kind. This isn't a feature, it just hasn't been
  atteneded to yet

*/

import (
	"github.com/sfomuseum/go-edtf"
	"github.com/sfomuseum/go-edtf/parser"
	"github.com/whosonfirst/go-whosonfirst-flags"
	_ "log"
)

const CONTAINS int = 0
const INTERSECTS int = 1

const INNER int = 0
const OUTER int = 1

type EDTFDateFlag struct {
	flags.DateFlag
	date     *edtf.EDTFDate
	mode     int
	boundary int
}

func NewEDTFDateFlagsArray(names ...string) ([]flags.DateFlag, error) {

	pt_flags := make([]flags.DateFlag, 0)

	for _, name := range names {

		fl, err := NewEDTFDateFlag(name)

		if err != nil {
			return nil, err
		}

		pt_flags = append(pt_flags, fl)
	}

	return pt_flags, nil
}

func NewEDTFDateFlag(edtf_str string) (flags.DateFlag, error) {

	d, err := parser.ParseString(edtf_str)

	if err != nil {
		return nil, err
	}

	return NewEDTFDateFlagWithDate(d)
}

func NewEDTFDateFlagWithDate(d *edtf.EDTFDate) (flags.DateFlag, error) {

	fl := EDTFDateFlag{
		date:     d,
		mode:     INTERSECTS,
		boundary: OUTER,
	}

	return &fl, nil
}

func (fl *EDTFDateFlag) InnerRange() (*int64, *int64) {

	start_ts := fl.date.Start.Upper.Timestamp
	end_ts := fl.date.End.Lower.Timestamp

	return fl.describeRange(start_ts, end_ts)
}

func (fl *EDTFDateFlag) OuterRange() (*int64, *int64) {

	start_ts := fl.date.Start.Lower.Timestamp
	end_ts := fl.date.End.Upper.Timestamp

	return fl.describeRange(start_ts, end_ts)
}

func (fl *EDTFDateFlag) describeRange(start_ts *edtf.Timestamp, end_ts *edtf.Timestamp) (*int64, *int64) {

	if start_ts != nil && end_ts != nil {
		start := start_ts.Unix()
		end := end_ts.Unix()
		return &start, &end
	}

	if start_ts != nil {
		start := start_ts.Unix()
		return &start, nil
	}

	if end_ts != nil {
		end := end_ts.Unix()
		return nil, &end
	}

	return nil, nil
}

func (fl *EDTFDateFlag) MatchesAny(others ...flags.DateFlag) bool {

	for _, o := range others {

		if fl.matches(o) {
			return true
		}
	}

	return false
}

func (fl *EDTFDateFlag) MatchesAll(others ...flags.DateFlag) bool {

	matches := 0

	for _, o := range others {

		if fl.matches(o) {
			matches += 1
		}

	}

	if matches == len(others) {
		return true
	}

	return false
}

func (fl *EDTFDateFlag) matches(o flags.DateFlag) bool {

	switch fl.boundary {
	case INNER:
		switch fl.mode {
		case INTERSECTS:
			return fl.intersectsInner(o)
		default:
			return fl.containsInner(o)
		}
	default:
		switch fl.mode {
		case INTERSECTS:
			return fl.intersectsOuter(o)
		default:
			return fl.containsOuter(o)
		}
	}
}

func (fl *EDTFDateFlag) containsInner(o flags.DateFlag) bool {

	start, end := fl.InnerRange()
	o_start, o_end := o.InnerRange()

	if start == nil || end == nil {
		return false
	}

	if o_start == nil || o_end == nil {
		return false
	}

	return fl.contains(*start, *end, *o_start, *o_end)
}

func (fl *EDTFDateFlag) containsOuter(o flags.DateFlag) bool {

	start, end := fl.OuterRange()
	o_start, o_end := o.OuterRange()

	if start == nil || end == nil {
		return false
	}

	if o_start == nil || o_end == nil {
		return false
	}

	return fl.contains(*start, *end, *o_start, *o_end)
}

func (fl *EDTFDateFlag) contains(start int64, end int64, o_start int64, o_end int64) bool {

	if start > o_start {
		return false
	}

	if end < o_end {
		return false
	}

	return true
}

func (fl *EDTFDateFlag) intersectsInner(o flags.DateFlag) bool {

	start, end := fl.InnerRange()
	o_start, o_end := o.InnerRange()

	if start == nil || end == nil {
		return false
	}

	if o_start == nil || o_end == nil {
		return false
	}

	return fl.intersects(*start, *end, *o_start, *o_end)
}

func (fl *EDTFDateFlag) intersectsOuter(o flags.DateFlag) bool {

	start, end := fl.OuterRange()
	o_start, o_end := o.OuterRange()

	if start == nil || end == nil {
		return false
	}

	if o_start == nil || o_end == nil {
		return false
	}

	return fl.intersects(*start, *end, *o_start, *o_end)
}

func (fl *EDTFDateFlag) intersects(start int64, end int64, o_start int64, o_end int64) bool {

	if o_start > end {
		return false
	}

	if o_end < start {
		return false
	}

	return true
}

func (fl *EDTFDateFlag) String() string {
	return fl.date.EDTF
}

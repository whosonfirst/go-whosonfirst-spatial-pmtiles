package document

import (
	"context"
	"fmt"
	_ "log"

	"github.com/sfomuseum/go-edtf/unix"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// AppendEDTFRanges appends numeric date ranges derived from `edtf:inception` and `edtf:cessation` properties
// to a Who's On First document.
func AppendEDTFRanges(ctx context.Context, body []byte) ([]byte, error) {

	props := gjson.ParseBytes(body)

	props_rsp := gjson.GetBytes(body, "properties")

	if props_rsp.Exists() {
		props = props_rsp
	}

	inception_range, err := deriveRanges(props, "edtf:inception")

	if err != nil {
		return nil, fmt.Errorf("Failed to derive inception ranges, %w", err)
	}

	cessation_range, err := deriveRanges(props, "edtf:cessation")

	if err != nil {
		return nil, fmt.Errorf("Failed to derive cessation ranges, %w", err)
	}

	to_assign := make(map[string]int64)

	if inception_range != nil {
		to_assign["date:inception_inner_start"] = inception_range.Inner.Start
		to_assign["date:inception_inner_end"] = inception_range.Inner.End
		to_assign["date:inception_outer_start"] = inception_range.Outer.Start
		to_assign["date:inception_outer_end"] = inception_range.Outer.End
	}

	if cessation_range != nil {
		to_assign["date:cessation_inner_start"] = cessation_range.Inner.Start
		to_assign["date:cessation_inner_end"] = cessation_range.Inner.End
		to_assign["date:cessation_outer_start"] = cessation_range.Outer.Start
		to_assign["date:cessation_outer_end"] = cessation_range.Outer.End
	}

	for k, v := range to_assign {

		path := k

		if props_rsp.Exists() {
			path = fmt.Sprintf("properties.%s", k)
		}

		body, err = sjson.SetBytes(body, path, v)

		if err != nil {
			return nil, fmt.Errorf("Failed to assign %s (%d), %w", path, v, err)
		}
	}

	return body, nil
}

func deriveRanges(props gjson.Result, path string) (*unix.DateRange, error) {

	edtf_rsp := props.Get(path)

	if !edtf_rsp.Exists() {
		return nil, nil
	}

	edtf_str := edtf_rsp.String()

	derived, date_range, err := unix.DeriveRanges(edtf_str)

	if err != nil {
		return nil, err
	}

	if !derived {
		return nil, nil
	}

	return date_range, nil
}

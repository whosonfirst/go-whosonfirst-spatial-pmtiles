package filter

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/whosonfirst/go-whosonfirst-spatial"
)

// NewSPRFilterFromQuery will instantiate new `spatial.Filter` instance derived from parameters
// in 'query'.
func NewSPRFilterFromQuery(query url.Values) (spatial.Filter, error) {

	inputs, err := NewSPRInputs()

	if err != nil {
		return nil, err
	}

	inputs.Placetypes = query["placetype"]
	inputs.Geometries = query["geometries"]
	inputs.AlternateGeometries = query["alternate_geometry"]

	inputs.InceptionDate = query.Get("inception_date")
	inputs.CessationDate = query.Get("cessation_date")

	is_current, err := atoi(query["is_current"])

	if err != nil {
		return nil, fmt.Errorf("Invalid ?is_current= parameter, %w", err)
	}

	is_deprecated, err := atoi(query["is_deprecated"])

	if err != nil {
		return nil, fmt.Errorf("Invalid ?is_deprecated= parameter, %w", err)
	}

	is_ceased, err := atoi(query["is_ceased"])

	if err != nil {
		return nil, fmt.Errorf("Invalid ?is_ceased= parameter, %w", err)
	}

	is_superseded, err := atoi(query["is_superseded"])

	if err != nil {
		return nil, fmt.Errorf("Invalid ?is_superseded= parameter, %w", err)
	}

	is_superseding, err := atoi(query["is_superseding"])

	if err != nil {
		return nil, fmt.Errorf("Invalid ?is_superseding= parameter, %w", err)
	}

	inputs.IsCurrent = is_current
	inputs.IsDeprecated = is_deprecated
	inputs.IsCeased = is_ceased
	inputs.IsSuperseded = is_superseded
	inputs.IsSuperseding = is_superseding

	return NewSPRFilterFromInputs(inputs)
}

func atoi(strings []string) ([]int64, error) {

	numbers := make([]int64, len(strings))

	for idx, str := range strings {

		i, err := strconv.ParseInt(str, 10, 64)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse integer value at offset %d, %w", idx, err)
		}

		numbers[idx] = i
	}

	return numbers, nil
}

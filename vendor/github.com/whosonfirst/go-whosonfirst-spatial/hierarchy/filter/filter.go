package filter

// To do: Update to use aaronland/go-roster so these can be defined with a URI-based syntax
import (
	"context"
	"fmt"

	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
)

// FilterSPRResultsFunc defines a custom function for deriving a single `spr.StandardPlacesResult` from a list of `spr.StandardPlacesResult` instances.
type FilterSPRResultsFunc func(context.Context, reader.Reader, []byte, []spr.StandardPlacesResult) (spr.StandardPlacesResult, error)

// FirstButForgivingSPRResultsFunc returns the first record in 'possible' or nil.
func FirstButForgivingSPRResultsFunc(ctx context.Context, r reader.Reader, body []byte, possible []spr.StandardPlacesResult) (spr.StandardPlacesResult, error) {

	if len(possible) == 0 {
		return nil, nil
	}

	parent_spr := possible[0]
	return parent_spr, nil
}

// FirstSPRResultsFunc returns the first record in 'possible' or an error.
func FirstSPRResultsFunc(ctx context.Context, r reader.Reader, body []byte, possible []spr.StandardPlacesResult) (spr.StandardPlacesResult, error) {

	if len(possible) == 0 {
		return nil, fmt.Errorf("No results")
	}

	parent_spr := possible[0]
	return parent_spr, nil
}

// FirstSPRResultsFunc returns the first record in 'possible' unless there are multiple results in which can an error is returned.
func SingleSPRResultsFunc(ctx context.Context, r reader.Reader, body []byte, possible []spr.StandardPlacesResult) (spr.StandardPlacesResult, error) {

	if len(possible) != 1 {
		return nil, fmt.Errorf("Number of results != 1")
	}

	parent_spr := possible[0]
	return parent_spr, nil
}

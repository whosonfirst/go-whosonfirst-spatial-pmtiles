// Package sort provides interfaces for sorting `spr.StandardPlacesResults` instances
package sort

import (
	"context"
	"fmt"
	"github.com/aaronland/go-roster"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
	"net/url"
	"sort"
	"strings"
)

// SortedStandardPlacesResults implements the `spr.StandardPlacesResults` interface for sorted results.
type SortedStandardPlacesResults struct {
	spr.StandardPlacesResults `json:",omitempty"`
	Places []spr.StandardPlacesResult `json:"places"`
}

// Results returns a list of `spr.StandardPlacesResults` instances.
func (r *SortedStandardPlacesResults) Results() []spr.StandardPlacesResult {
	return r.Places
}

func NewSortedStandardPlacesResults(places []spr.StandardPlacesResult) spr.StandardPlacesResults {
	return &SortedStandardPlacesResults{
		Places: places,
	}
}

// Sorter provides an interface for sorting `spr.StandardPlacesResults` instances
type Sorter interface {
	// Sort sorts a `spr.StandardPlacesResults` instance according to rules defined by the interface implementation.
	Sort(context.Context, spr.StandardPlacesResults, ...Sorter) (spr.StandardPlacesResults, error)
}

var sorter_roster roster.Roster

// SorterInitializationFunc is a function defined by individual sorter package and used to create
// an instance of that sorter
type SorterInitializationFunc func(ctx context.Context, uri string) (Sorter, error)

// RegisterSorter registers 'scheme' as a key pointing to 'init_func' in an internal lookup table
// used to create new `Sorter` instances by the `NewSorter` method.
func RegisterSorter(ctx context.Context, scheme string, init_func SorterInitializationFunc) error {

	err := ensureSorterRoster()

	if err != nil {
		return err
	}

	return sorter_roster.Register(ctx, scheme, init_func)
}

func ensureSorterRoster() error {

	if sorter_roster == nil {

		r, err := roster.NewDefaultRoster()

		if err != nil {
			return err
		}

		sorter_roster = r
	}

	return nil
}

// NewSorter returns a new `Sorter` instance configured by 'uri'. The value of 'uri' is parsed
// as a `url.URL` and its scheme is used as the key for a corresponding `SorterInitializationFunc`
// function used to instantiate the new `Sorter`. It is assumed that the scheme (and initialization
// function) have been registered by the `RegisterSorter` method.
func NewSorter(ctx context.Context, uri string) (Sorter, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	scheme := u.Scheme

	i, err := sorter_roster.Driver(ctx, scheme)

	if err != nil {
		return nil, err
	}

	init_func := i.(SorterInitializationFunc)
	return init_func(ctx, uri)
}

// Schemes returns the list of schemes that have been registered.
func Schemes() []string {

	ctx := context.Background()
	schemes := []string{}

	err := ensureSorterRoster()

	if err != nil {
		return schemes
	}

	for _, dr := range sorter_roster.Drivers(ctx) {
		scheme := fmt.Sprintf("%s://", strings.ToLower(dr))
		schemes = append(schemes, scheme)
	}

	sort.Strings(schemes)
	return schemes
}

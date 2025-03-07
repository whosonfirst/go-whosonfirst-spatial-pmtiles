package query

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/aaronland/go-roster"
	"github.com/paulmach/orb"
	"github.com/whosonfirst/go-whosonfirst-spatial"
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
)

// SpatialFunction in an interface for spatial functions
type SpatialFunction interface {
	// Execute performs a specific spatial function
	Execute(context.Context, database.SpatialDatabase, orb.Geometry, ...spatial.Filter) (spr.StandardPlacesResults, error)
}

var function_roster roster.Roster

// SpatialFunctionInitializationFunc is a function defined by individual function package and used to create
// an instance of that function
type SpatialFunctionInitializationFunc func(ctx context.Context, uri string) (SpatialFunction, error)

// RegisterSpatialFunction registers 'scheme' as a key pointing to 'init_func' in an internal lookup table
// used to create new `SpatialFunction` instances by the `NewSpatialFunction` method.
func RegisterSpatialFunction(ctx context.Context, scheme string, init_func SpatialFunctionInitializationFunc) error {

	err := ensureSpatialFunctionRoster()

	if err != nil {
		return err
	}

	return function_roster.Register(ctx, scheme, init_func)
}

func ensureSpatialFunctionRoster() error {

	if function_roster == nil {

		r, err := roster.NewDefaultRoster()

		if err != nil {
			return err
		}

		function_roster = r
	}

	return nil
}

// NewSpatialFunction returns a new `SpatialFunction` instance configured by 'uri'. The value of 'uri' is parsed
// as a `url.URL` and its scheme is used as the key for a corresponding `SpatialFunctionInitializationFunc`
// function used to instantiate the new `SpatialFunction`. It is assumed that the scheme (and initialization
// function) have been registered by the `RegisterSpatialFunction` method.
func NewSpatialFunction(ctx context.Context, uri string) (SpatialFunction, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	scheme := u.Scheme

	i, err := function_roster.Driver(ctx, scheme)

	if err != nil {
		return nil, err
	}

	init_func := i.(SpatialFunctionInitializationFunc)
	return init_func(ctx, uri)
}

// SpatialFunctionSchemes returns the list of schemes that have been registered.
func SpatialFunctionSchemes() []string {

	ctx := context.Background()
	schemes := []string{}

	err := ensureSpatialFunctionRoster()

	if err != nil {
		return schemes
	}

	for _, dr := range function_roster.Drivers(ctx) {
		scheme := fmt.Sprintf("%s://", strings.ToLower(dr))
		schemes = append(schemes, scheme)
	}

	sort.Strings(schemes)
	return schemes
}

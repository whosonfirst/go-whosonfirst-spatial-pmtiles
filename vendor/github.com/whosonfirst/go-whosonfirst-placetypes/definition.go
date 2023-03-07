package placetypes

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/aaronland/go-roster"
)

// type Definition provides an interface for working with "core" and custom placetype specifications
// and for creating them using a URI-based syntax
type Definition interface {
	// Specification returns the `WOFPlacetypeSpecification` instance associated with the definition
	Specification() *WOFPlacetypeSpecification
	// Property return the relative (base) property name whose values (placetypes) are associated with the definition's placetype specification
	Property() string
	// URI() returns the URI used to create the definition
	URI() string
}

var definition_roster roster.Roster

// DefinitionInitializationFunc is a function defined by individual definition package and used to create
// an instance of that definition
type DefinitionInitializationFunc func(ctx context.Context, uri string) (Definition, error)

// RegisterDefinition registers 'scheme' as a key pointing to 'init_func' in an internal lookup table
// used to create new `Definition` instances by the `NewDefinition` method.
func RegisterDefinition(ctx context.Context, scheme string, init_func DefinitionInitializationFunc) error {

	err := ensureDefinitionRoster()

	if err != nil {
		return err
	}

	return definition_roster.Register(ctx, scheme, init_func)
}

func ensureDefinitionRoster() error {

	if definition_roster == nil {

		r, err := roster.NewDefaultRoster()

		if err != nil {
			return err
		}

		definition_roster = r
	}

	return nil
}

// NewDefinition returns a new `Definition` instance configured by 'uri'. The value of 'uri' is parsed
// as a `url.URL` and its scheme is used as the key for a corresponding `DefinitionInitializationFunc`
// function used to instantiate the new `Definition`. It is assumed that the scheme (and initialization
// function) have been registered by the `RegisterDefinition` method.
func NewDefinition(ctx context.Context, uri string) (Definition, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	scheme := u.Scheme

	i, err := definition_roster.Driver(ctx, scheme)

	if err != nil {
		return nil, err
	}

	init_func := i.(DefinitionInitializationFunc)
	return init_func(ctx, uri)
}

// Schemes returns the list of schemes that have been registered.
func Schemes() []string {

	ctx := context.Background()
	schemes := []string{}

	err := ensureDefinitionRoster()

	if err != nil {
		return schemes
	}

	for _, dr := range definition_roster.Drivers(ctx) {
		scheme := fmt.Sprintf("%s://", strings.ToLower(dr))
		schemes = append(schemes, scheme)
	}

	sort.Strings(schemes)
	return schemes
}

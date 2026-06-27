package iterate

import (
	"context"
	"fmt"
	"iter"
	"net/url"
	"sort"
	"strings"

	"github.com/aaronland/go-roster"
)

// Iterator defines an interface for iterating through collections  of Who's On First documents.
type Iterator interface {
	// Iterate will return an `iter.Seq2[*Record, error]` for each record encountered in one or more URIs.
	Iterate(context.Context, ...string) iter.Seq2[*Record, error]
	// Seen() returns the total number of records processed so far.
	Seen() int64
	// IsIterating() returns a boolean value indicating whether 'it' is still processing documents.
	IsIterating() bool
	// Close performs any implementation specific tasks before terminating the iterator.
	Close() error
}

// IteratorInitializationFunc is a function defined by individual iterator package and used to create
// an instance of that iterator
type IteratorInitializationFunc func(ctx context.Context, uri string) (Iterator, error)

// iterators is a `aaronland/go-roster.Roster` instance used to maintain a list of registered `IteratorInitializationFunc` initialization functions.
var iterators roster.Roster

// RegisterIterator() associates 'scheme' with 'init_func' in an internal list of avilable `Iterator` implementations.
func RegisterIterator(ctx context.Context, scheme string, f IteratorInitializationFunc) error {

	err := ensureRoster()

	if err != nil {
		return fmt.Errorf("Failed to register %s scheme, %w", scheme, err)
	}

	return iterators.Register(ctx, scheme, f)
}

// NewIterator() returns a new `Iterator` instance derived from 'uri'. The semantics of and requirements for
// 'uri' as specific to the package implementing the interface.
func NewIterator(ctx context.Context, uri string) (Iterator, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	scheme := u.Scheme

	if scheme == "" {
		return nil, fmt.Errorf("Emittter URI is missing scheme '%s'", uri)
	}

	err = ensureRoster()

	if err != nil {
		return nil, fmt.Errorf("Failed to register %s scheme, %w", scheme, err)
	}

	i, err := iterators.Driver(ctx, scheme)

	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve driver for '%s' scheme, %w", scheme, err)
	}

	fn := i.(IteratorInitializationFunc)

	if fn == nil {
		return nil, fmt.Errorf("Unregistered initialization function for '%s' scheme", scheme)
	}

	if fn == nil {
		return nil, fmt.Errorf("Undefined initialization function")
	}

	it, err := fn(ctx, uri)

	if err != nil {
		return nil, err
	}

	return NewConcurrentIterator(ctx, uri, it)
}

// IteratorSchemes() returns the list of schemes that have been "registered".
func IteratorSchemes() []string {

	ctx := context.Background()
	schemes := []string{}

	err := ensureRoster()

	if err != nil {
		return schemes
	}

	for _, dr := range iterators.Drivers(ctx) {
		scheme := fmt.Sprintf("%s://", strings.ToLower(dr))
		schemes = append(schemes, scheme)
	}

	sort.Strings(schemes)
	return schemes
}

// ensureRoster() ensures that a `aaronland/go-roster.Roster` instance used to maintain a list of registered `IteratorInitializationFunc`
// initialization functions is present
func ensureRoster() error {

	if iterators == nil {

		r, err := roster.NewDefaultRoster()

		if err != nil {
			return fmt.Errorf("Failed to create new roster, %w", err)
		}

		iterators = r
	}

	return nil
}

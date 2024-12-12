package cache

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/aaronland/go-roster"
)

type CacheManager interface {
	CacheFeature(context.Context, []byte) (*FeatureCache, error)
	GetFeatureCache(context.Context, string) (*FeatureCache, error)
	Close() error
}

var cache_manager_roster roster.Roster

// CacheManagerInitializationFunc is a function defined by individual cache_manager package and used to create
// an instance of that cache_manager
type CacheManagerInitializationFunc func(ctx context.Context, uri string) (CacheManager, error)

// RegisterCacheManager registers 'scheme' as a key pointing to 'init_func' in an internal lookup table
// used to create new `CacheManager` instances by the `NewCacheManager` method.
func RegisterCacheManager(ctx context.Context, scheme string, init_func CacheManagerInitializationFunc) error {

	err := ensureCacheManagerRoster()

	if err != nil {
		return err
	}

	return cache_manager_roster.Register(ctx, scheme, init_func)
}

func ensureCacheManagerRoster() error {

	if cache_manager_roster == nil {

		r, err := roster.NewDefaultRoster()

		if err != nil {
			return err
		}

		cache_manager_roster = r
	}

	return nil
}

// NewCacheManager returns a new `CacheManager` instance configured by 'uri'. The value of 'uri' is parsed
// as a `url.URL` and its scheme is used as the key for a corresponding `CacheManagerInitializationFunc`
// function used to instantiate the new `CacheManager`. It is assumed that the scheme (and initialization
// function) have been registered by the `RegisterCacheManager` method.
func NewCacheManager(ctx context.Context, uri string) (CacheManager, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	scheme := u.Scheme

	i, err := cache_manager_roster.Driver(ctx, scheme)

	if err != nil {
		return nil, err
	}

	init_func := i.(CacheManagerInitializationFunc)
	return init_func(ctx, uri)
}

// Schemes returns the list of schemes that have been registered.
func CacheManagerSchemes() []string {

	ctx := context.Background()
	schemes := []string{}

	err := ensureCacheManagerRoster()

	if err != nil {
		return schemes
	}

	for _, dr := range cache_manager_roster.Drivers(ctx) {
		scheme := fmt.Sprintf("%s://", strings.ToLower(dr))
		schemes = append(schemes, scheme)
	}

	sort.Strings(schemes)
	return schemes
}

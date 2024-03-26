package database

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/aaronland/go-roster"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-spatial"
	"github.com/whosonfirst/go-writer/v3"
)

type SpatialDatabase interface {
	reader.Reader
	writer.Writer
	spatial.SpatialIndex
}

type SpatialDatabaseInitializeFunc func(ctx context.Context, uri string) (SpatialDatabase, error)

var spatial_databases roster.Roster

func ensureSpatialRoster() error {

	if spatial_databases == nil {

		r, err := roster.NewDefaultRoster()

		if err != nil {
			return err
		}

		spatial_databases = r
	}

	return nil
}

func RegisterSpatialDatabase(ctx context.Context, scheme string, f SpatialDatabaseInitializeFunc) error {

	err := ensureSpatialRoster()

	if err != nil {
		return err
	}

	return spatial_databases.Register(ctx, scheme, f)
}

func Schemes() []string {

	ctx := context.Background()
	schemes := []string{}

	err := ensureSpatialRoster()

	if err != nil {
		return schemes
	}

	for _, dr := range spatial_databases.Drivers(ctx) {
		scheme := fmt.Sprintf("%s://", strings.ToLower(dr))
		schemes = append(schemes, scheme)
	}

	sort.Strings(schemes)
	return schemes
}

func NewSpatialDatabase(ctx context.Context, uri string) (SpatialDatabase, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	scheme := u.Scheme

	i, err := spatial_databases.Driver(ctx, scheme)

	if err != nil {
		return nil, err
	}

	f := i.(SpatialDatabaseInitializeFunc)
	return f(ctx, uri)
}

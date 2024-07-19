package placetypes

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/whosonfirst/go-whosonfirst-flags"
	wof_placetypes "github.com/whosonfirst/go-whosonfirst-placetypes"
)

// placetype_definitions is a local cache of go-whosonfirst-placetypes.Definition instances.
// It is used in by the NewPlacetypeFlag so we don't have to incur the overhead of populating
// placetype relationships multiple times.
var placetype_definitions *sync.Map

func init() {

	placetype_definitions = new(sync.Map)

	// Be proactive about fetching and caching "core" Who's On First placetypes

	go func() {
		ctx := context.Background()

		def_uri := "whosonfirst://"
		def, err := wof_placetypes.NewDefinition(ctx, def_uri)

		if err != nil {
			panic(err)
		}

		placetype_definitions.Store(def_uri, def)
	}()
}

type PlacetypeFlag struct {
	flags.PlacetypeFlag
	pt             *wof_placetypes.WOFPlacetype
	definition_uri string
}

func NewPlacetypeFlagsArray(names ...string) ([]flags.PlacetypeFlag, error) {

	pt_flags := make([]flags.PlacetypeFlag, 0)

	for _, name := range names {

		fl, err := NewPlacetypeFlag(name)

		if err != nil {
			return nil, fmt.Errorf("Failed to create new flag for '%s', %w", name, err)
		}

		pt_flags = append(pt_flags, fl)
	}

	return pt_flags, nil
}

func NewPlacetypeFlag(placetype_fl string) (flags.PlacetypeFlag, error) {

	logger := slog.Default()
	logger = logger.With("flag", placetype_fl)

	logger.Debug("Create new placetype flag")

	definition_uri := "whosonfirst://"
	placetype_name := placetype_fl

	parts := strings.Split(placetype_fl, "#")

	if len(parts) == 2 {
		placetype_name = parts[0]
		definition_uri = parts[1]
	}

	logger.Debug("Parse flag", "name", placetype_name, "definition", definition_uri)

	var placetype_def wof_placetypes.Definition

	v, exists := placetype_definitions.Load(definition_uri)

	if exists {
		placetype_def = v.(wof_placetypes.Definition)
	} else {

		ctx := context.Background()
		def, err := wof_placetypes.NewDefinition(ctx, definition_uri)

		if err != nil {
			return nil, fmt.Errorf("Failed to derive new foo, %w", err)
		}

		placetype_definitions.Store(definition_uri, def)
		placetype_def = def
	}

	placetype_spec := placetype_def.Specification()

	pt, err := placetype_spec.GetPlacetypeByName(placetype_name)

	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve placetype with name '%s', %w", placetype_name, err)
	}

	f := PlacetypeFlag{
		pt:             pt,
		definition_uri: definition_uri,
	}

	return &f, nil
}

func (f *PlacetypeFlag) MatchesAny(others ...flags.PlacetypeFlag) bool {

	logger := slog.Default()
	logger = logger.With("placetype", f.Placetype())
	logger = logger.With("definition", f.definition_uri)

	for _, o := range others {

		is_match := f.Placetype() == o.Placetype()
		logger.Debug("Test matches any", "other", o.Placetype(), "is match", is_match)

		if is_match {
			return true
		}

	}

	return false
}

func (f *PlacetypeFlag) MatchesAll(others ...flags.PlacetypeFlag) bool {

	logger := slog.Default()
	logger = logger.With("placetype", f.Placetype())
	logger = logger.With("definition", f.definition_uri)

	matches := 0

	for _, o := range others {

		is_match := f.Placetype() == o.Placetype()

		logger.Debug("Test matches any", "other", o.Placetype(), "is match", is_match)

		if is_match {
			matches += 1
		}

	}

	logger.Debug("Matches", "count", matches)

	if matches == len(others) {
		return true
	}

	return false
}

func (f *PlacetypeFlag) Placetype() string {
	return f.pt.Name
}

func (f *PlacetypeFlag) String() string {
	return f.Placetype()
}

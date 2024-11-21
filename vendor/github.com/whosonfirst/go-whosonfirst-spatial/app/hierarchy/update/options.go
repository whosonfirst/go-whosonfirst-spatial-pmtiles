package update

import (
	"context"
	"flag"

	"github.com/sfomuseum/go-flags/flagset"
	"github.com/whosonfirst/go-whosonfirst-export/v2"
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
	"github.com/whosonfirst/go-whosonfirst-spatial/filter"
	"github.com/whosonfirst/go-whosonfirst-spatial/hierarchy"
	hierarchy_filter "github.com/whosonfirst/go-whosonfirst-spatial/hierarchy/filter"
	"github.com/whosonfirst/go-writer/v3"
)

type RunOptions struct {
	Writer                writer.Writer
	WriterURI             string
	Exporter              export.Exporter
	ExporterURI           string
	MapshaperServerURI    string
	SpatialDatabase       database.SpatialDatabase
	SpatialDatabaseURI    string
	TargetIteratorSources map[string][]string
	SourceIteratorSources map[string][]string
	SPRFilterInputs       *filter.SPRInputs
	SPRResultsFunc        hierarchy_filter.FilterSPRResultsFunc                   // This one chooses one result among many (or nil)
	PIPUpdateFunc         hierarchy.PointInPolygonHierarchyResolverUpdateCallback // This one constructs a map[string]interface{} to update the target record (or not)
}

func RunOptionsFromFlagSet(ctx context.Context, fs *flag.FlagSet) (*RunOptions, error) {

	flagset.Parse(fs)

	inputs := &filter.SPRInputs{}

	inputs.IsCurrent = is_current
	inputs.IsCeased = is_ceased
	inputs.IsDeprecated = is_deprecated
	inputs.IsSuperseded = is_superseded
	inputs.IsSuperseding = is_superseding

	opts := &RunOptions{
		WriterURI:          writer_uri,
		ExporterURI:        exporter_uri,
		SpatialDatabaseURI: spatial_database_uri,
		MapshaperServerURI: mapshaper_server,
		SPRResultsFunc:     hierarchy_filter.FirstButForgivingSPRResultsFunc, // sudo make me configurable
		SPRFilterInputs:    inputs,
	}

	if len(source_iterator_uris) > 0 {
		opts.SourceIteratorSources = source_iterator_uris.AsMap()
	}

	if len(target_iterator_uris) > 0 {
		opts.TargetIteratorSources = target_iterator_uris.AsMap()
	}

	return opts, nil
}

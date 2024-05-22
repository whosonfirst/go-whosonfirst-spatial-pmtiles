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
	Writer             writer.Writer
	WriterURI          string
	Exporter           export.Exporter
	ExporterURI        string
	MapshaperServerURI string
	SpatialDatabase    database.SpatialDatabase
	SpatialDatabaseURI string
	ToIterator         string
	FromIterator       string
	SPRFilterInputs    *filter.SPRInputs
	SPRResultsFunc     hierarchy_filter.FilterSPRResultsFunc                   // This one chooses one result among many (or nil)
	PIPUpdateFunc      hierarchy.PointInPolygonHierarchyResolverUpdateCallback // This one constructs a map[string]interface{} to update the target record (or not)
	To                 []string
	From               []string
}

func RunOptionsFromFlagSet(ctx context.Context, fs *flag.FlagSet) (*RunOptions, error) {

	flagset.Parse(fs)

	inputs := &filter.SPRInputs{}

	inputs.IsCurrent = is_current
	inputs.IsCeased = is_ceased
	inputs.IsDeprecated = is_deprecated
	inputs.IsSuperseded = is_superseded
	inputs.IsSuperseding = is_superseding

	hierarchy_paths := fs.Args()

	opts := &RunOptions{
		WriterURI:          writer_uri,
		ExporterURI:        exporter_uri,
		SpatialDatabaseURI: spatial_database_uri,
		MapshaperServerURI: mapshaper_server,
		SPRResultsFunc:     hierarchy_filter.FirstButForgivingSPRResultsFunc, // sudo make me configurable
		SPRFilterInputs:    inputs,
		ToIterator:         iterator_uri,
		FromIterator:       spatial_iterator_uri,
		To:                 hierarchy_paths,
		From:               spatial_paths,
	}

	return opts, nil
}

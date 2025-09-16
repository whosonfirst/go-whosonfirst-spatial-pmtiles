package tables

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"embed"
	"fmt"
	"slices"
	"text/template"

	database_sql "github.com/sfomuseum/go-database/sql"
)

const index_alt_all string = "*"

//go:embed *.schema
var fs embed.FS

func LoadSchema(db *sql.DB, table_name string) (string, error) {

	driver := database_sql.Driver(db)

	fname := fmt.Sprintf("%s.%s.schema", table_name, driver)

	data, err := fs.ReadFile(fname)

	if err != nil {
		return "", fmt.Errorf("Failed to read %s, %w", fname, err)
	}

	t, err := template.New(table_name).Parse(string(data))

	if err != nil {
		return "", fmt.Errorf("Failed to parse %s template, %w", fname, err)
	}

	vars := struct {
		Name string
	}{
		Name: table_name,
	}

	var buf bytes.Buffer
	wr := bufio.NewWriter(&buf)

	err = t.Execute(wr, vars)

	if err != nil {
		return "", fmt.Errorf("Failed to process %s template, %w", fname, err)
	}

	wr.Flush()

	return buf.String(), nil
}

type InitTablesOptions struct {
	// Initialize the rtree table
	RTree bool
	// Initialize the geojson table
	GeoJSON bool
	// Initialize the properties table
	Properties bool
	// Initialize the spr table
	SPR bool
	// Initialize the spelunker table
	Spelunker bool
	// Initialize the concordances table
	Concordances bool
	// Initialize the ancestors table
	Ancestors bool
	// Initialize the search table
	Search bool
	// Initialize the names table
	Names bool
	// Initialize the supersedes table
	Supersedes bool
	// Initialize the geometries table
	Geometries bool
	// Initialize tables necessary for whosonfirst/go-whosonfirst-spatial operations
	SpatialTables bool
	// Initialize tables necessary for the Who's On First spelunker
	SpelunkerTables bool
	// Initialize all tables
	All bool
	// Zero or more table names where alt geometry files should be indexed.
	IndexAlt []string
	// Be strict when indexing alt geometries
	StrictAltFiles bool
}

func InitTables(ctx context.Context, db *sql.DB, opts *InitTablesOptions) ([]database_sql.Table, error) {

	if opts.SpatialTables {
		opts.RTree = true
		opts.GeoJSON = true
		opts.Properties = true
		opts.SPR = true
	}

	if opts.SpelunkerTables {
		opts.SPR = true
		opts.Spelunker = true
		opts.GeoJSON = true
		opts.Concordances = true
		opts.Ancestors = true
		opts.Search = true

		to_create_alt := []string{
			GEOJSON_TABLE_NAME,
		}

		for _, table_name := range to_create_alt {

			if !slices.Contains(opts.IndexAlt, table_name) {
				opts.IndexAlt = append(opts.IndexAlt, table_name)
			}
		}

	}

	db_driver := database_sql.Driver(db)

	switch db_driver {
	case database_sql.POSTGRES_DRIVER:

		if opts.SpatialTables {
			opts.RTree = false
			opts.Geometries = true
		}

		if opts.RTree {
			return nil, fmt.Errorf("-rtree table not supported by the %s driver", db_driver)
		}

		if opts.Search {
			return nil, fmt.Errorf("-search table not (yet) supported by the %s driver", db_driver)
		}
	}

	tables_list := make([]database_sql.Table, 0)

	if opts.GeoJSON || opts.All {

		geojson_opts, err := DefaultGeoJSONTableOptions()

		if err != nil {
			return nil, fmt.Errorf("failed to create '%s' table options because %s", GEOJSON_TABLE_NAME, err)
		}

		if slices.Contains(opts.IndexAlt, GEOJSON_TABLE_NAME) || slices.Contains(opts.IndexAlt, index_alt_all) {
			geojson_opts.IndexAltFiles = true
		}

		gt, err := NewGeoJSONTableWithDatabaseAndOptions(ctx, db, geojson_opts)

		if err != nil {
			return nil, fmt.Errorf("failed to create '%s' table because %s", GEOJSON_TABLE_NAME, err)
		}

		tables_list = append(tables_list, gt)
	}

	if opts.Supersedes || opts.All {

		t, err := NewSupersedesTableWithDatabase(ctx, db)

		if err != nil {
			return nil, fmt.Errorf("failed to create '%s' table because %s", SUPERSEDES_TABLE_NAME, err)
		}

		tables_list = append(tables_list, t)
	}

	if opts.RTree || opts.All {

		rtree_opts, err := DefaultRTreeTableOptions()

		if err != nil {
			return nil, fmt.Errorf("failed to create 'rtree' table options because %s", err)
		}

		if slices.Contains(opts.IndexAlt, RTREE_TABLE_NAME) || slices.Contains(opts.IndexAlt, index_alt_all) {
			rtree_opts.IndexAltFiles = true
		}

		gt, err := NewRTreeTableWithDatabaseAndOptions(ctx, db, rtree_opts)

		if err != nil {
			return nil, fmt.Errorf("failed to create 'rtree' table because %s", err)
		}

		tables_list = append(tables_list, gt)
	}

	if opts.Properties || opts.All {

		properties_opts, err := DefaultPropertiesTableOptions()

		if err != nil {
			return nil, fmt.Errorf("failed to create 'properties' table options because %s", err)
		}

		if slices.Contains(opts.IndexAlt, PROPERTIES_TABLE_NAME) || slices.Contains(opts.IndexAlt, index_alt_all) {
			properties_opts.IndexAltFiles = true
		}

		gt, err := NewPropertiesTableWithDatabaseAndOptions(ctx, db, properties_opts)

		if err != nil {
			return nil, fmt.Errorf("failed to create 'properties' table because %s", err)
		}

		tables_list = append(tables_list, gt)
	}

	if opts.SPR || opts.All {

		spr_opts, err := DefaultSPRTableOptions()

		if err != nil {
			return nil, fmt.Errorf("Failed to create '%s' table options because %v", SPR_TABLE_NAME, err)
		}

		if slices.Contains(opts.IndexAlt, SPR_TABLE_NAME) || slices.Contains(opts.IndexAlt, index_alt_all) {
			spr_opts.IndexAltFiles = true
		}

		st, err := NewSPRTableWithDatabaseAndOptions(ctx, db, spr_opts)

		if err != nil {
			return nil, fmt.Errorf("failed to create '%s' table because %s", SPR_TABLE_NAME, err)
		}

		tables_list = append(tables_list, st)
	}

	if opts.Spelunker || opts.All {

		spelunker_opts, err := DefaultSpelunkerTableOptions()

		if err != nil {
			return nil, fmt.Errorf("Failed to create '%s' table options because %v", SPELUNKER_TABLE_NAME, err)
		}

		if slices.Contains(opts.IndexAlt, SPELUNKER_TABLE_NAME) || slices.Contains(opts.IndexAlt, index_alt_all) {
			spelunker_opts.IndexAltFiles = true
		}

		st, err := NewSpelunkerTableWithDatabaseAndOptions(ctx, db, spelunker_opts)

		if err != nil {
			return nil, fmt.Errorf("failed to create '%s' table because %s", SPELUNKER_TABLE_NAME, err)
		}

		tables_list = append(tables_list, st)
	}

	if opts.Names || opts.All {

		nm, err := NewNamesTableWithDatabase(ctx, db)

		if err != nil {
			return nil, fmt.Errorf("failed to create '%s' table because %s", NAMES_TABLE_NAME, err)
		}

		tables_list = append(tables_list, nm)
	}

	if opts.Ancestors || opts.All {

		an, err := NewAncestorsTableWithDatabase(ctx, db)

		if err != nil {
			return nil, fmt.Errorf("failed to create '%s' table because %s", ANCESTORS_TABLE_NAME, err)
		}

		tables_list = append(tables_list, an)
	}

	if opts.Concordances || opts.All {

		cn, err := NewConcordancesTableWithDatabase(ctx, db)

		if err != nil {
			return nil, fmt.Errorf("failed to create '%s' table because %s", CONCORDANCES_TABLE_NAME, err)
		}

		tables_list = append(tables_list, cn)
	}

	// see the way we don't check all here - that's so people who don't have
	// spatialite installed can still use all (20180122/thisisaaronland)

	if opts.Geometries {

		geometries_opts, err := DefaultGeometriesTableOptions()

		if err != nil {
			return nil, fmt.Errorf("failed to create '%s' table options because %v", GEOMETRIES_TABLE_NAME, err)
		}

		if slices.Contains(opts.IndexAlt, CONCORDANCES_TABLE_NAME) || slices.Contains(opts.IndexAlt, index_alt_all) {
			geometries_opts.IndexAltFiles = true
		}

		gm, err := NewGeometriesTableWithDatabaseAndOptions(ctx, db, geometries_opts)

		if err != nil {
			return nil, fmt.Errorf("failed to create '%s' table because %v", CONCORDANCES_TABLE_NAME, err)
		}

		tables_list = append(tables_list, gm)
	}

	// see the way we don't check all here either - that's because this table can be
	// brutally slow to index and should probably really just be a separate database
	// anyway... (20180214/thisisaaronland)

	if opts.Search {

		// ALT FILES...

		st, err := NewSearchTableWithDatabase(ctx, db)

		if err != nil {
			return nil, fmt.Errorf("failed to create 'search' table because %v", err)
		}

		tables_list = append(tables_list, st)
	}

	if len(tables_list) == 0 {
		return nil, fmt.Errorf("You forgot to specify which (any) tables to index")
	}

	return tables_list, nil
}

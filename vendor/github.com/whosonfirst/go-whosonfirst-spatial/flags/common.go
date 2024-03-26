package flags

import (
	"flag"
	"fmt"

	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-flags/lookup"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-spatial/database"
)

func CommonFlags() (*flag.FlagSet, error) {

	fs := flagset.NewFlagSet("common")

	err := AppendCommonFlags(fs)

	if err != nil {
		return nil, fmt.Errorf("Failed to append common application flags, %w", err)
	}

	return fs, nil
}

func AppendCommonFlags(fs *flag.FlagSet) error {

	// spatial databases

	available_databases := database.Schemes()
	desc_databases := fmt.Sprintf("A valid whosonfirst/go-whosonfirst-spatial/data.SpatialDatabase URI. options are: %s", available_databases)

	fs.String(SpatialDatabaseURIFlag, "", desc_databases)

	available_readers := reader.Schemes()
	desc_readers := fmt.Sprintf("A valid whosonfirst/go-reader.Reader URI. Available options are: %s", available_readers)

	fs.String(PropertiesReaderURIFlag, "", fmt.Sprintf("%s. If the value is {spatial-database-uri} then the value of the '-spatial-database-uri' implements the reader.Reader interface and will be used.", desc_readers))

	fs.Bool(IS_WOF, true, "Input data is WOF-flavoured GeoJSON. (Pass a value of '0' or 'false' if you need to index non-WOF documents.")

	fs.Bool(EnableCustomPlacetypesFlag, false, "Enable wof:placetype values that are not explicitly defined in the whosonfirst/go-whosonfirst-placetypes repository.")

	// Pending changes in the app/placetypes.go package to support
	// alternate sources (20210324/thisisaaronland)
	// fs.String(CUSTOM_PLACETYPES_SOURCE, "", "...")

	fs.String(CustomPlacetypesFlag, "", "A JSON-encoded string containing custom placetypes defined using the syntax described in the whosonfirst/go-whosonfirst-placetypes repository.")

	fs.Bool(VerboseFlag, false, "Be chatty.")

	return nil
}

func ValidateCommonFlags(fs *flag.FlagSet) error {

	spatial_database_uri, err := lookup.StringVar(fs, SpatialDatabaseURIFlag)

	if err != nil {
		return fmt.Errorf("Failed to lookup %s flag, %w", SpatialDatabaseURIFlag, err)
	}

	if spatial_database_uri == "" {
		return fmt.Errorf("Invalid or missing -%s flag", SpatialDatabaseURIFlag)
	}

	_, err = lookup.StringVar(fs, PropertiesReaderURIFlag)

	if err != nil {
		return fmt.Errorf("Failed to lookup %s flag, %w", PropertiesReaderURIFlag, err)
	}

	return nil
}

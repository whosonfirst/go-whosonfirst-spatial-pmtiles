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
		return nil, err
	}

	return fs, nil
}

func AppendCommonFlags(fs *flag.FlagSet) error {

	// spatial databases

	available_databases := database.Schemes()
	desc_databases := fmt.Sprintf("A valid whosonfirst/go-whosonfirst-spatial/data.SpatialDatabase URI. options are: %s", available_databases)

	fs.String(SPATIAL_DATABASE_URI, "", desc_databases)

	available_readers := reader.Schemes()
	desc_readers := fmt.Sprintf("A valid whosonfirst/go-reader.Reader URI. Available options are: %s", available_readers)

	fs.String(PROPERTIES_READER_URI, "", fmt.Sprintf("%s. If the value is {spatial-database-uri} then the value of the '-spatial-database-uri' implements the reader.Reader interface and will be used.", desc_readers))

	fs.Bool(IS_WOF, true, "Input data is WOF-flavoured GeoJSON. (Pass a value of '0' or 'false' if you need to index non-WOF documents.")

	fs.Bool(ENABLE_CUSTOM_PLACETYPES, false, "Enable wof:placetype values that are not explicitly defined in the whosonfirst/go-whosonfirst-placetypes repository.")

	// Pending changes in the app/placetypes.go package to support
	// alternate sources (20210324/thisisaaronland)
	// fs.String(CUSTOM_PLACETYPES_SOURCE, "", "...")

	fs.String(CUSTOM_PLACETYPES, "", "A JSON-encoded string containing custom placetypes defined using the syntax described in the whosonfirst/go-whosonfirst-placetypes repository.")

	fs.Bool(VERBOSE, false, "Be chatty.")

	return nil
}

func ValidateCommonFlags(fs *flag.FlagSet) error {

	spatial_database_uri, err := lookup.StringVar(fs, SPATIAL_DATABASE_URI)

	if err != nil {
		return err
	}

	if spatial_database_uri == "" {
		return fmt.Errorf("Invalid or missing -%s flag", SPATIAL_DATABASE_URI)
	}

	_, err = lookup.StringVar(fs, PROPERTIES_READER_URI)

	if err != nil {
		return err
	}

	return nil
}

package app

import (
	"context"
	"flag"
	"github.com/sfomuseum/go-flags/lookup"
	"github.com/whosonfirst/go-whosonfirst-placetypes"
	"github.com/whosonfirst/go-whosonfirst-spatial/flags"
	"io"
	"strings"
)

func AppendCustomPlacetypesWithFlagSet(ctx context.Context, fs *flag.FlagSet) error {

	enable_custom_placetypes, _ := lookup.BoolVar(fs, flags.ENABLE_CUSTOM_PLACETYPES)

	// Alternate sources for custom placetypes are not supported yet - once they
	// are the corresponding flag in the flags/common.go package should be reenabled
	// (20210324/thisisaaronland)
	// custom_placetypes_source, _ := lookup.StringVar(fs, flags.CUSTOM_PLACETYPES_SOURCE)

	custom_placetypes_source := ""

	custom_placetypes, _ := lookup.StringVar(fs, flags.CUSTOM_PLACETYPES)

	if !enable_custom_placetypes {
		return nil
	}

	var custom_reader io.Reader

	if custom_placetypes_source == "" {
		custom_reader = strings.NewReader(custom_placetypes)
	} else {
		// whosonfirst/go-reader or ... ?
	}

	spec, err := placetypes.NewWOFPlacetypeSpecificationWithReader(custom_reader)

	if err != nil {
		return err
	}

	return placetypes.AppendPlacetypeSpecification(spec)
}

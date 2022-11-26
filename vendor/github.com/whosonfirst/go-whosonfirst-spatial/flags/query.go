package flags

import (
	"errors"
	"flag"
	"github.com/sfomuseum/go-flags/lookup"
	"github.com/sfomuseum/go-flags/multi"
	"github.com/whosonfirst/go-whosonfirst-spatial/geo"
)

func AppendQueryFlags(fs *flag.FlagSet) error {

	fs.Float64(LATITUDE, 0.0, "A valid latitude.")
	fs.Float64(LONGITUDE, 0.0, "A valid longitude.")

	fs.String(GEOMETRIES, "all", "Valid options are: all, alt, default.")

	fs.String(INCEPTION_DATE, "", "A valid EDTF date string.")
	fs.String(CESSATION_DATE, "", "A valid EDTF date string.")

	var props multi.MultiString
	fs.Var(&props, PROPERTIES, "One or more Who's On First properties to append to each result.")

	var placetypes multi.MultiString
	fs.Var(&placetypes, PLACETYPES, "One or more place types to filter results by.")

	var alt_geoms multi.MultiString
	fs.Var(&alt_geoms, ALTERNATE_GEOMETRIES, "One or more alternate geometry labels (wof:alt_label) values to filter results by.")

	var is_current multi.MultiInt64
	fs.Var(&is_current, IS_CURRENT, "One or more existential flags (-1, 0, 1) to filter results by.")

	var is_ceased multi.MultiInt64
	fs.Var(&is_ceased, IS_CEASED, "One or more existential flags (-1, 0, 1) to filter results by.")

	var is_deprecated multi.MultiInt64
	fs.Var(&is_deprecated, IS_DEPRECATED, "One or more existential flags (-1, 0, 1) to filter results by.")

	var is_superseded multi.MultiInt64
	fs.Var(&is_superseded, IS_SUPERSEDED, "One or more existential flags (-1, 0, 1) to filter results by.")

	var is_superseding multi.MultiInt64
	fs.Var(&is_superseding, IS_SUPERSEDING, "One or more existential flags (-1, 0, 1) to filter results by.")

	return nil
}

func ValidateQueryFlags(fs *flag.FlagSet) error {

	lat, err := lookup.Float64Var(fs, LATITUDE)

	if err != nil {
		return err
	}

	lon, err := lookup.Float64Var(fs, LONGITUDE)

	if err != nil {
		return err
	}

	if !geo.IsValidLatitude(lat) {
		return errors.New("Invalid latitude")
	}

	if !geo.IsValidLongitude(lon) {
		return errors.New("Invalid longitude")
	}

	_, err = lookup.StringVar(fs, GEOMETRIES)

	if err != nil {
		return err
	}

	_, err = lookup.MultiStringVar(fs, ALTERNATE_GEOMETRIES)

	if err != nil {
		return err
	}

	_, err = lookup.MultiInt64Var(fs, IS_CURRENT)

	if err != nil {
		return err
	}

	_, err = lookup.MultiInt64Var(fs, IS_CEASED)

	if err != nil {
		return err
	}

	_, err = lookup.MultiInt64Var(fs, IS_DEPRECATED)

	if err != nil {
		return err
	}

	_, err = lookup.MultiInt64Var(fs, IS_DEPRECATED)

	if err != nil {
		return err
	}

	_, err = lookup.MultiInt64Var(fs, IS_SUPERSEDED)

	if err != nil {
		return err
	}

	_, err = lookup.MultiInt64Var(fs, IS_SUPERSEDING)

	if err != nil {
		return err
	}

	return nil
}

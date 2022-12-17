package flags

import (
	"flag"
	"fmt"
	"github.com/sfomuseum/go-flags/lookup"
	"github.com/sfomuseum/go-flags/multi"
	"github.com/whosonfirst/go-whosonfirst-spatial/geo"
)

func AppendQueryFlags(fs *flag.FlagSet) error {

	fs.Float64(LatitudeFlag, 0.0, "A valid latitude.")
	fs.Float64(LongitudeFlag, 0.0, "A valid longitude.")

	fs.String(GeometriesFlag, "all", "Valid options are: all, alt, default.")

	fs.String(InceptionDateFlag, "", "A valid EDTF date string.")
	fs.String(CessationDateFlag, "", "A valid EDTF date string.")

	var props multi.MultiString
	fs.Var(&props, PropertyFlag, "One or more Who's On First properties to append to each result.")

	var placetypes multi.MultiString
	fs.Var(&placetypes, PlacetypeFlag, "One or more place types to filter results by.")

	var alt_geoms multi.MultiString
	fs.Var(&alt_geoms, AlternateGeometriesFlag, "One or more alternate geometry labels (wof:alt_label) values to filter results by.")

	var is_current multi.MultiInt64
	fs.Var(&is_current, IsCurrentFlag, "One or more existential flags (-1, 0, 1) to filter results by.")

	var is_ceased multi.MultiInt64
	fs.Var(&is_ceased, IsCeasedFlag, "One or more existential flags (-1, 0, 1) to filter results by.")

	var is_deprecated multi.MultiInt64
	fs.Var(&is_deprecated, IsDeprecatedFlag, "One or more existential flags (-1, 0, 1) to filter results by.")

	var is_superseded multi.MultiInt64
	fs.Var(&is_superseded, IsSupersededFlag, "One or more existential flags (-1, 0, 1) to filter results by.")

	var is_superseding multi.MultiInt64
	fs.Var(&is_superseding, IsSupersedingFlag, "One or more existential flags (-1, 0, 1) to filter results by.")

	var sort multi.MultiString
	fs.Var(&sort, SortURIFlag, "Zero or more whosonfirst/go-whosonfirst-spr/sort URIs.")

	return nil
}

func ValidateQueryFlags(fs *flag.FlagSet) error {

	lat, err := lookup.Float64Var(fs, LatitudeFlag)

	if err != nil {
		return fmt.Errorf("Failed to lookup %s flag, %w", LatitudeFlag, err)
	}

	lon, err := lookup.Float64Var(fs, LongitudeFlag)

	if err != nil {
		return fmt.Errorf("Failed to lookup %s flag, %w", LongitudeFlag, err)
	}

	if !geo.IsValidLatitude(lat) {
		return fmt.Errorf("Invalid latitude")
	}

	if !geo.IsValidLongitude(lon) {
		return fmt.Errorf("Invalid longitude")
	}

	_, err = lookup.StringVar(fs, GeometriesFlag)

	if err != nil {
		return fmt.Errorf("Failed to lookup %s flag, %w", GeometriesFlag, err)
	}

	_, err = lookup.MultiStringVar(fs, AlternateGeometriesFlag)

	if err != nil {
		return fmt.Errorf("Failed to lookup %s flag, %w", AlternateGeometriesFlag, err)
	}

	_, err = lookup.MultiInt64Var(fs, IsCurrentFlag)

	if err != nil {
		return fmt.Errorf("Failed to lookup %s flag, %w", IsCurrentFlag, err)
	}

	_, err = lookup.MultiInt64Var(fs, IsCeasedFlag)

	if err != nil {
		return fmt.Errorf("Failed to lookup %s flag, %w", IsCeasedFlag, err)
	}

	_, err = lookup.MultiInt64Var(fs, IsDeprecatedFlag)

	if err != nil {
		return fmt.Errorf("Failed to lookup %s flag, %w", IsDeprecatedFlag, err)
	}

	_, err = lookup.MultiInt64Var(fs, IsSupersededFlag)

	if err != nil {
		return fmt.Errorf("Failed to lookup %s flag, %w", IsSupersededFlag, err)
	}

	_, err = lookup.MultiInt64Var(fs, IsSupersedingFlag)

	if err != nil {
		return fmt.Errorf("Failed to lookup %s flag, %w", IsSupersedingFlag, err)
	}

	return nil
}

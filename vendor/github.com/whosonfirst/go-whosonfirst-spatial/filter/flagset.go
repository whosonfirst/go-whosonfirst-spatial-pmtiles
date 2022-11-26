package filter

import (
	"flag"
	"github.com/sfomuseum/go-flags/lookup"
	"github.com/whosonfirst/go-whosonfirst-spatial"
	"github.com/whosonfirst/go-whosonfirst-spatial/flags"
)

func NewSPRFilterFromFlagSet(fs *flag.FlagSet) (spatial.Filter, error) {

	inputs, err := NewSPRInputsFromFlagSet(fs)

	if err != nil {
		return nil, err
	}

	return NewSPRFilterFromInputs(inputs)
}

func NewSPRInputsFromFlagSet(fs *flag.FlagSet) (*SPRInputs, error) {

	inputs, err := NewSPRInputs()

	if err != nil {
		return nil, err
	}

	placetypes, err := lookup.MultiStringVar(fs, flags.PLACETYPES)

	if err != nil {
		return nil, err
	}

	inputs.Placetypes = placetypes

	inception_date, err := lookup.StringVar(fs, flags.INCEPTION_DATE)

	if err != nil {
		return nil, err
	}

	inputs.InceptionDate = inception_date

	cessation_date, err := lookup.StringVar(fs, flags.CESSATION_DATE)

	if err != nil {
		return nil, err
	}

	inputs.CessationDate = cessation_date

	geometries, err := lookup.StringVar(fs, flags.GEOMETRIES)

	if err != nil {
		return nil, err
	}

	inputs.Geometries = []string{geometries}

	alt_geoms, err := lookup.MultiStringVar(fs, flags.ALTERNATE_GEOMETRIES)

	if err != nil {
		return nil, err
	}

	inputs.AlternateGeometries = alt_geoms

	is_current, err := lookup.MultiInt64Var(fs, flags.IS_CURRENT)

	if err != nil {
		return nil, err
	}

	inputs.IsCurrent = is_current

	is_ceased, err := lookup.MultiInt64Var(fs, flags.IS_CEASED)

	if err != nil {
		return nil, err
	}

	inputs.IsCeased = is_ceased

	is_deprecated, err := lookup.MultiInt64Var(fs, flags.IS_DEPRECATED)

	if err != nil {
		return nil, err
	}

	inputs.IsDeprecated = is_deprecated

	is_superseded, err := lookup.MultiInt64Var(fs, flags.IS_SUPERSEDED)

	if err != nil {
		return nil, err
	}

	inputs.IsSuperseded = is_superseded

	is_superseding, err := lookup.MultiInt64Var(fs, flags.IS_SUPERSEDING)

	if err != nil {
		return nil, err
	}

	inputs.IsSuperseding = is_superseding

	return inputs, nil
}

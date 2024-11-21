package filter

import (
	"fmt"
	"log/slog"
	"runtime"

	"github.com/whosonfirst/go-whosonfirst-flags/date"
	"github.com/whosonfirst/go-whosonfirst-flags/geometry"
	"github.com/whosonfirst/go-whosonfirst-flags/placetypes"
	"github.com/whosonfirst/go-whosonfirst-spatial"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
)

func FilterSPR(filters spatial.Filter, s spr.StandardPlacesResult) error {

	var ok bool

	slog.Debug("Create placetype flag for SPR filtering", "placetype", s.Placetype())

	pf, err := placetypes.NewPlacetypeFlag(s.Placetype())

	if err != nil {
		slog.Warn("Unable to parse placetype, skipping placetype filters", "id", s.Id(), "placetype", s.Placetype(), "error", err)
	} else {

		ok = filters.HasPlacetypes(pf)

		if !ok {
			return fmt.Errorf("Failed 'placetype' test")
		}
	}

	inc_fl, err := date.NewEDTFDateFlagWithDate(s.Inception())

	if err != nil {
		return fmt.Errorf("Failed to parse inception date '%s', %v", s.Inception(), err)
	} else {

		ok := filters.MatchesInception(inc_fl)

		if !ok {
			return fmt.Errorf("Failed inception test")
		}
	}

	cessation_fl, err := date.NewEDTFDateFlagWithDate(s.Cessation())

	if err != nil {
		return fmt.Errorf("Failed to parse cessation date '%s', %v", s.Cessation(), err)
	} else {

		ok := filters.MatchesCessation(cessation_fl)

		if !ok {
			return fmt.Errorf("Failed cessation test")
		}
	}

	ok = filters.IsCurrent(s.IsCurrent())

	if !ok {
		return fmt.Errorf("Failed 'is current' test")
	}

	ok = filters.IsDeprecated(s.IsDeprecated())

	if !ok {
		return fmt.Errorf("Failed 'is deprecated' test")
	}

	ok = filters.IsCeased(s.IsCeased())

	if !ok {
		return fmt.Errorf("Failed 'is ceased' test")
	}

	ok = filters.IsSuperseded(s.IsSuperseded())

	if !ok {
		return fmt.Errorf("Failed 'is superseded' test")
	}

	ok = filters.IsSuperseding(s.IsSuperseding())

	if !ok {
		return fmt.Errorf("Failed 'is superseding' test")
	}

	switch runtime.GOOS {
	case "js":
		// This will always fail under JS (WASM)
	default:

		af, err := geometry.NewAlternateGeometryFlag(s.Path())

		if err != nil {
			slog.Warn("Unable to parse alternate geometry, skipping alt geometry filters", "id", s.Id(), "path", s.Path(), "error", err)

		} else {

			ok = filters.IsAlternateGeometry(af)

			if !ok {
				return fmt.Errorf("Failed 'is alternate geometry' test")
			}

			ok = filters.HasAlternateGeometry(af)

			if !ok {
				return fmt.Errorf("Failed 'has alternate geometry' test")
			}
		}
	}

	return nil
}

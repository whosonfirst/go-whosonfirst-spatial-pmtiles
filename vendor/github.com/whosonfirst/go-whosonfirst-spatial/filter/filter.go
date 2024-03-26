package filter

import (
	"errors"
	"fmt"
	"log"
	"runtime"

	"github.com/whosonfirst/go-whosonfirst-flags/date"
	"github.com/whosonfirst/go-whosonfirst-flags/geometry"
	"github.com/whosonfirst/go-whosonfirst-flags/placetypes"
	"github.com/whosonfirst/go-whosonfirst-spatial"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
)

func FilterSPR(filters spatial.Filter, s spr.StandardPlacesResult) error {

	var ok bool

	pf, err := placetypes.NewPlacetypeFlag(s.Placetype())

	if err != nil {
		msg := fmt.Sprintf("Unable to parse placetype (%s) for ID %s, because '%s' - skipping placetype filters", s.Placetype(), s.Id(), err)
		log.Println(msg)
	} else {

		ok = filters.HasPlacetypes(pf)

		if !ok {
			return errors.New("Failed 'placetype' test")
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
		return errors.New("Failed 'is current' test")
	}

	ok = filters.IsDeprecated(s.IsDeprecated())

	if !ok {
		return errors.New("Failed 'is deprecated' test")
	}

	ok = filters.IsCeased(s.IsCeased())

	if !ok {
		return errors.New("Failed 'is ceased' test")
	}

	ok = filters.IsSuperseded(s.IsSuperseded())

	if !ok {
		return errors.New("Failed 'is superseded' test")
	}

	ok = filters.IsSuperseding(s.IsSuperseding())

	if !ok {
		return errors.New("Failed 'is superseding' test")
	}

	switch runtime.GOOS {
	case "js":
		// This will always fail under JS (WASM)
	default:

		af, err := geometry.NewAlternateGeometryFlag(s.Path())

		if err != nil {

			msg := fmt.Sprintf("Unable to parse alternate geometry (%s) for ID %s, because '%s' - skipping alternate geometry filters", s.Path(), s.Id(), err)
			log.Println(msg)

		} else {

			ok = filters.IsAlternateGeometry(af)

			if !ok {
				return errors.New("Failed 'is alternate geometry' test")
			}

			ok = filters.HasAlternateGeometry(af)

			if !ok {
				return errors.New("Failed 'has alternate geometry' test")
			}
		}
	}

	return nil
}

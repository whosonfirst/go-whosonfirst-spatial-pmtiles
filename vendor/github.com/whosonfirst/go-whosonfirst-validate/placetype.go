package validate

import (
	"fmt"

	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"github.com/whosonfirst/go-whosonfirst-placetypes"
)

func ValidatePlacetype(body []byte) error {

	pt, err := properties.Placetype(body)

	if err != nil {
		return fmt.Errorf("Failed to derive wof:placetype from body, %w", err)
	}

	if pt == "" {
		return fmt.Errorf("Empty wof:placetype string")
	}

	if !placetypes.IsValidPlacetype(pt) {
		return fmt.Errorf("Invalid placetype")
	}

	return nil
}

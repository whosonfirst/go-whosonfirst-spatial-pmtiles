package validate

import (
	"fmt"

	"github.com/whosonfirst/go-whosonfirst-feature/properties"
)

func ValidateIsCurrent(body []byte) error {

	_, err := properties.IsCurrent(body)

	if err != nil {
		return fmt.Errorf("Failed to derive mz:is_current from body, %w", err)
	}

	return nil
}

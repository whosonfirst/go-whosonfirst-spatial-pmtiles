package validate

import (
	"fmt"

	"github.com/whosonfirst/go-whosonfirst/v4/feature"
	"github.com/whosonfirst/go-whosonfirst/v4/feature/properties"
)

func ValidateId(body []byte) error {

	_, err := properties.Id(body)

	if err != nil && !feature.IsPropertyNotFoundError(err) {
		return fmt.Errorf("Failed to derive wof:id from body, %w", err)
	}

	return nil
}

package validate

import (
	"fmt"

	"github.com/whosonfirst/go-whosonfirst/v4/feature/properties"
)

func ValidateRepo(body []byte) error {

	repo, err := properties.Repo(body)

	if err != nil {
		return fmt.Errorf("Failed to derive wof:repo from body, %w", err)
	}

	if repo == "" {
		return fmt.Errorf("Empty wof:repo string")
	}

	return nil
}

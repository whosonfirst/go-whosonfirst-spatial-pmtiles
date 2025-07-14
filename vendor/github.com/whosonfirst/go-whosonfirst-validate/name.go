package validate

import (
	"fmt"

	"github.com/whosonfirst/go-whosonfirst-feature/properties"
)

func ValidateName(body []byte) error {

	name, err := properties.Name(body)

	if err != nil {
		return fmt.Errorf("Failed to derive wof:name from body, %w", err)
	}

	if name == "" {
		return fmt.Errorf("Empty wof:name string")
	}

	// TBD: Language detection and ensuring that wof:name is English (or whatever the default language is)
	// https://github.com/pemistahl/lingua-go

	return nil
}

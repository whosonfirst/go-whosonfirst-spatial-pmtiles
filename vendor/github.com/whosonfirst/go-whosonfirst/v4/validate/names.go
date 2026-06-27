package validate

import (
	"fmt"

	"github.com/whosonfirst/go-whosonfirst/v4/feature/properties"
	"github.com/whosonfirst/go-whosonfirst/v4/names/tags"
)

func ValidateNames(body []byte) error {

	names := properties.Names(body)

	for tag, _ := range names {

		_, err := tags.NewLangTag(tag)

		if err != nil {
			return fmt.Errorf("'%s' is not a valid language tag, %w", tag, err)
		}
	}

	return nil
}

package validate

import (
	"fmt"

	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"github.com/whosonfirst/go-whosonfirst-names/tags"
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

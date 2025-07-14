package validate

import (
	"fmt"

	"github.com/sfomuseum/go-edtf"
	"github.com/sfomuseum/go-edtf/parser"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
)

func ValidateEDTF(body []byte) error {

	var err error

	inception := properties.Inception(body)
	cessation := properties.Cessation(body)
	deprecated := properties.Deprecated(body)

	err = validateEDTFString(inception)

	if err != nil {
		return fmt.Errorf("Failed to validate EDTF inception date, %w", err)
	}

	err = validateEDTFString(cessation)

	if err != nil {
		return fmt.Errorf("Failed to validate EDTF cessation date, %w", err)
	}

	if deprecated != "" {

		err = validateEDTFString(deprecated)

		if err != nil {
			return fmt.Errorf("Failed to validate EDTF deprecated date, %w", err)
		}
	}

	return nil
}

func validateEDTFString(edtf_str string) error {

	if edtf_str == edtf.UNKNOWN {
		return nil
	}

	_, err := parser.ParseString(edtf_str)
	return err
}

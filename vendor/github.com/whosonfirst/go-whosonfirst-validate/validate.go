package validate

import (
	"fmt"
	"io"

	"github.com/tidwall/geojson"
)

// To do: Replace Options{ ValidateFoo, ValidateBar } with Options{ Validators []Validator }
// where Validator is an interface whose signature is something like Validate(context.Context, []byte) error

type Options struct {
	ValidateId        bool
	ValidateName      bool
	ValidatePlacetype bool
	ValidateRepo      bool
	ValidateNames     bool
	ValidateEDTF      bool
	ValidateIsCurrent bool
}

func DefaultValidateOptions() *Options {

	return &Options{
		ValidateId:        true,
		ValidateName:      true,
		ValidatePlacetype: true,
		ValidateRepo:      true,
		ValidateNames:     true,
		ValidateEDTF:      true,
	}
}

func EnsureValidGeoJSON(r io.Reader) ([]byte, error) {

	body, err := io.ReadAll(r)

	if err != nil {
		return nil, fmt.Errorf("Failed to read body, %w", err)
	}

	// Earlier releases used paulmach/orb/geojson to do this but
	// that dependency introduces a prohibitive size requirement
	// on binaries produced by go-whosonfirst-validate-wasm. Specifically
	// in an AWS Lambda context which has a hard limit of 6MB on
	// the size of response bodies. (20230307/thisisaaronland)

	parse_opts := geojson.DefaultParseOptions

	_, err = geojson.Parse(string(body), parse_opts)

	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal body, %w", err)
	}

	return body, nil
}

func Validate(body []byte) error {

	opts := DefaultValidateOptions()
	return ValidateWithOptions(body, opts)
}

func ValidateAlt(body []byte) error {

	opts := DefaultValidateOptions()
	opts.ValidateName = false
	opts.ValidatePlacetype = false	
	opts.ValidateNames = false
	opts.ValidateEDTF = false
	opts.ValidateIsCurrent = false
	
	return ValidateWithOptions(body, opts)
}

func ValidateWithOptions(body []byte, options *Options) error {

	if options.ValidateId {

		err := ValidateId(body)

		if err != nil {
			return fmt.Errorf("Failed to validate ID, %w", err)
		}
	}

	if options.ValidateName {

		err := ValidateName(body)

		if err != nil {
			return fmt.Errorf("Failed to validate name, %w", err)
		}
	}

	if options.ValidatePlacetype {

		err := ValidatePlacetype(body)

		if err != nil {
			return fmt.Errorf("Failed to validate placetype, %w", err)
		}
	}

	if options.ValidateRepo {

		err := ValidateRepo(body)

		if err != nil {
			return fmt.Errorf("Failed to validate repo, %w", err)
		}
	}

	if options.ValidateNames {

		err := ValidateNames(body)

		if err != nil {
			return fmt.Errorf("Failed to validate name tags, %w", err)
		}
	}

	if options.ValidateEDTF {

		err := ValidateEDTF(body)

		if err != nil {
			return fmt.Errorf("Failed to validate EDTF, %w", err)
		}
	}

	if options.ValidateIsCurrent {

		err := ValidateIsCurrent(body)

		if err != nil {
			return fmt.Errorf("Failed to validate is current property, %w", err)
		}
	}

	return nil
}

package sources

import (
	"embed"
	"encoding/json"
	"errors"
	"log"
)

//go:embed *.json
var FS embed.FS

type WOFSource struct {
	Id          int64  `json:"id"`
	Fullname    string `json:"fullname"`
	Name        string `json:"name"`
	Prefix      string `json:"prefix"`
	Key         string `json:"key"`
	URL         string `json:"url"`
	License     string `json:"license"`
	Description string `json:"description"`
}

type WOFSourceSpecification map[string]WOFSource

var specification *WOFSourceSpecification

func init() {

	r, err := FS.Open("spec.json")

	if err != nil {
		log.Fatalf("Failed to open spec for reading, %v", err)
	}

	defer r.Close()

	dec := json.NewDecoder(r)
	err = dec.Decode(&specification)

	if err != nil {
		log.Fatalf("Failed to decode spec, %v", err)
	}

}

func Spec() (*WOFSourceSpecification, error) {

	return specification, nil
}

func IsValidSource(source string) bool {

	for _, details := range *specification {

		switch {
		case details.Prefix == source:
			return true
		case details.Name == source:
			return true
		default:
			// continue
		}
	}

	return false
}

func IsValidSourceId(source_id int64) bool {

	for _, details := range *specification {

		if details.Id == source_id {
			return true
		}
	}

	return false
}

func GetSourceByPrefix(source string) (*WOFSource, error) {

	for _, details := range *specification {

		if details.Prefix == source {
			return &details, nil
		}
	}

	return nil, errors.New("Invalid source")
}

func GetSourceByName(source string) (*WOFSource, error) {

	for _, details := range *specification {

		if details.Name == source {
			return &details, nil
		}
	}

	return nil, errors.New("Invalid source")
}

func GetSourceById(source_id int64) (*WOFSource, error) {

	for _, details := range *specification {

		if details.Id == source_id {
			return &details, nil
		}
	}

	return nil, errors.New("Invalid source")
}

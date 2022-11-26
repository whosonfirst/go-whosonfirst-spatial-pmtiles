package placetypes

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"sync"
)

//go:embed placetypes.json
var fs embed.FS

type WOFPlacetypeSpecification struct {
	catalog map[string]WOFPlacetype
	mu      *sync.RWMutex
}

func DefaultWOFPlacetypeSpecification() (*WOFPlacetypeSpecification, error) {

	r, err := fs.Open("placetypes.json")

	if err != nil {
		return nil, fmt.Errorf("Failed to open placetypes, %w", err)
	}

	return NewWOFPlacetypeSpecificationWithReader(r)
}

func NewWOFPlacetypeSpecificationWithReader(r io.Reader) (*WOFPlacetypeSpecification, error) {

	var catalog map[string]WOFPlacetype

	dec := json.NewDecoder(r)
	err := dec.Decode(&catalog)

	if err != nil {
		return nil, fmt.Errorf("Failed to decode reader, %w", err)
	}

	mu := new(sync.RWMutex)

	spec := &WOFPlacetypeSpecification{
		catalog: catalog,
		mu:      mu,
	}

	return spec, nil
}

func NewWOFPlacetypeSpecification(body []byte) (*WOFPlacetypeSpecification, error) {

	r := bytes.NewReader(body)
	return NewWOFPlacetypeSpecificationWithReader(r)
}

func (spec *WOFPlacetypeSpecification) GetPlacetypeByName(name string) (*WOFPlacetype, error) {

	// spec.mu.RLock()
	// defer spec.mu.RUnlock()

	for str_id, pt := range spec.catalog {

		if pt.Name == name {

			pt_id, err := strconv.Atoi(str_id)

			if err != nil {
				continue
			}

			pt_id64 := int64(pt_id)

			pt.Id = pt_id64
			return &pt, nil
		}
	}

	return nil, fmt.Errorf("Invalid placetype")
}

func (spec *WOFPlacetypeSpecification) GetPlacetypeById(id int64) (*WOFPlacetype, error) {

	// spec.mu.RLock()
	// defer spec.mu.RUnlock()

	for str_id, pt := range spec.catalog {

		pt_id, err := strconv.Atoi(str_id)

		if err != nil {
			continue
		}

		pt_id64 := int64(pt_id)

		if pt_id64 == id {
			pt.Id = pt_id64
			return &pt, nil
		}
	}

	return nil, fmt.Errorf("Invalid placetype")
}

func (spec *WOFPlacetypeSpecification) AppendPlacetypeSpecification(other_spec *WOFPlacetypeSpecification) error {

	for _, pt := range other_spec.Catalog() {

		err := spec.AppendPlacetype(pt)

		if err != nil {
			return fmt.Errorf("Failed to append placetype %v, %w", pt, err)
		}
	}

	return nil
}

func (spec *WOFPlacetypeSpecification) AppendPlacetype(pt WOFPlacetype) error {

	spec.mu.Lock()
	defer spec.mu.Unlock()

	existing_pt, _ := spec.GetPlacetypeById(pt.Id)

	if existing_pt != nil {
		return fmt.Errorf("Placetype ID already registered")
	}

	existing_pt, _ = spec.GetPlacetypeByName(pt.Name)

	if existing_pt != nil {
		return fmt.Errorf("Placetype name already registered")
	}

	for _, pid := range pt.Parent {

		_, err := spec.GetPlacetypeById(pid)

		if err != nil {
			return fmt.Errorf("Failed to get placetype by ID %d, %w", pid, err)
		}
	}

	str_id := strconv.FormatInt(pt.Id, 10)
	spec.catalog[str_id] = pt
	return nil
}

func (spec *WOFPlacetypeSpecification) Catalog() map[string]WOFPlacetype {
	return spec.catalog
}

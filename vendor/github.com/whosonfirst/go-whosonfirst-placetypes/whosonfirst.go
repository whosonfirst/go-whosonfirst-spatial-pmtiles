package placetypes

import (
	"context"
	"fmt"
)

const WHOSONFIRST_DEFINITION_SCHEME string = "whosonfirst"

type WhosOnFirstDefinition struct {
	Definition
	spec *WOFPlacetypeSpecification
	prop string
	uri string
}

func init(){
	ctx := context.Background()
	RegisterDefinition(ctx, "whosonfirst", NewWhosOnFirstDefinition)
}

func NewWhosOnFirstDefinition(ctx context.Context, uri string) (Definition, error) {

	spec, err := DefaultWOFPlacetypeSpecification()

	if err != nil {
		return nil, fmt.Errorf("Failed to create default WOF placetype specification, %w", err)
	}

	s := &WhosOnFirstDefinition{
		spec: spec,
		prop: "wof:placetype",
		uri: uri,
	}

	return s, nil
}

func (s *WhosOnFirstDefinition) Specification() *WOFPlacetypeSpecification {
	return s.spec
}

func (s *WhosOnFirstDefinition) Property() string {
	return s.prop
}

func (s *WhosOnFirstDefinition) URI() string {
	return s.uri
}
		

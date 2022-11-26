package sort

import (
	"context"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-placetypes"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
	"sort"
)

func init() {
	ctx := context.Background()
	RegisterSorter(ctx, "placetype", NewPlacetypeSorter)
}

type byPlacetype []spr.StandardPlacesResult

func (s byPlacetype) Len() int {
	return len(s)
}

func (s byPlacetype) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byPlacetype) Less(i, j int) bool {

	i_pt, err := placetypes.GetPlacetypeByName(s[i].Placetype())

	if err != nil {
		return false
	}

	j_pt, err := placetypes.GetPlacetypeByName(s[j].Placetype())

	if err != nil {
		return false
	}

	return placetypes.IsDescendant(i_pt, j_pt)
}

type PlacetypeSorter struct {
	Sorter
}

func NewPlacetypeSorter(ctx context.Context, uri string) (Sorter, error) {
	s := &PlacetypeSorter{}
	return s, nil
}

func (s *PlacetypeSorter) Sort(ctx context.Context, results spr.StandardPlacesResults, follow_on_sorters ...Sorter) (spr.StandardPlacesResults, error) {

	to_sort := results.Results()
	sort.Sort(byPlacetype(to_sort))

	switch len(follow_on_sorters) {
	case 0:

		return NewSortedStandardPlacesResults(to_sort), nil

	default:

		key_func := func(ctx context.Context, s spr.StandardPlacesResult) (string, error) {
			return s.Placetype(), nil
		}

		final, err := ApplyFollowOnSorters(ctx, to_sort, key_func, follow_on_sorters...)

		if err != nil {
			return nil, fmt.Errorf("Failed to apply follow on sorters, %w", err)
		}

		return NewSortedStandardPlacesResults(final), nil
	}
}

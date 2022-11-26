package sort

import (
	"context"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
	"sort"
)

func init() {
	ctx := context.Background()
	RegisterSorter(ctx, "inception", NewInceptionSorter)
}

type byInception []spr.StandardPlacesResult

func (s byInception) Len() int {
	return len(s)
}

func (s byInception) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s byInception) Less(i, j int) bool {

	i_inception := s[i].Inception()
	j_inception := s[j].Inception()

	if i_inception.String() == "" {
		return false
	}

	if j_inception.String() == "" {
		return true
	}

	is_before, err := i_inception.Before(j_inception)

	if err != nil {
		return false
	}

	return is_before
}

type InceptionSorter struct {
	Sorter
}

func NewInceptionSorter(ctx context.Context, uri string) (Sorter, error) {
	s := &InceptionSorter{}
	return s, nil
}

func (s *InceptionSorter) Sort(ctx context.Context, results spr.StandardPlacesResults, follow_on_sorters ...Sorter) (spr.StandardPlacesResults, error) {

	to_sort := results.Results()
	sort.Sort(byInception(to_sort))

	switch len(follow_on_sorters) {
	case 0:

		return NewSortedStandardPlacesResults(to_sort), nil

	default:

		// TBD apply a formatting or degree-of-granularity rule to s.Inception() ?

		key_func := func(ctx context.Context, s spr.StandardPlacesResult) (string, error) {
			return s.Inception().String(), nil
		}

		final, err := ApplyFollowOnSorters(ctx, to_sort, key_func, follow_on_sorters...)

		if err != nil {
			return nil, fmt.Errorf("Failed to apply follow on sorters, %w", err)
		}

		return NewSortedStandardPlacesResults(final), nil
	}
}

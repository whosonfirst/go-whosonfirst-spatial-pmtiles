package sort

import (
	"context"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
	"sort"
)

func init() {
	ctx := context.Background()
	RegisterSorter(ctx, "name", NewNameSorter)
}

type NameSorter struct {
	Sorter
}

func NewNameSorter(ctx context.Context, uri string) (Sorter, error) {
	s := &NameSorter{}
	return s, nil
}

func (s *NameSorter) Sort(ctx context.Context, results spr.StandardPlacesResults, follow_on_sorters ...Sorter) (spr.StandardPlacesResults, error) {

	lookup := make(map[string][]spr.StandardPlacesResult)

	for _, s := range results.Results() {

		_results, ok := lookup[s.Name()]

		if !ok {
			_results = make([]spr.StandardPlacesResult, 0)
		}

		_results = append(_results, s)
		lookup[s.Name()] = _results

	}

	names := make([]string, 0)

	for n, _ := range lookup {
		names = append(names, n)
	}

	sort.Strings(names)

	sorted := make([]spr.StandardPlacesResult, 0)

	for _, n := range names {

		for _, s := range lookup[n] {
			sorted = append(sorted, s)
		}
	}

	switch len(follow_on_sorters) {
	case 0:

		return NewSortedStandardPlacesResults(sorted), nil

	default:

		key_func := func(ctx context.Context, s spr.StandardPlacesResult) (string, error) {
			return s.Name(), nil
		}

		final, err := ApplyFollowOnSorters(ctx, sorted, key_func, follow_on_sorters...)

		if err != nil {
			return nil, fmt.Errorf("Failed to apply follow on sorters, %w", err)
		}

		return NewSortedStandardPlacesResults(final), nil
	}
}

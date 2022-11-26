package sort

import (
	"context"
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
)

type ApplyFollowOnSortersKeyFunc func(context.Context, spr.StandardPlacesResult) (string, error)

func ApplyFollowOnSorters(ctx context.Context, results []spr.StandardPlacesResult, key_func ApplyFollowOnSortersKeyFunc, follow_on_sorters ...Sorter) ([]spr.StandardPlacesResult, error) {

	count_follow_on := len(follow_on_sorters)

	next_sorter := follow_on_sorters[0]
	var other_sorters []Sorter

	if count_follow_on > 1 {
		other_sorters = follow_on_sorters[1:]
	}

	tmp := make(map[string][]spr.StandardPlacesResult)
	final := make([]spr.StandardPlacesResult, 0)

	last_key := ""

	doNextSort := func(key string) error {

		_results, _ := tmp[key]

		key_results := NewSortedStandardPlacesResults(_results)

		key_sorted, err := next_sorter.Sort(ctx, key_results, other_sorters...)

		if err != nil {
			return fmt.Errorf("Failed to apply next sorter to placetype '%s', %w", key, err)
		}

		for _, key_s := range key_sorted.Results() {
			final = append(final, key_s)
		}

		return nil
	}

	for _, s := range results {

		key, err := key_func(ctx, s)

		if err != nil {
			return nil, fmt.Errorf("Failed to derive key from key func, %w", err)
		}

		if key != last_key {

			if last_key != "" {

				err := doNextSort(last_key)

				if err != nil {
					return nil, fmt.Errorf("Failed to perform next sort for %s, %w", key, err)
				}
			}

			last_key = key
		}

		_results, ok := tmp[key]

		if !ok {
			_results = make([]spr.StandardPlacesResult, 0)
		}

		_results = append(_results, s)
		tmp[key] = _results
	}

	err := doNextSort(last_key)

	if err != nil {
		return nil, fmt.Errorf("Failed to perform next sort for %s, %w", last_key, err)
	}

	return final, nil
}

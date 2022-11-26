package spatial

import (
	"fmt"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
)

func SpatialIdWithFeature(body []byte, extra ...interface{}) (string, error) {

	id, err := properties.Id(body)

	if err != nil {
		return "", fmt.Errorf("Failed to derive ID for feature, %w", err)
	}

	alt_label, _ := properties.AltLabel(body)

	sp_id := fmt.Sprintf("%d#%s", id, alt_label)

	if len(extra) > 0 {

		for _, v := range extra {
			sp_id = fmt.Sprintf("%s:%v", sp_id, v)
		}
	}
	return sp_id, nil
}

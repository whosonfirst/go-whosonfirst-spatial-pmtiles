package export

import (
	"context"
	"fmt"
	"time"

	wof_properties "github.com/whosonfirst/go-whosonfirst-feature/properties"
)

// DeprecateRecord will assign the relevant properties to make 'old_body' as deprecated using the current time.
// This method does not handle assigning or updating "supersedes" or "superseded_by" properties.
func DeprecateRecord(ctx context.Context, ex Exporter, old_body []byte) ([]byte, error) {
	t := time.Now()
	return DeprecateRecordWithTime(ctx, ex, t, old_body)
}

// DeprecateRecordWithTime will assign the relevant properties to make 'old_body' as deprecated using the time defined by 't'.
// This method does not handle assigning or updating "supersedes" or "superseded_by" properties.
func DeprecateRecordWithTime(ctx context.Context, ex Exporter, t time.Time, old_body []byte) ([]byte, error) {

	to_update := map[string]interface{}{
		wof_properties.PATH_EDTF_DEPRECATED: t.Format("2006-01-02"),
		wof_properties.PATH_MZ_ISCURRENT:    0,
	}

	new_body, err := AssignProperties(ctx, old_body, to_update)

	if err != nil {
		return nil, fmt.Errorf("Failed to assign properties, %w", err)
	}

	_, new_body, err = ex.Export(ctx, new_body)

	if err != nil {
		return nil, fmt.Errorf("Failed to export body, %w", err)
	}

	return new_body, nil
}

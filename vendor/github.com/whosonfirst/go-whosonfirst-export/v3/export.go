package export

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/whosonfirst/go-whosonfirst-export/v3/properties"
	"github.com/whosonfirst/go-whosonfirst-feature/alt"
	"github.com/whosonfirst/go-whosonfirst-format"
	"github.com/whosonfirst/go-whosonfirst-validate"
)

// Export will perform all the steps necessary to "export" (as in create or update) 'feature' taking care to ensure correct formatting, default values and validation. It returns a boolean value indicating whether the feature was changed during the export process.
func Export(ctx context.Context, feature []byte) (bool, []byte, error) {

	var new_feature []byte

	tmp_feature, err := properties.RemoveTimestamps(ctx, feature)

	if err != nil {
		return false, nil, fmt.Errorf("Failed to remove timestamps from input record, %w", err)
	}

	is_alt := alt.IsAlt(feature)

	if is_alt {
		new_feature, err = PrepareAltFeatureWithoutTimestamps(ctx, feature)
	} else {
		new_feature, err = PrepareFeatureWithoutTimestamps(ctx, feature)
	}

	if err != nil {
		return false, nil, fmt.Errorf("Failed to prepare input record, %w", err)
	}

	new_feature, err = format.FormatBytes(new_feature)

	if err != nil {
		return false, nil, fmt.Errorf("Failed to format tmp record, %w", err)
	}

	// Do we really need to care about leading/trailing whitespace for this comparison?
	// I don't think we need to.

	if bytes.Equal(bytes.TrimSpace(tmp_feature), bytes.TrimSpace(new_feature)) {
		return false, feature, nil
	}

	new_feature, err = PrepareTimestamps(ctx, new_feature)

	if err != nil {
		return true, nil, fmt.Errorf("Failed to prepare record, %w", err)
	}

	if is_alt {
		err = validate.ValidateAlt(new_feature)
	} else {
		err = validate.Validate(new_feature)
	}

	if err != nil {
		return true, nil, fmt.Errorf("Failed to validate record, %w", err)
	}

	new_feature, err = format.FormatBytes(new_feature)

	if err != nil {
		return true, nil, fmt.Errorf("Failed to format record, %w", err)
	}

	return true, new_feature, nil
}

// Export will perform all the steps necessary to "export" (as in create or update) 'feature' taking care to ensure correct formatting, default values and validation writing data to 'wr' if the feature has been updated. It returns a boolean value indicating whether the feature was changed during the export process.
func WriteExportIfChanged(ctx context.Context, feature []byte, wr io.Writer) (bool, error) {

	has_changed, body, err := Export(ctx, feature)

	if err != nil {
		return has_changed, fmt.Errorf("Failed to export feature, %w", err)
	}

	if !has_changed {
		return false, nil
	}

	r := bytes.NewReader(body)
	_, err = io.Copy(wr, r)

	if err != nil {
		return true, fmt.Errorf("Failed to copy feature to writer, %w", err)
	}

	return true, nil
}

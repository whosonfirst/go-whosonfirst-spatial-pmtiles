package export

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/whosonfirst/go-whosonfirst-export/v3/properties"
)

// EnsureProperties ensure that all the properties in 'to_ensure' are present in 'body', assigning them
// if necessary.
func EnsureProperties(ctx context.Context, body []byte, to_ensure map[string]interface{}) ([]byte, error) {

	to_assign := make(map[string]interface{})

	for path, v := range to_ensure {

		rsp := gjson.GetBytes(body, path)

		if rsp.Exists() {
			continue
		}

		to_assign[path] = v
	}

	return AssignProperties(ctx, body, to_assign)
}

// EnsureProperties writes all the properties in 'to_assign' to 'body'.
func AssignProperties(ctx context.Context, body []byte, to_assign map[string]interface{}) ([]byte, error) {

	var err error

	for path, v := range to_assign {

		body, err = sjson.SetBytes(body, path, v)

		if err != nil {
			return nil, properties.SetPropertyFailed(path, err)
		}
	}

	return body, nil
}

// EnsureProperties writes all the properties in 'to_assign' to 'body' if they are absent or contain a new value.
func AssignPropertiesIfChanged(ctx context.Context, body []byte, to_assign map[string]interface{}) (bool, []byte, error) {

	var err error
	changed := false

	for path, v := range to_assign {

		rsp := gjson.GetBytes(body, path)

		if rsp.Exists() {

			old, err := json.Marshal(rsp.Value())

			if err != nil {
				return changed, nil, err
			}

			new, err := json.Marshal(v)

			if bytes.Equal(old, new) {
				continue
			}

			if err != nil {
				return changed, nil, err
			}
		}

		body, err = sjson.SetBytes(body, path, v)

		if err != nil {
			return changed, nil, properties.SetPropertyFailed(path, err)
		}

		changed = true
	}

	return changed, body, nil
}

// RemovePropertiesIfChanged filters 'to_remove' to ensure they are present in 'body' before removing them.
func RemovePropertiesIfChanged(ctx context.Context, body []byte, to_remove []string) (bool, []byte, error) {

	props := make([]string, 0)

	for _, path := range to_remove {

		rsp := gjson.GetBytes(body, path)

		if rsp.Exists() {
			props = append(props, path)
		}
	}

	if len(props) == 0 {
		return false, body, nil
	}

	new_body, err := RemoveProperties(ctx, body, props)

	if err != nil {
		return true, nil, err
	}

	return true, new_body, err
}

// RemoveProperties removes all the properties in 'to_remove' from 'body'.
func RemoveProperties(ctx context.Context, body []byte, to_remove []string) ([]byte, error) {

	var err error

	for _, path := range to_remove {

		body, err = sjson.DeleteBytes(body, path)

		if err != nil {
			return nil, properties.RemovePropertyFailed(path, err)
		}
	}

	return body, nil
}

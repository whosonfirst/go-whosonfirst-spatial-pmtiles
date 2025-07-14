package properties

import (
	"context"

	"github.com/sfomuseum/go-edtf"
	"github.com/sfomuseum/go-edtf/parser"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	wof_properties "github.com/whosonfirst/go-whosonfirst-feature/properties"
)

const dateFmt string = "2006-01-02"

func EnsureEDTF(ctx context.Context, feature []byte) ([]byte, error) {

	feature, err := EnsureInception(ctx, feature)

	if err != nil {
		return nil, err
	}

	feature, err = EnsureCessation(ctx, feature)

	if err != nil {
		return nil, err
	}

	return feature, nil
}

func EnsureInception(ctx context.Context, feature []byte) ([]byte, error) {
	return updatePath(feature, wof_properties.PATH_EDTF_INCEPTION, wof_properties.PATH_DATE_INCEPTION_UPPER, wof_properties.PATH_DATE_INCEPTION_LOWER)
}

func EnsureCessation(ctx context.Context, feature []byte) ([]byte, error) {
	return updatePath(feature, wof_properties.PATH_EDTF_CESSATION, wof_properties.PATH_DATE_CESSATION_UPPER, wof_properties.PATH_DATE_CESSATION_LOWER)
}

func updatePath(feature []byte, path string, upperPath string, lowerPath string) ([]byte, error) {

	property := gjson.GetBytes(feature, path)

	if !property.Exists() {
		return setProperties(feature, edtf.UNKNOWN, path, upperPath, lowerPath)
	}

	edtfStr := property.String()

	return setProperties(feature, edtfStr, path, upperPath, lowerPath)
}

func setProperties(feature []byte, edtfStr string, path string, upperPath, lowerPath string) ([]byte, error) {

	switch edtfStr {
	case edtf.OPEN_2012:
		edtfStr = edtf.OPEN
	case edtf.UNKNOWN_2012:
		edtfStr = edtf.UNKNOWN
	}

	feature, err := sjson.SetBytes(feature, path, edtfStr)

	if err != nil {
		return nil, SetPropertyFailed(path, err)
	}

	switch edtfStr {
	case edtf.UNKNOWN, edtf.OPEN:
		return removeUpperLower(feature, upperPath, lowerPath)
	default:
		return setUpperLower(feature, edtfStr, upperPath, lowerPath)
	}
}

func setUpperLower(feature []byte, edtfStr string, upperPath string, lowerPath string) ([]byte, error) {

	dt, err := parser.ParseString(edtfStr)

	if err != nil {
		return nil, err
	}

	lowerTime, err := dt.Lower()

	if err != nil {
		return nil, err
	}

	feature, err = sjson.SetBytes(feature, lowerPath, lowerTime.Format(dateFmt))

	if err != nil {
		return nil, SetPropertyFailed(lowerPath, err)
	}

	upperTime, err := dt.Upper()

	if err != nil {
		return nil, err
	}

	feature, err = sjson.SetBytes(feature, upperPath, upperTime.Format(dateFmt))

	if err != nil {
		return nil, SetPropertyFailed(upperPath, err)
	}

	return feature, nil
}

func removeUpperLower(feature []byte, upperPath string, lowerPath string) ([]byte, error) {

	feature, err := sjson.DeleteBytes(feature, upperPath)

	if err != nil {
		return nil, RemovePropertyFailed(upperPath, err)
	}

	feature, err = sjson.DeleteBytes(feature, lowerPath)

	if err != nil {
		return nil, RemovePropertyFailed(lowerPath, err)
	}

	return feature, nil
}

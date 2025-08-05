package properties

import (
	"context"
	"fmt"
	"sync"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	wof_properties "github.com/whosonfirst/go-whosonfirst-feature/properties"
	"github.com/whosonfirst/go-whosonfirst-id"
)

const ID_PROVIDER string = "org.whosonfirst.id.provider"

var provider_once = sync.OnceValues(func() (id.Provider, error) {
	return id.NewProvider(context.Background())
})

func idProvider(ctx context.Context) (id.Provider, error) {

	v := ctx.Value(ID_PROVIDER)

	if v != nil {

		if _, ok := v.(id.Provider); ok {
			return v.(id.Provider), nil
		} else {
			return nil, fmt.Errorf("Invalid Id provider %T", v)
		}
	}

	return provider_once()
}

func EnsureWOFIdAlt(ctx context.Context, feature []byte) ([]byte, error) {

	rsp := gjson.GetBytes(feature, wof_properties.PATH_WOF_ID)

	if !rsp.Exists() {
		return nil, wof_properties.MissingProperty(wof_properties.PATH_WOF_ID)
	}

	return feature, nil
}

func EnsureWOFId(ctx context.Context, feature []byte) ([]byte, error) {

	provider, err := idProvider(ctx)

	if err != nil {
		return nil, err
	}

	var wof_id int64

	rsp := gjson.GetBytes(feature, wof_properties.PATH_WOF_ID)

	if rsp.Exists() {

		wof_id = rsp.Int()

	} else {

		i, err := provider.NewID(ctx)

		if err != nil {
			return nil, err
		}

		wof_id = i

		feature, err = sjson.SetBytes(feature, wof_properties.PATH_WOF_ID, wof_id)

		if err != nil {
			return nil, SetPropertyFailed(wof_properties.PATH_WOF_ID, err)
		}
	}

	feature, err = sjson.SetBytes(feature, wof_properties.PATH_ID, wof_id)

	if err != nil {
		return nil, SetPropertyFailed(wof_properties.PATH_ID, err)
	}

	return feature, nil
}

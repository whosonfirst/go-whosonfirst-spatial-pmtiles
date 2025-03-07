package spatial

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	_ "log"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/whosonfirst/go-reader"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
)

type PropertiesResponse map[string]interface{}

type PropertiesResponseResults struct {
	Properties []*PropertiesResponse `json:"places"` // match spr response
}

type PropertiesResponseOptions struct {
	Reader       reader.Reader
	SourcePrefix string
	TargetPrefix string
	Keys         []string
}

func PropertiesResponseResultsWithStandardPlacesResults(ctx context.Context, opts *PropertiesResponseOptions, results spr.StandardPlacesResults) (*PropertiesResponseResults, error) {

	previous_results := results.Results()

	new_results := make([]*PropertiesResponse, len(previous_results))

	for idx, r := range previous_results {

		path := r.Path()

		fh, err := opts.Reader.Read(ctx, path)

		if err != nil {
			return nil, fmt.Errorf("Failed to open %s for reading, %w", path, err)
		}

		defer fh.Close()

		source, err := io.ReadAll(fh)

		if err != nil {
			return nil, fmt.Errorf("Failed to read body from %s, %w", path, err)
		}

		target, err := json.Marshal(r)

		if err != nil {
			return nil, fmt.Errorf("Failed to marshal %s, %w", path, err)
		}

		target, err = AppendPropertiesWithJSON(ctx, opts, source, target)

		if err != nil {
			return nil, err
		}

		var props *PropertiesResponse
		err = json.Unmarshal(target, &props)

		if err != nil {
			return nil, fmt.Errorf("Failed to unmarshal props for %s, %w", path, err)
		}

		new_results[idx] = props
	}

	props_rsp := &PropertiesResponseResults{
		Properties: new_results,
	}

	return props_rsp, nil
}

func AppendPropertiesWithJSON(ctx context.Context, opts *PropertiesResponseOptions, source []byte, target []byte) ([]byte, error) {

	var err error

	for _, e := range opts.Keys {

		paths := make([]string, 0)

		if strings.HasSuffix(e, "*") || strings.HasSuffix(e, ":") {

			e = strings.Replace(e, "*", "", -1)

			var props gjson.Result

			if opts.SourcePrefix != "" {
				props = gjson.GetBytes(source, opts.SourcePrefix)
			} else {
				props = gjson.ParseBytes(source)
			}

			for k, _ := range props.Map() {

				if strings.HasPrefix(k, e) {
					paths = append(paths, k)
				}
			}

		} else {
			paths = append(paths, e)
		}

		for _, p := range paths {

			get_path := p
			set_path := p

			if opts.SourcePrefix != "" {
				get_path = fmt.Sprintf("%s.%s", opts.SourcePrefix, get_path)
			}

			if opts.TargetPrefix != "" {
				set_path = fmt.Sprintf("%s.%s", opts.TargetPrefix, p)
			}

			v := gjson.GetBytes(source, get_path)

			/*
				log.Println("GET", get_path)
				log.Println("SET", set_path)
				log.Println("VALUE", v.Value())
			*/

			if !v.Exists() {
				continue
			}

			target, err = sjson.SetBytes(target, set_path, v.Value())

			if err != nil {
				return nil, fmt.Errorf("Failed to set %s, %w", set_path, err)
			}
		}
	}

	return target, nil
}

package provider

import (
	"fmt"
	"net/url"
	"strconv"
)

type TilezenOptions struct {
	EnableTilepack bool
	TilepackPath   string
	TilepackURL    string
}

func TilezenOptionsFromURL(u *url.URL) (*TilezenOptions, error) {

	opts := &TilezenOptions{
		EnableTilepack: false,
		TilepackPath:   "",
		TilepackURL:    "",
	}

	q := u.Query()

	q_enable_tilepack := q.Get("tilezen-enable-tilepack")

	if q_enable_tilepack != "" {

		v, err := strconv.ParseBool(q_enable_tilepack)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse ?tilezen-enable-tilepack= query parameter, %w", err)
		}

		if v == true {
			opts.EnableTilepack = true
			opts.TilepackPath = q.Get("tilezen-tilepack-path")
			opts.TilepackURL = q.Get("tilezen-tilepack-url")
		}
	}

	return opts, nil
}

package provider

import (
	"flag"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/aaronland/go-http-tangramjs"
)

// The name of the commandline flag or query parameter used to assign the `map_provider` variable.
const MapProviderFlag string = "map-provider"

// The name (label) of the map provider to use.
var map_provider string

// The name of the commandline flag or query parameter used to assign the `leaflet_enable_hash` variable.
const LeafletEnableHashFlag string = "leaflet-enable-hash"

// Enable the Leaflet.Hash plugin.
var leaflet_enable_hash bool

// The name of the commandline flag or query parameter used to assign the `leaflet_enable_fullscreen` variable.
const LeafletEnableFullscreenFlag string = "leaflet-enable-fullscreen"

// A boolean value to enable the Leaflet.Fullscreen plugin.
var leaflet_enable_fullscreen bool

// The name of the commandline flag or query parameter used to assign the `leaflet_enable_draw` variable.
const LeafletEnableDrawFlag string = "leaflet-enable-draw"

// A boolean value to enable the Leaflet.Draw plugin.
var leaflet_enable_draw bool

// The name of the commandline flag or query parameter used to assign the `leaflet_tile_url` variable.
const LeafletTileURLFlag string = "leaflet-tile-url"

// A valid Leaflet 'tileLayer' layer URL. Only necessary if `map_provider` is "leaflet".
var leaflet_tile_url string

// The name of the commandline flag or query parameter used to assign the `nextzen_apikey` variable.
const NextzenAPIKeyFlag string = "nextzen-apikey"

// A valid Nextzen API key. Only necessary if `map-provider` is "tangram".
var nextzen_apikey string

// The name of the commandline flag or query parameter used to assign the `nextzen_style_url` variable.
const NextzenStyleURLFlag string = "nextzen-style-url"

// A valid URL for loading a Tangram.js style bundle. Only necessary if `map_provider` is "tangram".
var nextzen_style_url string

// The name of the commandline flag or query parameter used to assign the `nextzen_tile_url` variable.
const NextzenTileURLFlag string = "nextzen-tile-url"

// A valid Nextzen tile URL template for loading map tiles. Only necessary if `map_provider` is "tangram".
var nextzen_tile_url string

// The name of the commandline flag or query parameter used to assign the `tilezen_enable_tilepack` variable.
const TilezenEnableTilepack string = "tilezen-enable-tilepack"

// A boolean flag to enable to use of Tilezen MBTiles tilepack for tile-serving. Only necessary if `map_provider` is "tangram".
var tilezen_enable_tilepack bool

// The name of the commandline flag or query parameter used to assign the `tilezen_tilepack_path` variable.
const TilezenTilepackPath string = "tilezen-tilepack-path"

// The path to the Tilezen MBTiles tilepack to use for serving tiles. Only necessary if `map_provider` is "tangram" and 1tilezen_enable_tilezen` is true.
var tilezen_tilepack_path string

// The name of the commandline flag or query parameter used to assign the `protomaps_tile_url` variable.
const ProtomapsTileURLFlag string = "protomaps-tile-url"

// A valid Protomaps .pmtiles URL for loading map tiles. Only necessary if `map_provider` is "protomaps".
var protomaps_tile_url string

// The name of the commandline flag or query parameter used to assign the `protomaps_serve_tiles` variable.
const ProtomapsServeTilesFlag string = "protomaps-serve-tiles"

// A boolean flag to signal whether to serve Protomaps tiles locally. Only necessary if `map_provider` is "protomaps".
var protomaps_serve_tiles bool

// The name of the commandline flag or query parameter used to assign the `protomaps_cache_size` variable.
const ProtomapsCacheSizeFlag string = "protomaps-caches-size"

// whether to serve Protomaps tiles locally. Only necessary if `map_provider` is "protomaps" and `protomaps_serve_tiles` is true.
var protomaps_cache_size int

// The name of the commandline flag or query parameter used to assign the `protomaps_bucket_uri` variable.
const ProtomapsBucketURIFlag string = "protomaps-bucket-uri"

// The `gocloud.dev/blob.Bucket` URI where Protomaps tiles are stored. Only necessary if `map_provider` is "protomaps" and `protomaps_serve_tiles` is true.
var protomaps_bucket_uri string

// The name of the commandline flag or query parameter used to assign the `protomaps_database` variable.
const ProtomapsDatabaseFlag string = "protomaps-database"

// The name of the Protomaps database to serve tiles from. Only necessary if `map_provider` is "protomaps" and `protomaps_serve_tiles` is true.
var protomaps_database string

// The names of the commandline flag or query parameter used to assign the `protomaps_paint_rules_uri` variable.
const ProtomapsPaintRulesURIFlag string = "protomaps-paint-rules-uri"

// An optional `gocloud.dev/runtimevar` URI referencing a custom Javascript variable used to define Protomaps paint rules.
var protomaps_paint_rules_uri string

// The names of the commandline flag or query parameter used to assign the `protomaps_paint_rules_uri` variable.
const ProtomapsLabelRulesURIFlag string = "protomaps-label-rules-uri"

// An optional `gocloud.dev/runtimevar` URI referencing a custom Javascript variable used to define Protomaps label rules.
var protomaps_label_rules_uri string

// The names of the commandline flag or query parameter used to assign the `javascript_at_eof` variable.
const JavaScriptAtEOFFlag string = "javascript-at-eof"

// An optional boolean flag to indicate that JavaScript resources (<script> tags) should be appended to the end of the HTML output.
var javascript_at_eof bool

// The names of the commandline flag or query parameter used to assign the `rollup_assets` variable.
const RollupAssetsFlag string = "rollup-assets"

// An optional boolean flag to indicate that multiple JavaScript and CSS assets should be minified and combined in to single files.
var rollup_assets bool

const MapPrefixFlag string = "map-prefix"

var map_prefix string

func AppendProviderFlags(fs *flag.FlagSet) error {

	schemes := Schemes()
	labels := make([]string, len(schemes))

	for idx, s := range schemes {
		labels[idx] = strings.Replace(s, "://", "", 1)
	}

	str_schemes := strings.Join(labels, ", ")
	map_provider_desc := fmt.Sprintf("The name of the map provider to use. Valid options are: %s", str_schemes)

	fs.StringVar(&map_provider, MapProviderFlag, "", map_provider_desc)

	fs.StringVar(&map_prefix, MapPrefixFlag, "", "...")

	fs.BoolVar(&javascript_at_eof, JavaScriptAtEOFFlag, false, "An optional boolean flag to indicate that JavaScript resources (<script> tags) should be appended to the end of the HTML output.")

	fs.BoolVar(&rollup_assets, RollupAssetsFlag, false, "An optional boolean flag to indicate that multiple JavaScript and CSS assets should be minified and combined in to single files.")

	err := AppendLeafletFlags(fs)

	if err != nil {
		return fmt.Errorf("Failed to append Leaflet flags, %w", err)
	}

	err = AppendTangramProviderFlags(fs)

	if err != nil {
		return fmt.Errorf("Failed to append TangramJS flags, %w", err)
	}

	err = AppendProtomapsProviderFlags(fs)

	if err != nil {
		return fmt.Errorf("Failed to append Protomaps flags, %w", err)
	}

	return nil
}

func AppendLeafletFlags(fs *flag.FlagSet) error {

	fs.BoolVar(&leaflet_enable_hash, LeafletEnableHashFlag, true, "Enable the Leaflet.Hash plugin.")
	fs.BoolVar(&leaflet_enable_fullscreen, LeafletEnableFullscreenFlag, false, "Enable the Leaflet.Fullscreen plugin.")
	fs.BoolVar(&leaflet_enable_draw, LeafletEnableDrawFlag, false, "Enable the Leaflet.Draw plugin.")

	fs.StringVar(&leaflet_tile_url, LeafletTileURLFlag, "", "A valid Leaflet 'tileLayer' layer URL. Only necessary if -map-provider is \"leaflet\".")
	return nil
}

func AppendTangramProviderFlags(fs *flag.FlagSet) error {

	fs.StringVar(&nextzen_apikey, NextzenAPIKeyFlag, "", "A valid Nextzen API key. Only necessary if -map-provider is \"tangram\".")
	fs.StringVar(&nextzen_style_url, NextzenStyleURLFlag, "/tangram/refill-style.zip", "A valid URL for loading a Tangram.js style bundle. Only necessary if -map-provider is \"tangram\".")
	fs.StringVar(&nextzen_tile_url, NextzenTileURLFlag, tangramjs.NEXTZEN_MVT_ENDPOINT, "A valid Nextzen tile URL template for loading map tiles. Only necessary if -map-provider is \"tangram\".")

	fs.BoolVar(&tilezen_enable_tilepack, TilezenEnableTilepack, false, "Enable to use of Tilezen MBTiles tilepack for tile-serving. Only necessary if -map-provider is \"tangram\".")
	fs.StringVar(&tilezen_tilepack_path, TilezenTilepackPath, "", "The path to the Tilezen MBTiles tilepack to use for serving tiles. Only necessary if -map-provider is \"tangram\" and -tilezen-enable-tilezen is true.")

	return nil
}

func AppendProtomapsProviderFlags(fs *flag.FlagSet) error {

	fs.StringVar(&protomaps_tile_url, ProtomapsTileURLFlag, "/tiles/", "A valid Protomaps .pmtiles URL for loading map tiles. Only necessary if -map-provider is \"protomaps\".")

	fs.BoolVar(&protomaps_serve_tiles, ProtomapsServeTilesFlag, false, "A boolean flag signaling whether to serve Protomaps tiles locally. Only necessary if -map-provider is \"protomaps\".")
	fs.IntVar(&protomaps_cache_size, ProtomapsCacheSizeFlag, 64, "The size of the internal Protomaps cache if serving tiles locally. Only necessary if -map-provider is \"protomaps\" and -protomaps-serve-tiles is true.")
	fs.StringVar(&protomaps_bucket_uri, ProtomapsBucketURIFlag, "", "The gocloud.dev/blob.Bucket URI where Protomaps tiles are stored. Only necessary if -map-provider is \"protomaps\" and -protomaps-serve-tiles is true.")
	fs.StringVar(&protomaps_database, ProtomapsDatabaseFlag, "", "The name of the Protomaps database to serve tiles from. Only necessary if -map-provider is \"protomaps\" and -protomaps-serve-tiles is true.")

	fs.StringVar(&protomaps_paint_rules_uri, ProtomapsPaintRulesURIFlag, "", "// An optional `gocloud.dev/runtimevar` URI referencing a custom Javascript variable used to define Protomaps paint rules.")
	fs.StringVar(&protomaps_label_rules_uri, ProtomapsLabelRulesURIFlag, "", "// An optional `gocloud.dev/runtimevar` URI referencing a custom Javascript variable used to define Protomaps label rules.")

	return nil
}

func ProviderURIFromFlagSet(fs *flag.FlagSet) (string, error) {

	u := url.URL{}
	u.Scheme = map_provider

	q := url.Values{}

	if leaflet_enable_hash {
		q.Set("leaflet-enable-hash", strconv.FormatBool(leaflet_enable_hash))
	}

	if leaflet_enable_fullscreen {
		q.Set("leaflet-enable-fullscreen", strconv.FormatBool(leaflet_enable_fullscreen))
	}

	if leaflet_enable_draw {
		q.Set("leaflet-enable-draw", strconv.FormatBool(leaflet_enable_draw))
	}

	if javascript_at_eof {
		q.Set(JavaScriptAtEOFFlag, strconv.FormatBool(javascript_at_eof))
	}

	if rollup_assets {
		q.Set(RollupAssetsFlag, strconv.FormatBool(rollup_assets))
	}

	q.Set(MapPrefixFlag, map_prefix)

	switch map_provider {
	case "leaflet":

		q.Set(LeafletTileURLFlag, leaflet_tile_url)

	case "protomaps":

		q.Set(ProtomapsTileURLFlag, protomaps_tile_url)

		if protomaps_serve_tiles {

			q.Set(ProtomapsServeTilesFlag, strconv.FormatBool(protomaps_serve_tiles))
			q.Set(ProtomapsCacheSizeFlag, strconv.Itoa(protomaps_cache_size))
			q.Set(ProtomapsBucketURIFlag, protomaps_bucket_uri)
			q.Set(ProtomapsDatabaseFlag, protomaps_database)
		}

		q.Set(ProtomapsPaintRulesURIFlag, protomaps_paint_rules_uri)
		q.Set(ProtomapsLabelRulesURIFlag, protomaps_label_rules_uri)

	case "tangram":

		q.Set("nextzen-apikey", nextzen_apikey)

		if nextzen_style_url != "" {
			q.Set("nextzen-style-url", nextzen_style_url)
		}

		if nextzen_tile_url != "" {
			q.Set("nextzen-tile-url", nextzen_tile_url)
		}

		if tilezen_enable_tilepack {

			q.Set("tilezen-enable-tilepack", strconv.FormatBool(tilezen_enable_tilepack))
			q.Set("tilezen-tilepack-path", tilezen_tilepack_path)
			q.Set("tilezen-tilepack-url", "/tilezen/")

			q.Del("nextzen-tile-url")
			q.Set("nextzen-tile-url", "/tilezen/vector/v1/512/all/{z}/{x}/{y}.mvt")
		}

	default:
		// pass
	}

	u.RawQuery = q.Encode()
	return u.String(), nil
}

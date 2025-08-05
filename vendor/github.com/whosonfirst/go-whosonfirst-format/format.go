package format

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/paulmach/orb/geojson"
	"github.com/tidwall/pretty"
)

// two space indent
const indent = "  "

// FormatFeature transforms a byte array `b` into a correctly formatted WOF file
func FormatBytes(b []byte) ([]byte, error) {

	f, err := geojson.UnmarshalFeature(b)

	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal bytes in to Feature, %w", err)
	}

	return FormatFeature(f)
}

// FormatFeature transforms a Feature into a correctly formatted WOF file
func FormatFeature(feature *geojson.Feature) ([]byte, error) {
	var buf bytes.Buffer

	_, err := buf.WriteString("{\n")
	if err != nil {
		return buf.Bytes(), err
	}

	err = writeKey(&buf, "id", feature.ID, true, false)
	if err != nil {
		return buf.Bytes(), err
	}

	err = writeKey(&buf, "type", "Feature", true, false)
	if err != nil {
		return buf.Bytes(), err
	}

	err = writeKey(&buf, "properties", feature.Properties, true, false)
	if err != nil {
		return buf.Bytes(), err
	}

	err = writeKey(&buf, "bbox", feature.BBox, true, false)
	if err != nil {
		return buf.Bytes(), err
	}

	// See this? It's important. For whatever reason orb/geojson.Feature.Geometry is of
	// type orb.Geometry (rather than orb/geojson.Geometry). Computers...

	err = writeKey(&buf, "geometry", geojson.NewGeometry(feature.Geometry), false, true)
	if err != nil {
		return buf.Bytes(), err
	}

	_, err = buf.WriteString("\n}\n")
	if err != nil {
		return buf.Bytes(), err
	}

	return buf.Bytes(), nil
}

func writeKey(buf *bytes.Buffer, key string, value interface{}, usePretty, lastLine bool) error {
	valueJSON, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if usePretty {
		prefix := indent
		prettyOpts := &pretty.Options{Indent: indent, SortKeys: true, Prefix: prefix}
		valueJSON = pretty.PrettyOptions(valueJSON, prettyOpts)
		// Trim the newline that comes back from pretty, so we can control it last
		valueJSON = valueJSON[:len(valueJSON)-1]
		// Trim the first prefix
		valueJSON = valueJSON[len(indent):]
	} else {
		valueJSON = pretty.Ugly(valueJSON)
	}

	trailing := ",\n"
	if lastLine {
		trailing = ""
	}

	_, err = buf.WriteString(indent)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(buf, "\"%s\": %s%s", key, valueJSON, trailing)
	return err
}

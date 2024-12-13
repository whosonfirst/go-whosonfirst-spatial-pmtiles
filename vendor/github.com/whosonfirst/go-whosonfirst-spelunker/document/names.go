package document

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// AppendNameStats appends statistics about the `name:*` properties in a Who's On First record.
// Specifically:
// * The unique set of language translations
// * The total number of names
// * The total number of languages
// * The total number of "preferred" names
// * The total number of "variant" names
func AppendNameStats(ctx context.Context, body []byte) ([]byte, error) {

	root := gjson.ParseBytes(body)

	props_rsp := gjson.GetBytes(body, "properties")

	if props_rsp.Exists() {
		root = props_rsp
	}

	translations_key := new(sync.Map)
	lang_key := new(sync.Map)

	count_names_total := 0
	count_names_languages := 0
	count_names_preferred := 0
	count_names_colloquial := 0
	count_names_variant := 0

	for k, v := range root.Map() {

		if !strings.HasPrefix(k, "name:") {
			continue
		}

		k = strings.Replace(k, "name:", "", 1)
		parts := strings.Split(k, "_x_")

		if len(parts) < 2 {
			continue
		}

		lang := parts[0]
		qualifier := parts[1]

		translations_key.Store(k, true)
		translations_key.Store(lang, true)

		count_names := len(v.Array())
		count_names_total += count_names

		_, ok := lang_key.Load(lang)

		if !ok {
			count_names_languages += 1
			lang_key.Store(lang, true)
		}

		switch qualifier {
		case "preferred":
			count_names_preferred += count_names
		case "variant":
			count_names_variant += count_names
		case "colloquial":
			count_names_colloquial += count_names
		default:
			// pass
		}

	}

	translations := make([]string, 0)

	translations_key.Range(func(k interface{}, v interface{}) bool {
		t := k.(string)
		translations = append(translations, t)
		return true
	})

	count_props := map[string]interface{}{
		"translations":           translations,
		"counts:names_total":     count_names_total,
		"counts:names_preferred": count_names_preferred,
		"counts:names_variant":   count_names_variant,
		"counts:names_languages": count_names_languages,
	}

	var err error

	for k, v := range count_props {

		path := k

		if props_rsp.Exists() {
			path = fmt.Sprintf("properties.%s", k)
		}

		body, err = sjson.SetBytes(body, path, v)

		if err != nil {
			return nil, err
		}
	}

	return body, nil
}

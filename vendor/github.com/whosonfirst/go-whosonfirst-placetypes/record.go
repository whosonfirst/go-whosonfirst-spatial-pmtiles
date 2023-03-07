package placetypes

// Type WOFPlacetypeRecord is a struct that maps to the JSON record files
// for individual placetypes in the `whosonfirst-placetypes` repo. Note that as of
// this writing it does not account for BCP-47 name: properties.
//
// This needs to be renamed to "Record" either at a /v1 or a v2 release. Either
// way it will be a breaking change. Doing it v2 (even if there is no explicit v1 release)
// might be "cleaner"...
type WOFPlacetypeRecord struct {
	Id           int64             `json:"wof:id"`
	Name         string            `json:"wof:name"`
	Role         string            `json:"wof:role"`
	Parent       []string          `json:"wof:parent"`
	Concordances map[string]string `json:"wof:concordances"`
}

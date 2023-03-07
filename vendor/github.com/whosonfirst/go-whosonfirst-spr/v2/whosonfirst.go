package spr

import (
	"fmt"
	"github.com/sfomuseum/go-edtf"
	"github.com/sfomuseum/go-edtf/parser"
	"github.com/whosonfirst/go-whosonfirst-feature/alt"
	"github.com/whosonfirst/go-whosonfirst-feature/geometry"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"github.com/whosonfirst/go-whosonfirst-flags"
	"github.com/whosonfirst/go-whosonfirst-flags/existential"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"strconv"
)

// It would be nice to be able to omit zero-length arrays (wof:belongsto, etc)
// but apparently it's harder than you think...
// https://github.com/golang/go/issues/45669
// https://github.com/golang/go/issues/22480

// WOFStandardPlacesResult is a struct that implements the `StandardPlacesResult` for
// Who's On First GeoJSON Feature records.
type WOFStandardPlacesResult struct {
	StandardPlacesResult `json:",omitempty"`
	EDTFInception        string  `json:"edtf:inception"`
	EDTFCessation        string  `json:"edtf:cessation"`
	WOFId                int64   `json:"wof:id"`
	WOFParentId          int64   `json:"wof:parent_id"`
	WOFName              string  `json:"wof:name"`
	WOFPlacetype         string  `json:"wof:placetype"`
	WOFCountry           string  `json:"wof:country"`
	WOFRepo              string  `json:"wof:repo"`
	WOFPath              string  `json:"wof:path"`
	WOFSupersededBy      []int64 `json:"wof:superseded_by"`
	WOFSupersedes        []int64 `json:"wof:supersedes"`
	WOFBelongsTo         []int64 `json:"wof:belongsto"`
	MZURI                string  `json:"mz:uri"`
	MZLatitude           float64 `json:"mz:latitude"`
	MZLongitude          float64 `json:"mz:longitude"`
	MZMinLatitude        float64 `json:"mz:min_latitude"`
	MZMinLongitude       float64 `json:"mz:min_longitude"`
	MZMaxLatitude        float64 `json:"mz:max_latitude"`
	MZMaxLongitude       float64 `json:"mz:max_longitude"`
	MZIsCurrent          int64   `json:"mz:is_current"`
	MZIsCeased           int64   `json:"mz:is_ceased"`
	MZIsDeprecated       int64   `json:"mz:is_deprecated"`
	MZIsSuperseded       int64   `json:"mz:is_superseded"`
	MZIsSuperseding      int64   `json:"mz:is_superseding"`
	WOFLastModified      int64   `json:"wof:lastmodified"`
}

// WhosOnFirstSPR will derive a new `WOFStandardPlacesResult` instance from 'f'.
func WhosOnFirstSPR(f []byte) (StandardPlacesResult, error) {

	if alt.IsAlt(f) {
		return nil, fmt.Errorf("Can not create SPR for alternate geometry")
	}

	id, err := properties.Id(f)

	if err != nil {
		return nil, err
	}

	parent_id, err := properties.ParentId(f)

	if err != nil {
		return nil, err
	}

	name, err := properties.Name(f)

	if err != nil {
		return nil, err
	}

	placetype, err := properties.Placetype(f)

	if err != nil {
		return nil, err
	}

	country := properties.Country(f)

	repo, err := properties.Repo(f)

	if err != nil {
		return nil, err
	}

	inception := properties.Inception(f)
	cessation := properties.Cessation(f)

	// See this: We're accounting for all the pre-2019 EDTF spec
	// inception but mostly cessation strings by silently swapping
	// them out (20210321/straup)

	_, err = parser.ParseString(inception)

	if err != nil {

		if !edtf.IsDeprecated(inception) {
			return nil, err
		}

		replacement, err := edtf.ReplaceDeprecated(inception)

		if err != nil {
			return nil, err
		}

		inception = replacement
	}

	_, err = parser.ParseString(cessation)

	if err != nil {

		if !edtf.IsDeprecated(cessation) {
			return nil, err
		}

		replacement, err := edtf.ReplaceDeprecated(cessation)

		if err != nil {
			return nil, err
		}

		cessation = replacement
	}

	path, err := uri.Id2RelPath(id)

	if err != nil {
		return nil, err
	}

	uri, err := uri.Id2AbsPath("https://data.whosonfirst.org", id)

	if err != nil {
		return nil, err
	}

	is_current, err := properties.IsCurrent(f)

	if err != nil {
		return nil, err
	}

	is_ceased, err := properties.IsCeased(f)

	if err != nil {
		return nil, err
	}

	is_deprecated, err := properties.IsDeprecated(f)

	if err != nil {
		return nil, err
	}

	is_superseded, err := properties.IsSuperseded(f)

	if err != nil {
		return nil, err
	}

	is_superseding, err := properties.IsSuperseding(f)

	if err != nil {
		return nil, err
	}

	centroid, _, err := properties.Centroid(f)

	if err != nil {
		return nil, err
	}

	geojson_geom, err := geometry.Geometry(f)

	if err != nil {
		return nil, err
	}

	orb_geom := geojson_geom.Geometry()
	mbr := orb_geom.Bound()

	superseded_by := properties.SupersededBy(f)
	supersedes := properties.Supersedes(f)

	belongsto := properties.BelongsTo(f)

	lastmod := properties.LastModified(f)

	spr := WOFStandardPlacesResult{
		WOFId:           id,
		WOFParentId:     parent_id,
		WOFPlacetype:    placetype,
		WOFName:         name,
		WOFCountry:      country,
		WOFRepo:         repo,
		WOFPath:         path,
		WOFSupersedes:   supersedes,
		WOFSupersededBy: superseded_by,
		WOFBelongsTo:    belongsto,
		EDTFInception:   inception,
		EDTFCessation:   cessation,
		MZURI:           uri,
		MZLatitude:      centroid.Y(),
		MZLongitude:     centroid.X(),
		MZMinLatitude:   mbr.Min.Y(),
		MZMinLongitude:  mbr.Min.X(),
		MZMaxLatitude:   mbr.Max.Y(),
		MZMaxLongitude:  mbr.Max.X(),
		MZIsCurrent:     is_current.Flag(),
		MZIsCeased:      is_ceased.Flag(),
		MZIsDeprecated:  is_deprecated.Flag(),
		MZIsSuperseded:  is_superseded.Flag(),
		MZIsSuperseding: is_superseding.Flag(),
		WOFLastModified: lastmod,
	}

	return &spr, nil

}

func (spr *WOFStandardPlacesResult) Id() string {
	return strconv.FormatInt(spr.WOFId, 10)
}

func (spr *WOFStandardPlacesResult) ParentId() string {
	return strconv.FormatInt(spr.WOFParentId, 10)
}

func (spr *WOFStandardPlacesResult) Name() string {
	return spr.WOFName
}

func (spr *WOFStandardPlacesResult) Inception() *edtf.EDTFDate {
	return spr.edtfDate(spr.EDTFInception)
}

func (spr *WOFStandardPlacesResult) Cessation() *edtf.EDTFDate {
	return spr.edtfDate(spr.EDTFCessation)
}

func (spr *WOFStandardPlacesResult) edtfDate(edtf_str string) *edtf.EDTFDate {

	d, err := parser.ParseString(edtf_str)

	if err != nil {
		return nil
	}

	return d
}

func (spr *WOFStandardPlacesResult) Placetype() string {
	return spr.WOFPlacetype
}

func (spr *WOFStandardPlacesResult) Country() string {
	return spr.WOFCountry
}

func (spr *WOFStandardPlacesResult) Repo() string {
	return spr.WOFRepo
}

func (spr *WOFStandardPlacesResult) Path() string {
	return spr.WOFPath
}

func (spr *WOFStandardPlacesResult) URI() string {
	return spr.MZURI
}

func (spr *WOFStandardPlacesResult) Latitude() float64 {
	return spr.MZLatitude
}

func (spr *WOFStandardPlacesResult) Longitude() float64 {
	return spr.MZLongitude
}

func (spr *WOFStandardPlacesResult) MinLatitude() float64 {
	return spr.MZMinLatitude
}

func (spr *WOFStandardPlacesResult) MinLongitude() float64 {
	return spr.MZMinLongitude
}

func (spr *WOFStandardPlacesResult) MaxLatitude() float64 {
	return spr.MZLatitude
}

func (spr *WOFStandardPlacesResult) MaxLongitude() float64 {
	return spr.MZMaxLongitude
}

func (spr *WOFStandardPlacesResult) IsCurrent() flags.ExistentialFlag {
	return existentialFlag(spr.MZIsCurrent)
}

func (spr *WOFStandardPlacesResult) IsCeased() flags.ExistentialFlag {
	return existentialFlag(spr.MZIsCeased)
}

func (spr *WOFStandardPlacesResult) IsDeprecated() flags.ExistentialFlag {
	return existentialFlag(spr.MZIsDeprecated)
}

func (spr *WOFStandardPlacesResult) IsSuperseded() flags.ExistentialFlag {
	return existentialFlag(spr.MZIsSuperseded)
}

func (spr *WOFStandardPlacesResult) IsSuperseding() flags.ExistentialFlag {
	return existentialFlag(spr.MZIsSuperseding)
}

func (spr *WOFStandardPlacesResult) SupersededBy() []int64 {
	return spr.WOFSupersededBy
}

func (spr *WOFStandardPlacesResult) Supersedes() []int64 {
	return spr.WOFSupersedes
}

func (spr *WOFStandardPlacesResult) BelongsTo() []int64 {
	return spr.WOFBelongsTo
}

func (spr *WOFStandardPlacesResult) LastModified() int64 {
	return spr.WOFLastModified
}

// we're going to assume that this won't fail since we already go through
// the process of instantiating `flags.ExistentialFlag` thingies in SPR()
// if we need to we'll just cache those instances in the `spr *WOFStandardPlacesResult`
// thingy (and omit them from the JSON output) but today that is unnecessary
// (20170816/thisisaaronland)

func existentialFlag(i int64) flags.ExistentialFlag {
	fl, _ := existential.NewKnownUnknownFlag(i)
	return fl
}

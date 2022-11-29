package spr

import (
	"fmt"
	"github.com/sfomuseum/go-edtf"
	"github.com/whosonfirst/go-whosonfirst-feature/geometry"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"github.com/whosonfirst/go-whosonfirst-flags"
	"github.com/whosonfirst/go-whosonfirst-uri"
	"strconv"
	"strings"
)

// WOFStandardPlacesResult is a struct that implements the `StandardPlacesResult` for
// Who's On First GeoJSON Feature alternate geometry records.
type WOFAltStandardPlacesResult struct {
	StandardPlacesResult `json:",omitempty"`
	WOFId                string  `json:"wof:id"`
	WOFParentId	int64 `json:wof:parent_id"`
	WOFName              string  `json:"wof:name"`
	WOFPlacetype         string  `json:"wof:placetype"`
	MZLatitude           float64 `json:"mz:latitude"`
	MZLongitude          float64 `json:"mz:longitude"`
	MZMinLatitude        float64 `json:"mz:min_latitude"`
	MZMinLongitude       float64 `json:"mz:min_longitude"`
	MZMaxLatitude        float64 `json:"mz:max_latitude"`
	MZMaxLongitude       float64 `json:"mz:max_longitude"`
	WOFPath              string  `json:"wof:path"`
	WOFRepo              string  `json:"wof:repo"`
}

// WhosOnFirstAltSPR will derive a new `WOFStandardPlacesResult` instance from 'f'.
func WhosOnFirstAltSPR(f []byte) (StandardPlacesResult, error) {

	id, err := properties.Id(f)

	if err != nil {
		return nil, fmt.Errorf("Failed to derive ID, %w", err)
	}

	source, err := properties.Source(f)

	if err != nil {
		return nil, fmt.Errorf("Failed to derive source, %w", err)
	}

	name := fmt.Sprintf("%d alt geometry (%s)", id, source)

	alt_label, err := properties.AltLabel(f)

	if err != nil {
		return nil, fmt.Errorf("Failed to derive alt label, %w", err)
	}

	label_parts := strings.Split(alt_label, "-")

	if len(label_parts) == 0 {
		return nil, fmt.Errorf("Invalid src:alt_label property")
	}

	alt_geom := &uri.AltGeom{
		Source: label_parts[0],
	}

	if len(label_parts) >= 2 {
		alt_geom.Function = label_parts[1]
	}

	if len(label_parts) >= 3 {
		alt_geom.Extras = label_parts[2:]
	}

	uri_args := &uri.URIArgs{
		IsAlternate: true,
		AltGeom:     alt_geom,
	}

	rel_path, err := uri.Id2RelPath(id, uri_args)

	if err != nil {
		return nil, fmt.Errorf("Failed to derive path for %d, %w", id, err)
	}

	repo, err := properties.Repo(f)

	if err != nil {
		return nil, fmt.Errorf("Failed to derive repo, %w", err)
	}

	geojson_geom, err := geometry.Geometry(f)

	if err != nil {
		return nil, err
	}

	orb_geom := geojson_geom.Geometry()
	mbr := orb_geom.Bound()

	lat := mbr.Min.Y() + ((mbr.Max.Y() - mbr.Min.Y()) / 2.0)
	lon := mbr.Min.X() + ((mbr.Max.X() - mbr.Min.X()) / 2.0)

	str_id := strconv.FormatInt(id, 10)

	spr := WOFAltStandardPlacesResult{
		WOFId:          str_id,
		WOFParentId: -1,
		WOFPlacetype:   "alt",
		WOFName:        name,
		MZLatitude:     lat,
		MZLongitude:    lon,
		MZMinLatitude:  mbr.Min.Y(),
		MZMinLongitude: mbr.Min.X(),
		MZMaxLatitude:  mbr.Max.Y(),
		MZMaxLongitude: mbr.Max.X(),
		WOFPath:        rel_path,
		WOFRepo:        repo,
	}

	return &spr, nil
}

func (spr *WOFAltStandardPlacesResult) Id() string {
	return spr.WOFId
}

func (spr *WOFAltStandardPlacesResult) ParentId() string {
	return "-1"
}

func (spr *WOFAltStandardPlacesResult) Name() string {
	return spr.WOFName
}

func (spr *WOFAltStandardPlacesResult) Placetype() string {
	return spr.WOFPlacetype
}

func (spr *WOFAltStandardPlacesResult) Country() string {
	return "XX"
}

func (spr *WOFAltStandardPlacesResult) Repo() string {
	return spr.WOFRepo
}

func (spr *WOFAltStandardPlacesResult) Path() string {
	return spr.WOFPath
}

func (spr *WOFAltStandardPlacesResult) URI() string {
	return ""
}

func (spr *WOFAltStandardPlacesResult) Latitude() float64 {
	return spr.MZLatitude
}

func (spr *WOFAltStandardPlacesResult) Longitude() float64 {
	return spr.MZLongitude
}

func (spr *WOFAltStandardPlacesResult) MinLatitude() float64 {
	return spr.MZMinLatitude
}

func (spr *WOFAltStandardPlacesResult) MinLongitude() float64 {
	return spr.MZMinLongitude
}

func (spr *WOFAltStandardPlacesResult) MaxLatitude() float64 {
	return spr.MZLatitude
}

func (spr *WOFAltStandardPlacesResult) MaxLongitude() float64 {
	return spr.MZMaxLongitude
}

func (spr *WOFAltStandardPlacesResult) Inception() *edtf.EDTFDate {
	return nil
}

func (spr *WOFAltStandardPlacesResult) Cessation() *edtf.EDTFDate {
	return nil
}

func (spr *WOFAltStandardPlacesResult) IsCurrent() flags.ExistentialFlag {
	return existentialFlag(-1)
}

func (spr *WOFAltStandardPlacesResult) IsCeased() flags.ExistentialFlag {
	return existentialFlag(-1)
}

func (spr *WOFAltStandardPlacesResult) IsDeprecated() flags.ExistentialFlag {
	return existentialFlag(-1)
}

func (spr *WOFAltStandardPlacesResult) IsSuperseded() flags.ExistentialFlag {
	return existentialFlag(-1)
}

func (spr *WOFAltStandardPlacesResult) IsSuperseding() flags.ExistentialFlag {
	return existentialFlag(-1)
}

func (spr *WOFAltStandardPlacesResult) SupersededBy() []int64 {
	return []int64{}
}

func (spr *WOFAltStandardPlacesResult) Supersedes() []int64 {
	return []int64{}
}

func (spr *WOFAltStandardPlacesResult) BelongsTo() []int64 {
	return []int64{}
}

func (spr *WOFAltStandardPlacesResult) LastModified() int64 {
	return -1
}

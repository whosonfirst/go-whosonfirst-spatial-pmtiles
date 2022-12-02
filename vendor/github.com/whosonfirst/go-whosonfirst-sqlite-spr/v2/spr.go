package spr

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/aaronland/go-sqlite/v2"
	"github.com/sfomuseum/go-edtf"
	"github.com/sfomuseum/go-edtf/parser"
	"github.com/whosonfirst/go-whosonfirst-flags"
	"github.com/whosonfirst/go-whosonfirst-flags/existential"
	wof_spr "github.com/whosonfirst/go-whosonfirst-spr/v2"
	"github.com/whosonfirst/go-whosonfirst-uri"
	_ "log"
	"strconv"
	"strings"
)

// SQLiteResults is a struct that implements the `whosonfirst/go-whosonfirst-spr.StandardPlacesResults`
// interface for returning a list of `StandardPlacesResults.StandardPlacesResults` instances.
type SQLiteResults struct {
	wof_spr.StandardPlacesResults `json:",omitempty"`
	// Places are the list of `StandardPlacesResults.StandardPlacesResults` instances contained by the struct.
	Places                        []wof_spr.StandardPlacesResult `json:"places"`
}

// Results returns a list of `StandardPlacesResults.StandardPlacesResults` instances.
func (r *SQLiteResults) Results() []wof_spr.StandardPlacesResult {
	return r.Places
}

// SQLiteStandardPlacesResult is a struct that implements the `whosonfirst/go-whosonfirst-spr.StandardPlacesResult`
// interface for records stored in a SQLite database and which have been indexed by the `whosonfirst/go-whosonfirst-sqlite-features`
// package.
type SQLiteStandardPlacesResult struct {
	wof_spr.StandardPlacesResult `json:",omitempty"`
	WOFId                        string  `json:"wof:id"`
	WOFParentId                  string  `json:"wof:parent_id"`
	WOFName                      string  `json:"wof:name"`
	WOFCountry                   string  `json:"wof:country"`
	WOFPlacetype                 string  `json:"wof:placetype"`
	MZLatitude                   float64 `json:"mz:latitude"`
	MZLongitude                  float64 `json:"mz:longitude"`
	MZMinLatitude                float64 `json:"mz:min_latitude"`
	MZMinLongitude               float64 `json:"mz:min_longitude"`
	MZMaxLatitude                float64 `json:"mz:max_latitude"`
	MZMaxLongitude               float64 `json:"mz:max_longitude"`
	MZIsCurrent                  int64   `json:"mz:is_current"`
	MZIsDeprecated               int64   `json:"mz:is_deprecated"`
	MZIsCeased                   int64   `json:"mz:is_ceased"`
	MZIsSuperseded               int64   `json:"mz:is_superseded"`
	MZIsSuperseding              int64   `json:"mz:is_superseding"`
	EDTFInception                string  `json:"edtf:inception"`
	EDTFCessation                string  `json:"edtf:cessation"`
	WOFSupersedes                []int64 `json:"wof:supersedes"`
	WOFSupersededBy              []int64 `json:"wof:superseded_by"`
	WOFBelongsTo                 []int64 `json:"wof:belongsto"`
	WOFPath                      string  `json:"wof:path"`
	WOFRepo                      string  `json:"wof:repo"`
	WOFLastModified              int64   `json:"wof:lastmodified"`
}

func (spr *SQLiteStandardPlacesResult) Id() string {
	return spr.WOFId
}

func (spr *SQLiteStandardPlacesResult) ParentId() string {
	return spr.WOFParentId
}

func (spr *SQLiteStandardPlacesResult) Name() string {
	return spr.WOFName
}

func (spr *SQLiteStandardPlacesResult) Placetype() string {
	return spr.WOFPlacetype
}

func (spr *SQLiteStandardPlacesResult) Country() string {
	return spr.WOFCountry
}

func (spr *SQLiteStandardPlacesResult) Repo() string {
	return spr.WOFRepo
}

func (spr *SQLiteStandardPlacesResult) Path() string {
	return spr.WOFPath
}

func (spr *SQLiteStandardPlacesResult) URI() string {
	return ""
}

func (spr *SQLiteStandardPlacesResult) Latitude() float64 {
	return spr.MZLatitude
}

func (spr *SQLiteStandardPlacesResult) Longitude() float64 {
	return spr.MZLongitude
}

func (spr *SQLiteStandardPlacesResult) MinLatitude() float64 {
	return spr.MZMinLatitude
}

func (spr *SQLiteStandardPlacesResult) MinLongitude() float64 {
	return spr.MZMinLongitude
}

func (spr *SQLiteStandardPlacesResult) MaxLatitude() float64 {
	return spr.MZMaxLatitude
}

func (spr *SQLiteStandardPlacesResult) MaxLongitude() float64 {
	return spr.MZMaxLongitude
}

func (spr *SQLiteStandardPlacesResult) IsCurrent() flags.ExistentialFlag {
	return existentialFlag(spr.MZIsCurrent)
}

func (spr *SQLiteStandardPlacesResult) IsCeased() flags.ExistentialFlag {
	return existentialFlag(spr.MZIsCeased)
}

func (spr *SQLiteStandardPlacesResult) IsDeprecated() flags.ExistentialFlag {
	return existentialFlag(spr.MZIsDeprecated)
}

func (spr *SQLiteStandardPlacesResult) IsSuperseded() flags.ExistentialFlag {
	return existentialFlag(spr.MZIsSuperseded)
}

func (spr *SQLiteStandardPlacesResult) Inception() *edtf.EDTFDate {

	d, err := parser.ParseString(spr.EDTFInception)

	if err != nil {
		return nil
	}

	return d
}

func (spr *SQLiteStandardPlacesResult) Cessation() *edtf.EDTFDate {

	d, err := parser.ParseString(spr.EDTFCessation)

	if err != nil {
		return nil
	}

	return d
}

func (spr *SQLiteStandardPlacesResult) IsSuperseding() flags.ExistentialFlag {
	return existentialFlag(spr.MZIsSuperseding)
}

func (spr *SQLiteStandardPlacesResult) SupersededBy() []int64 {
	return spr.WOFSupersededBy
}

func (spr *SQLiteStandardPlacesResult) Supersedes() []int64 {
	return spr.WOFSupersedes
}

func (spr *SQLiteStandardPlacesResult) BelongsTo() []int64 {
	return spr.WOFBelongsTo
}

func (spr *SQLiteStandardPlacesResult) LastModified() int64 {
	return spr.WOFLastModified
}

func RetrieveSPR(ctx context.Context, spr_db sqlite.Database, spr_table sqlite.Table, id int64, alt_label string) (wof_spr.StandardPlacesResult, error) {

	conn, err := spr_db.Conn(ctx)

	if err != nil {
		return nil, err
	}

	args := []interface{}{
		id,
		alt_label,
	}

	// supersedes and superseding need to be added here pending
	// https://github.com/whosonfirst/go-whosonfirst-sqlite-features/issues/14

	spr_q := fmt.Sprintf(`SELECT 
		id, parent_id, name, placetype,
		inception, cessation,
		country, repo,
		latitude, longitude,
		min_latitude, min_longitude,
		max_latitude, max_longitude,
		is_current, is_deprecated, is_ceased,is_superseded, is_superseding,
		supersedes, superseded_by, belongsto,
		is_alt, alt_label,
		lastmodified
	FROM %s WHERE id = ? AND alt_label = ?`, spr_table.Name())

	row := conn.QueryRowContext(ctx, spr_q, args...)
	return RetrieveSPRWithRow(ctx, row)
}

// See notes below

func RetrieveSPRWithRow(ctx context.Context, row *sql.Row) (wof_spr.StandardPlacesResult, error) {
	return retrieveSPRWithScanner(ctx, row)
}

// See notes below

func RetrieveSPRWithRows(ctx context.Context, rows *sql.Rows) (wof_spr.StandardPlacesResult, error) {
	return retrieveSPRWithScanner(ctx, rows)
}

// We go to the trouble of all this indirection because neither *sql.Row or *sql.Rows implement
// the sql.Scanner interface.
// The latter expects: Scan(src interface{}) error
// But the former pxpect: Scan(src ...interface{}) error

func retrieveSPRWithScanner(ctx context.Context, scanner interface{}) (wof_spr.StandardPlacesResult, error) {

	switch scanner.(type) {
	case *sql.Row, *sql.Rows:
		// pass
	default:
		return nil, fmt.Errorf("Unsupported scanner")
	}

	var spr_id string
	var parent_id string
	var name string
	var placetype string
	var country string
	var repo string

	var inception string
	var cessation string

	var latitude float64
	var longitude float64
	var min_latitude float64
	var max_latitude float64
	var min_longitude float64
	var max_longitude float64

	var is_current int64
	var is_deprecated int64
	var is_ceased int64
	var is_superseded int64
	var is_superseding int64

	// supersedes and superseding and belongsto need to be added here pending
	// https://github.com/whosonfirst/go-whosonfirst-sqlite-features/issues/14

	var str_supersedes string
	var str_superseded_by string
	var str_belongs_to string

	var is_alt int64
	var alt_label string

	var lastmodified int64

	// supersedes and superseding need to be added here pending
	// https://github.com/whosonfirst/go-whosonfirst-sqlite-features/issues/14

	var scanner_err error

	switch scanner.(type) {
	case *sql.Rows:

		scanner_err = scanner.(*sql.Rows).Scan(
			&spr_id, &parent_id, &name, &placetype,
			&inception, &cessation,
			&country, &repo,
			&latitude, &longitude,
			&min_latitude, &max_latitude, &min_longitude, &max_longitude,
			&is_current, &is_deprecated, &is_ceased, &is_superseded, &is_superseding,
			&str_supersedes, &str_superseded_by, &str_belongs_to,
			&is_alt, &alt_label,
			&lastmodified,
		)

	default:

		scanner_err = scanner.(*sql.Row).Scan(
			&spr_id, &parent_id, &name, &placetype,
			&inception, &cessation,
			&country, &repo,
			&latitude, &longitude,
			&min_latitude, &max_latitude, &min_longitude, &max_longitude,
			&is_current, &is_deprecated, &is_ceased, &is_superseded, &is_superseding,
			&str_supersedes, &str_superseded_by, &str_belongs_to,
			&is_alt, &alt_label,
			&lastmodified,
		)
	}

	if scanner_err != nil {
		return nil, scanner_err
	}

	id, err := strconv.ParseInt(spr_id, 10, 64)

	if err != nil {
		return nil, err
	}

	path, err := uri.Id2RelPath(id)

	if err != nil {
		return nil, err
	}

	_, err = parser.ParseString(inception)

	if err != nil {
		return nil, err
	}

	_, err = parser.ParseString(cessation)

	if err != nil {
		return nil, err
	}

	supersedes, err := stringToInt64(str_supersedes)

	if err != nil {
		return nil, err
	}

	superseded_by, err := stringToInt64(str_superseded_by)

	if err != nil {
		return nil, err
	}

	belongs_to, err := stringToInt64(str_belongs_to)

	if err != nil {
		return nil, err
	}

	s := &SQLiteStandardPlacesResult{
		WOFId:           spr_id,
		WOFParentId:     parent_id,
		WOFName:         name,
		WOFCountry:      country,
		WOFPlacetype:    placetype,
		MZLatitude:      latitude,
		MZLongitude:     longitude,
		MZMinLatitude:   min_latitude,
		MZMaxLatitude:   max_latitude,
		MZMinLongitude:  min_longitude,
		MZMaxLongitude:  max_longitude,
		MZIsCurrent:     is_current,
		MZIsDeprecated:  is_deprecated,
		MZIsCeased:      is_ceased,
		MZIsSuperseded:  is_superseded,
		MZIsSuperseding: is_superseding,
		WOFSupersedes:   supersedes,
		WOFSupersededBy: superseded_by,
		WOFBelongsTo:    belongs_to,
		WOFPath:         path,
		WOFRepo:         repo,
		WOFLastModified: lastmodified,
		EDTFInception:   inception,
		EDTFCessation:   cessation,
	}

	return s, nil
}

func existentialFlag(i int64) flags.ExistentialFlag {
	fl, _ := existential.NewKnownUnknownFlag(i)
	return fl
}

func stringToInt64(str string) ([]int64, error) {

	str = strings.Trim(str, " ")

	if str == "" {
		return []int64{}, nil
	}

	parts := strings.Split(str, ",")
	ints := make([]int64, len(parts))

	for idx, s := range parts {

		i, err := strconv.ParseInt(s, 10, 64)

		if err != nil {
			return nil, err
		}

		ints[idx] = i
	}

	return ints, nil
}

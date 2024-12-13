package tables

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	database_sql "github.com/sfomuseum/go-database/sql"
	"github.com/whosonfirst/go-whosonfirst-feature/alt"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
)

const SPR_TABLE_NAME string = "spr"

type SPRTableOptions struct {
	IndexAltFiles bool
}

func DefaultSPRTableOptions() (*SPRTableOptions, error) {

	opts := SPRTableOptions{
		IndexAltFiles: false,
	}

	return &opts, nil
}

type SPRTable struct {
	database_sql.Table
	FeatureTable
	name    string
	options *SPRTableOptions
}

func NewSPRTable(ctx context.Context) (database_sql.Table, error) {

	opts, err := DefaultSPRTableOptions()

	if err != nil {
		return nil, err
	}

	return NewSPRTableWithOptions(ctx, opts)
}

func NewSPRTableWithOptions(ctx context.Context, opts *SPRTableOptions) (database_sql.Table, error) {

	t := SPRTable{
		name:    SPR_TABLE_NAME,
		options: opts,
	}

	return &t, nil
}

func NewSPRTableWithDatabase(ctx context.Context, db *sql.DB) (database_sql.Table, error) {

	opts, err := DefaultSPRTableOptions()

	if err != nil {
		return nil, err
	}

	return NewSPRTableWithDatabaseAndOptions(ctx, db, opts)
}

func NewSPRTableWithDatabaseAndOptions(ctx context.Context, db *sql.DB, opts *SPRTableOptions) (database_sql.Table, error) {

	t, err := NewSPRTableWithOptions(ctx, opts)

	if err != nil {
		return nil, err
	}

	err = t.InitializeTable(ctx, db)

	if err != nil {
		return nil, err
	}

	return t, nil
}

func (t *SPRTable) InitializeTable(ctx context.Context, db *sql.DB) error {
	return database_sql.CreateTableIfNecessary(ctx, db, t)
}

func (t *SPRTable) Name() string {
	return t.name
}

func (t *SPRTable) Schema(db *sql.DB) (string, error) {
	return LoadSchema(db, SPR_TABLE_NAME)
}

func (t *SPRTable) IndexRecord(ctx context.Context, db *sql.DB, i interface{}) error {
	return t.IndexFeature(ctx, db, i.([]byte))
}

func (t *SPRTable) IndexFeature(ctx context.Context, db *sql.DB, f []byte) error {

	is_alt := alt.IsAlt(f)

	if is_alt {

		if !t.options.IndexAltFiles {
			return nil
		}
	}

	alt_label, err := properties.AltLabel(f)

	if err != nil {
		return MissingPropertyError(t, "alt label", err)
	}

	var s spr.StandardPlacesResult

	if is_alt {

		_s, err := spr.WhosOnFirstAltSPR(f)

		if err != nil {
			return database_sql.WrapError(t, fmt.Errorf("Failed to generate SPR for alt geom, %w", err))
		}

		s = _s

	} else {

		_s, err := spr.WhosOnFirstSPR(f)

		if err != nil {
			return database_sql.WrapError(t, fmt.Errorf("Failed to SPR, %w", err))
		}

		s = _s

	}

	sql := fmt.Sprintf(`INSERT OR REPLACE INTO %s (
		id, parent_id, name, placetype,
		inception, cessation,
		country, repo,
		latitude, longitude,
		min_latitude, min_longitude,
		max_latitude, max_longitude,
		is_current, is_deprecated, is_ceased,
		is_superseded, is_superseding,
		superseded_by, supersedes, belongsto,
		is_alt, alt_label,
		lastmodified
		) VALUES (
		?, ?, ?, ?,
		?, ?,
		?, ?,
		?, ?,
		?, ?,
		?, ?,
		?, ?, ?,
		?, ?, ?,
		?, ?,
		?, ?,
		?
		)`, t.Name()) // ON CONFLICT DO BLAH BLAH BLAH

	superseded_by := int64ToString(s.SupersededBy())
	supersedes := int64ToString(s.Supersedes())
	belongs_to := int64ToString(s.BelongsTo())

	str_inception := ""
	str_cessation := ""

	inception := s.Inception()
	cessation := s.Cessation()

	if inception != nil {
		str_inception = inception.String()
	}

	if cessation != nil {
		str_cessation = cessation.String()
	}

	args := []interface{}{
		s.Id(), s.ParentId(), s.Name(), s.Placetype(),
		str_inception, str_cessation,
		s.Country(), s.Repo(),
		s.Latitude(), s.Longitude(),
		s.MinLatitude(), s.MinLongitude(),
		s.MaxLatitude(), s.MaxLongitude(),
		s.IsCurrent().Flag(), s.IsDeprecated().Flag(), s.IsCeased().Flag(),
		s.IsSuperseded().Flag(), s.IsSuperseding().Flag(),
		superseded_by, supersedes, belongs_to,
		is_alt, alt_label,
		s.LastModified(),
	}

	tx, err := db.Begin()

	if err != nil {
		return database_sql.BeginTransactionError(t, err)
	}

	stmt, err := tx.Prepare(sql)

	if err != nil {
		return database_sql.PrepareStatementError(t, err)
	}

	defer stmt.Close()

	_, err = stmt.Exec(args...)

	if err != nil {
		return database_sql.ExecuteStatementError(t, err)
	}

	err = tx.Commit()

	if err != nil {
		return database_sql.CommitTransactionError(t, err)
	}

	return nil
}

func int64ToString(ints []int64) string {

	str_ints := make([]string, len(ints))

	for idx, i := range ints {
		str_ints[idx] = strconv.FormatInt(i, 10)
	}

	return strings.Join(str_ints, ",")
}

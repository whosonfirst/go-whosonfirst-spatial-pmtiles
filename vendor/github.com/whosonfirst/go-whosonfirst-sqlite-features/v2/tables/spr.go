package tables

import (
	"context"
	"fmt"
	"github.com/aaronland/go-sqlite/v2"
	"github.com/whosonfirst/go-whosonfirst-feature/alt"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"github.com/whosonfirst/go-whosonfirst-spr/v2"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features/v2"
	_ "log"
	"strconv"
	"strings"
)

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
	features.FeatureTable
	name    string
	options *SPRTableOptions
}

func NewSPRTable(ctx context.Context) (sqlite.Table, error) {

	opts, err := DefaultSPRTableOptions()

	if err != nil {
		return nil, err
	}

	return NewSPRTableWithOptions(ctx, opts)
}

func NewSPRTableWithOptions(ctx context.Context, opts *SPRTableOptions) (sqlite.Table, error) {

	t := SPRTable{
		name:    "spr",
		options: opts,
	}

	return &t, nil
}

func NewSPRTableWithDatabase(ctx context.Context, db sqlite.Database) (sqlite.Table, error) {

	opts, err := DefaultSPRTableOptions()

	if err != nil {
		return nil, err
	}

	return NewSPRTableWithDatabaseAndOptions(ctx, db, opts)
}

func NewSPRTableWithDatabaseAndOptions(ctx context.Context, db sqlite.Database, opts *SPRTableOptions) (sqlite.Table, error) {

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

func (t *SPRTable) InitializeTable(ctx context.Context, db sqlite.Database) error {

	return sqlite.CreateTableIfNecessary(ctx, db, t)
}

func (t *SPRTable) Name() string {
	return t.name
}

func (t *SPRTable) Schema() string {

	sql := `CREATE TABLE %[1]s (
			id TEXT NOT NULL,
			parent_id INTEGER,
			name TEXT,
			placetype TEXT,
			inception TEXT,
			cessation TEXT,
			country TEXT,
			repo TEXT,
			latitude REAL,
			longitude REAL,
			min_latitude REAL,
			min_longitude REAL,
			max_latitude REAL,
			max_longitude REAL,
			is_current INTEGER,
			is_deprecated INTEGER,
			is_ceased INTEGER,
			is_superseded INTEGER,
			is_superseding INTEGER,
			superseded_by TEXT,
			supersedes TEXT,
			belongsto TEXT,
			is_alt TINYINT,
			alt_label TEXT,
			lastmodified INTEGER
	);

	CREATE UNIQUE INDEX spr_by_id ON %[1]s (id, alt_label);
	CREATE INDEX spr_by_lastmod ON %[1]s (lastmodified);
	CREATE INDEX spr_by_parent ON %[1]s (parent_id, is_current, lastmodified);
	CREATE INDEX spr_by_placetype ON %[1]s (placetype, is_current, lastmodified);
	CREATE INDEX spr_by_country ON %[1]s (country, placetype, is_current, lastmodified);
	CREATE INDEX spr_by_name ON %[1]s (name, placetype, is_current, lastmodified);
	CREATE INDEX spr_by_centroid ON %[1]s (latitude, longitude, is_current, lastmodified);
	CREATE INDEX spr_by_bbox ON %[1]s (min_latitude, min_longitude, max_latitude, max_longitude, placetype, is_current, lastmodified);
	CREATE INDEX spr_by_repo ON %[1]s (repo, lastmodified);
	CREATE INDEX spr_by_current ON %[1]s (is_current, lastmodified);
	CREATE INDEX spr_by_deprecated ON %[1]s (is_deprecated, lastmodified);
	CREATE INDEX spr_by_ceased ON %[1]s (is_ceased, lastmodified);
	CREATE INDEX spr_by_superseded ON %[1]s (is_superseded, lastmodified);
	CREATE INDEX spr_by_superseding ON %[1]s (is_superseding, lastmodified);
	CREATE INDEX spr_obsolete ON %[1]s (is_deprecated, is_superseded);
	`

	return fmt.Sprintf(sql, t.Name())
}

func (t *SPRTable) IndexRecord(ctx context.Context, db sqlite.Database, i interface{}) error {
	return t.IndexFeature(ctx, db, i.([]byte))
}

func (t *SPRTable) IndexFeature(ctx context.Context, db sqlite.Database, f []byte) error {

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
			return WrapError(t, fmt.Errorf("Failed to generate SPR for alt geom, %w", err))
		}

		s = _s

	} else {

		_s, err := spr.WhosOnFirstSPR(f)

		if err != nil {
			return WrapError(t, fmt.Errorf("Failed to SPR, %w", err))
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

	conn, err := db.Conn(ctx)

	if err != nil {
		return DatabaseConnectionError(t, err)
	}

	tx, err := conn.Begin()

	if err != nil {
		return BeginTransactionError(t, err)
	}

	stmt, err := tx.Prepare(sql)

	if err != nil {
		return PrepareStatementError(t, err)
	}

	defer stmt.Close()

	_, err = stmt.Exec(args...)

	if err != nil {
		return ExecuteStatementError(t, err)
	}

	err = tx.Commit()

	if err != nil {
		return CommitTransactionError(t, err)
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

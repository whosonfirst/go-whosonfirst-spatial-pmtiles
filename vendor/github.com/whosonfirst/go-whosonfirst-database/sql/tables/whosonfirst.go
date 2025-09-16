package tables

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/paulmach/orb/encoding/wkt"
	database_sql "github.com/sfomuseum/go-database/sql"
	"github.com/tidwall/gjson"
	"github.com/whosonfirst/go-whosonfirst-feature/geometry"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	_ "github.com/whosonfirst/go-whosonfirst-uri"
)

const WHOSONFIRST_TABLE_NAME string = "whosonfirst"

type WhosonfirstTable struct {
	database_sql.Table
}

func NewWhosonfirstTableWithDatabase(ctx context.Context, db *sql.DB) (database_sql.Table, error) {

	t, err := NewWhosonfirstTable(ctx)

	if err != nil {
		return nil, fmt.Errorf("Failed to create new whosonfirst table, %w", err)
	}

	err = t.InitializeTable(ctx, db)

	if err != nil {
		return nil, fmt.Errorf("Failed to initialize whosonfirst table, %w", err)
	}

	return t, nil
}

func NewWhosonfirstTable(ctx context.Context) (database_sql.Table, error) {
	t := WhosonfirstTable{}
	return &t, nil
}

func (t *WhosonfirstTable) Name() string {
	return WHOSONFIRST_TABLE_NAME
}

// https://dev.sql.com/doc/refman/8.0/en/json-functions.html
// https://www.percona.com/blog/2016/03/07/json-document-fast-lookup-with-mysql-5-7/
// https://archive.fosdem.org/2016/schedule/event/mysql57_json/attachments/slides/1291/export/events/attachments/mysql57_json/slides/1291/MySQL_57_JSON.pdf

func (t *WhosonfirstTable) Schema(db *sql.DB) (string, error) {
	return LoadSchema(db, WHOSONFIRST_TABLE_NAME)
}

func (t *WhosonfirstTable) InitializeTable(ctx context.Context, db *sql.DB) error {
	return database_sql.CreateTableIfNecessary(ctx, db, t)
}

func (t *WhosonfirstTable) IndexRecord(ctx context.Context, db *sql.DB, tx *sql.Tx, i interface{}) error {
	return t.IndexFeature(ctx, db, tx, i.([]byte))
}

func (t *WhosonfirstTable) IndexFeature(ctx context.Context, db *sql.DB, tx *sql.Tx, body []byte) error {

	id, err := properties.Id(body)

	if err != nil {
		return fmt.Errorf("Failed to derive ID, %w", err)
	}

	geojson_geom, err := geometry.Geometry(body)

	if err != nil {
		return fmt.Errorf("Failed to derive geometry, %w", err)
	}

	orb_geom := geojson_geom.Geometry()
	wkt_geom := wkt.MarshalString(orb_geom)

	centroid, _, err := properties.Centroid(body)

	if err != nil {
		return fmt.Errorf("Failed to derive centroid, %w", err)
	}

	// See the *centroid stuff? That's important because
	// the code in paulmach/orb/encoding/wkt/wkt.go is type-checking
	// on not-a-references

	wkt_centroid := wkt.MarshalString(*centroid)

	props := gjson.GetBytes(body, "properties")
	props_json, err := json.Marshal(props.Value())

	if err != nil {
		return fmt.Errorf("Failed to encode properties, %w", err)
	}

	lastmod := properties.LastModified(body)

	q := fmt.Sprintf(`REPLACE INTO %s (
		geometry, centroid, id, properties, lastmodified
	) VALUES (
		ST_GeomFromText('%s'), ST_GeomFromText('%s'), ?, ?, ?
	)`, WHOSONFIRST_TABLE_NAME, wkt_geom, wkt_centroid)

	_, err = tx.ExecContext(ctx, q, id, string(props_json), lastmod)

	if err != nil {
		return fmt.Errorf("Failed to update table, %w", err)
	}

	return nil
}

package tables

import (
	"context"
	"database/sql"
	"fmt"

	database_sql "github.com/sfomuseum/go-database/sql"
	"github.com/whosonfirst/go-whosonfirst-feature/alt"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"github.com/whosonfirst/go-whosonfirst-spelunker/document"
)

const SPELUNKER_TABLE_NAME string = "spelunker"

type SpelunkerTableOptions struct {
	IndexAltFiles          bool
	AllowMissingSourceGeom bool
}

func DefaultSpelunkerTableOptions() (*SpelunkerTableOptions, error) {

	opts := SpelunkerTableOptions{
		IndexAltFiles:          false,
		AllowMissingSourceGeom: true,
	}

	return &opts, nil
}

type SpelunkerTable struct {
	database_sql.Table
	FeatureTable
	name    string
	options *SpelunkerTableOptions
}

func NewSpelunkerTableWithDatabase(ctx context.Context, db *sql.DB) (database_sql.Table, error) {

	opts, err := DefaultSpelunkerTableOptions()

	if err != nil {
		return nil, err
	}

	return NewSpelunkerTableWithDatabaseAndOptions(ctx, db, opts)
}

func NewSpelunkerTableWithDatabaseAndOptions(ctx context.Context, db *sql.DB, opts *SpelunkerTableOptions) (database_sql.Table, error) {

	t, err := NewSpelunkerTableWithOptions(ctx, opts)

	if err != nil {
		return nil, err
	}

	err = t.InitializeTable(ctx, db)

	if err != nil {
		return nil, err
	}

	return t, nil
}

func NewSpelunkerTable(ctx context.Context) (database_sql.Table, error) {

	opts, err := DefaultSpelunkerTableOptions()

	if err != nil {
		return nil, err
	}

	return NewSpelunkerTableWithOptions(ctx, opts)
}

func NewSpelunkerTableWithOptions(ctx context.Context, opts *SpelunkerTableOptions) (database_sql.Table, error) {

	t := SpelunkerTable{
		name:    SPELUNKER_TABLE_NAME,
		options: opts,
	}

	return &t, nil
}

func (t *SpelunkerTable) Name() string {
	return t.name
}

func (t *SpelunkerTable) Schema(db *sql.DB) (string, error) {
	return LoadSchema(db, SPELUNKER_TABLE_NAME)
}

func (t *SpelunkerTable) InitializeTable(ctx context.Context, db *sql.DB) error {
	return database_sql.CreateTableIfNecessary(ctx, db, t)
}

func (t *SpelunkerTable) IndexRecord(ctx context.Context, db *sql.DB, i interface{}) error {
	return t.IndexFeature(ctx, db, i.([]byte))
}

func (t *SpelunkerTable) IndexFeature(ctx context.Context, db *sql.DB, f []byte) error {

	is_alt := alt.IsAlt(f)

	if is_alt && !t.options.IndexAltFiles {
		return nil
	}

	id, err := properties.Id(f)

	if err != nil {
		return MissingPropertyError(t, "id", err)
	}

	source, err := properties.Source(f)

	if err != nil {

		if !t.options.AllowMissingSourceGeom {
			return MissingPropertyError(t, "source", err)
		}

		source = "unknown"
	}

	alt_label, err := properties.AltLabel(f)

	if err != nil {
		return MissingPropertyError(t, "alt label", err)
	}

	lastmod := properties.LastModified(f)

	doc, err := document.PrepareSpelunkerV2Document(ctx, f)

	if err != nil {
		return fmt.Errorf("Failed to prepare spelunker document, %w", err)
	}

	tx, err := db.Begin()

	if err != nil {
		return database_sql.BeginTransactionError(t, err)
	}

	sql := fmt.Sprintf(`INSERT OR REPLACE INTO %s (
		id, body, source, is_alt, alt_label, lastmodified
	) VALUES (
		?, ?, ?, ?, ?, ?
	)`, t.Name())

	stmt, err := tx.Prepare(sql)

	if err != nil {
		return database_sql.PrepareStatementError(t, err)
	}

	defer stmt.Close()

	str_doc := string(doc)

	_, err = stmt.Exec(id, str_doc, source, is_alt, alt_label, lastmod)

	if err != nil {
		return database_sql.ExecuteStatementError(t, err)
	}

	err = tx.Commit()

	if err != nil {
		return database_sql.CommitTransactionError(t, err)
	}

	return nil
}

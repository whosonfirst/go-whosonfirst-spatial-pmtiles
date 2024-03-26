package tables

import (
	"context"
	"fmt"
	"github.com/aaronland/go-sqlite/v2"
	"github.com/whosonfirst/go-whosonfirst-feature/alt"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"github.com/whosonfirst/go-whosonfirst-names/tags"
	sql_tables "github.com/whosonfirst/go-whosonfirst-sql/tables"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features/v2"
)

type NamesTable struct {
	features.FeatureTable
	name string
}

type NamesRow struct {
	Id           int64
	Placetype    string
	Country      string
	Language     string
	ExtLang      string
	Script       string
	Region       string
	Variant      string
	Extension    string
	PrivateUse   string
	Name         string
	LastModified int64
}

func NewNamesTableWithDatabase(ctx context.Context, db sqlite.Database) (sqlite.Table, error) {

	t, err := NewNamesTable(ctx)

	if err != nil {
		return nil, err
	}

	err = t.InitializeTable(ctx, db)

	if err != nil {
		return nil, InitializeTableError(t, err)
	}

	return t, nil
}

func NewNamesTable(ctx context.Context) (sqlite.Table, error) {

	t := NamesTable{
		name: sql_tables.NAMES_TABLE_NAME,
	}

	return &t, nil
}

func (t *NamesTable) Name() string {
	return t.name
}

func (t *NamesTable) Schema() string {
	schema, _ := sql_tables.LoadSchema("sqlite", sql_tables.NAMES_TABLE_NAME)
	return schema
}

func (t *NamesTable) InitializeTable(ctx context.Context, db sqlite.Database) error {

	return sqlite.CreateTableIfNecessary(ctx, db, t)
}

func (t *NamesTable) IndexRecord(ctx context.Context, db sqlite.Database, i interface{}) error {
	return t.IndexFeature(ctx, db, i.([]byte))
}

func (t *NamesTable) IndexFeature(ctx context.Context, db sqlite.Database, f []byte) error {

	if alt.IsAlt(f) {
		return nil
	}

	id, err := properties.Id(f)

	if err != nil {
		return MissingPropertyError(t, "id", err)
	}

	pt, err := properties.Placetype(f)

	if err != nil {
		return MissingPropertyError(t, "placetype", err)
	}

	co := properties.Country(f)

	lastmod := properties.LastModified(f)
	names := properties.Names(f)

	conn, err := db.Conn(ctx)

	if err != nil {
		return DatabaseConnectionError(t, err)
	}

	tx, err := conn.Begin()

	if err != nil {
		return BeginTransactionError(t, err)
	}

	sql := fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, t.Name())

	stmt, err := tx.Prepare(sql)

	if err != nil {
		return PrepareStatementError(t, err)
	}

	defer stmt.Close()

	_, err = stmt.Exec(id)

	if err != nil {
		return ExecuteStatementError(t, err)
	}

	for tag, names := range names {

		lt, err := tags.NewLangTag(tag)

		if err != nil {
			return WrapError(t, fmt.Errorf("Failed to create new language tag for '%s', %w", tag, err))
		}

		for _, n := range names {

			sql := fmt.Sprintf(`INSERT INTO %s (
	    			id, placetype, country,
				language, extlang,
				region, script, variant,
	    			extension, privateuse,
				name,
	    			lastmodified
			) VALUES (
	    		  	?, ?, ?,
				?, ?,
				?, ?, ?,
				?, ?,
				?,
				?
			)`, t.Name())

			stmt, err := tx.Prepare(sql)

			if err != nil {
				return PrepareStatementError(t, err)
			}

			defer stmt.Close()

			_, err = stmt.Exec(id, pt, co, lt.Language(), lt.ExtLang(), lt.Script(), lt.Region(), lt.Variant(), lt.Extension(), lt.PrivateUse(), n, lastmod)

			if err != nil {
				return ExecuteStatementError(t, err)
			}

		}
	}

	err = tx.Commit()

	if err != nil {
		return CommitTransactionError(t, err)
	}

	return nil
}

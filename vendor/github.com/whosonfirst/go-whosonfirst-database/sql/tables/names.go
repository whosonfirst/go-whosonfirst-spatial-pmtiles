package tables

import (
	"context"
	"database/sql"
	"fmt"

	database_sql "github.com/sfomuseum/go-database/sql"
	"github.com/whosonfirst/go-whosonfirst-feature/alt"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"github.com/whosonfirst/go-whosonfirst-names/tags"
)

const NAMES_TABLE_NAME string = "names"

type NamesTable struct {
	database_sql.Table
	FeatureTable
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

func NewNamesTableWithDatabase(ctx context.Context, db *sql.DB) (database_sql.Table, error) {

	t, err := NewNamesTable(ctx)

	if err != nil {
		return nil, err
	}

	err = t.InitializeTable(ctx, db)

	if err != nil {
		return nil, database_sql.InitializeTableError(t, err)
	}

	return t, nil
}

func NewNamesTable(ctx context.Context) (database_sql.Table, error) {

	t := NamesTable{
		name: NAMES_TABLE_NAME,
	}

	return &t, nil
}

func (t *NamesTable) Name() string {
	return t.name
}

func (t *NamesTable) Schema(db *sql.DB) (string, error) {
	return LoadSchema(db, NAMES_TABLE_NAME)
}

func (t *NamesTable) InitializeTable(ctx context.Context, db *sql.DB) error {

	return database_sql.CreateTableIfNecessary(ctx, db, t)
}

func (t *NamesTable) IndexRecord(ctx context.Context, db *sql.DB, i interface{}) error {
	return t.IndexFeature(ctx, db, i.([]byte))
}

func (t *NamesTable) IndexFeature(ctx context.Context, db *sql.DB, f []byte) error {

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

	db_driver := database_sql.Driver(db)

	tx, err := db.Begin()

	if err != nil {
		return database_sql.BeginTransactionError(t, err)
	}

	var delete_q string

	switch db_driver {
	case database_sql.POSTGRES_DRIVER:
		delete_q = fmt.Sprintf(`DELETE FROM %s WHERE id = $1`, t.Name())
	default:
		delete_q = fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, t.Name())
	}

	stmt, err := tx.Prepare(delete_q)

	if err != nil {
		return database_sql.PrepareStatementError(t, err)
	}

	defer stmt.Close()

	_, err = stmt.Exec(id)

	if err != nil {
		return database_sql.ExecuteStatementError(t, err)
	}

	for tag, names := range names {

		lt, err := tags.NewLangTag(tag)

		if err != nil {
			return database_sql.WrapError(t, fmt.Errorf("Failed to create new language tag for '%s', %w", tag, err))
		}

		for _, n := range names {

			var insert_q string

			switch db_driver {
			case database_sql.POSTGRES_DRIVER:

				insert_q = fmt.Sprintf(`INSERT INTO %s (
	    				id, placetype, country,
					language, extlang,
					region, script, variant,
	    				extension, privateuse,
					name,
	    				lastmodified
				) VALUES (
	    		  		$1, $2, $3,
					$4, $5,
					$6, $7, $8,
					$9, $10,
					$11,
					$12
				)`, t.Name())

			default:

				insert_q = fmt.Sprintf(`INSERT INTO %s (
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
			}

			stmt, err := tx.Prepare(insert_q)

			if err != nil {
				return database_sql.PrepareStatementError(t, err)
			}

			defer stmt.Close()

			_, err = stmt.Exec(id, pt, co, lt.Language(), lt.ExtLang(), lt.Script(), lt.Region(), lt.Variant(), lt.Extension(), lt.PrivateUse(), n, lastmod)

			if err != nil {
				return database_sql.ExecuteStatementError(t, err)
			}

		}
	}

	err = tx.Commit()

	if err != nil {
		return database_sql.CommitTransactionError(t, err)
	}

	return nil
}

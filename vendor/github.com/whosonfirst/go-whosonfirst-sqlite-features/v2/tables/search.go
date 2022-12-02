package tables

import (
	"context"
	"fmt"
	"github.com/aaronland/go-sqlite/v2"
	"github.com/whosonfirst/go-whosonfirst-feature/alt"
	"github.com/whosonfirst/go-whosonfirst-feature/properties"
	"github.com/whosonfirst/go-whosonfirst-names/tags"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features/v2"
	_ "log"
	"strings"
)

type SearchTable struct {
	features.FeatureTable
	name string
}

func NewSearchTableWithDatabase(ctx context.Context, db sqlite.Database) (sqlite.Table, error) {

	t, err := NewSearchTable(ctx)

	if err != nil {
		return nil, err
	}

	err = t.InitializeTable(ctx, db)

	if err != nil {
		return nil, err
	}

	return t, nil
}

func NewSearchTable(ctx context.Context) (sqlite.Table, error) {

	t := SearchTable{
		name: "search",
	}

	return &t, nil
}

func (t *SearchTable) InitializeTable(ctx context.Context, db sqlite.Database) error {

	return sqlite.CreateTableIfNecessary(ctx, db, t)
}

func (t *SearchTable) Name() string {
	return t.name
}

func (t *SearchTable) Schema() string {

	schema := `CREATE VIRTUAL TABLE %s USING fts4(
		id, placetype,
		name, names_all, names_preferred, names_variant, names_colloquial,		
		is_current, is_ceased, is_deprecated, is_superseded
	);`

	// so dumb...
	return fmt.Sprintf(schema, t.Name())
}

func (t *SearchTable) IndexRecord(ctx context.Context, db sqlite.Database, i interface{}) error {
	return t.IndexFeature(ctx, db, i.([]byte))
}

func (t *SearchTable) IndexFeature(ctx context.Context, db sqlite.Database, f []byte) error {

	if alt.IsAlt(f) {
		return nil
	}

	id, err := properties.Id(f)

	if err != nil {
		return MissingPropertyError(t, "id", err)
	}

	placetype, err := properties.Placetype(f)

	if err != nil {
		return MissingPropertyError(t, "placetype", err)
	}

	is_current, err := properties.IsCurrent(f)

	if err != nil {
		return MissingPropertyError(t, "is current", err)
	}

	is_ceased, err := properties.IsCeased(f)

	if err != nil {
		return MissingPropertyError(t, "is ceased", err)
	}

	is_deprecated, err := properties.IsDeprecated(f)

	if err != nil {
		return MissingPropertyError(t, "is deprecated", err)
	}

	is_superseded, err := properties.IsSuperseded(f)

	if err != nil {
		return MissingPropertyError(t, "is superseded", err)
	}

	names_all := make([]string, 0)
	names_preferred := make([]string, 0)
	names_variant := make([]string, 0)
	names_colloquial := make([]string, 0)

	name, err := properties.Name(f)

	if err != nil {
		return MissingPropertyError(t, "name", err)
	}

	names_all = append(names_all, name)
	names_preferred = append(names_preferred, name)

	for tag, names := range properties.Names(f) {

		lt, err := tags.NewLangTag(tag)

		if err != nil {
			return WrapError(t, fmt.Errorf("Failed to create new lang tag for '%s', %w", tag, err))
		}

		possible := make([]string, 0)
		possible_map := make(map[string]bool)

		for _, n := range names {

			_, ok := possible_map[n]

			if !ok {
				possible_map[n] = true
			}
		}

		for n, _ := range possible_map {
			possible = append(possible, n)
		}

		for _, n := range possible {
			names_all = append(names_all, n)
		}

		switch lt.PrivateUse() {
		case "x_preferred":
			for _, n := range possible {
				names_preferred = append(names_preferred, n)
			}
		case "x_variant":
			for _, n := range possible {
				names_variant = append(names_variant, n)
			}
		case "x_colloquial":
			for _, n := range possible {
				names_colloquial = append(names_colloquial, n)
			}
		default:
			continue
		}
	}

	sql := fmt.Sprintf(`INSERT OR REPLACE INTO %s (
		id, placetype,
		name, names_all, names_preferred, names_variant, names_colloquial,		
		is_current, is_ceased, is_deprecated, is_superseded
		) VALUES (
		?, ?,
		?, ?, ?, ?, ?,
		?, ?, ?, ?
		)`, t.Name()) // ON CONFLICT DO BLAH BLAH BLAH

	args := []interface{}{
		id, placetype,
		name, strings.Join(names_all, " "), strings.Join(names_preferred, " "), strings.Join(names_variant, " "), strings.Join(names_colloquial, " "),
		is_current.Flag(), is_ceased.Flag(), is_deprecated.Flag(), is_superseded.Flag(),
	}

	conn, err := db.Conn(ctx)

	if err != nil {
		return DatabaseConnectionError(t, err)
	}

	tx, err := conn.Begin()

	if err != nil {
		return BeginTransactionError(t, err)
	}

	s, err := tx.Prepare(fmt.Sprintf("DELETE FROM %s WHERE id = ?", t.Name()))

	if err != nil {
		return PrepareStatementError(t, err)
	}

	defer s.Close()

	_, err = s.Exec(id)

	if err != nil {
		return ExecuteStatementError(t, err)
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

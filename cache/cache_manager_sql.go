package cache

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/sfomuseum/go-database"
)

func init() {

	ctx := context.Background()
	err := RegisterCacheManager(ctx, "sql", NewSQLCacheManager)

	if err != nil {
		panic(err)
	}
}

type SQLCacheManager struct {
	feature_collection *sql.DB
	is_tmp             bool
	tmp_path           string
}

type SQLCacheManagerOptions struct {
	FeatureCollection *sql.DB
}

func NewSQLCacheManager(ctx context.Context, uri string) (CacheManager, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	engine := u.Host
	q := u.Query()

	dsn := q.Get("dsn")

	is_tmp := false
	tmp_path := ""

	// START OF wrap me in a function?

	if strings.Contains(dsn, "{tmp}") {

		f, err := os.CreateTemp("", ".db")

		if err != nil {
			return nil, fmt.Errorf("Failed to create temp file, %w", err)
		}

		tmp_path = f.Name()
		is_tmp = true

		dsn = strings.Replace(dsn, "{tmp}", tmp_path, 1)
	}

	// END OF wrap me in a function?

	conn, err := sql.Open(engine, dsn)

	if err != nil {
		return nil, fmt.Errorf("Failed to open database connection, %w", err)
	}

	features_table := &database.SQLTable{
		Name:   "features",
		Schema: "CREATE TABLE features (id TEXT PRIMARY KEY, body TEXT)",
	}

	db_opts := database.DefaultConfigureSQLDatabaseOptions()

	db_opts.CreateTablesIfNecessary = true
	db_opts.Tables = []*database.SQLTable{
		features_table,
	}

	err = database.ConfigureSQLDatabase(ctx, conn, db_opts)

	if err != nil {
		return nil, fmt.Errorf("Failed to configure database, %w", err)
	}

	switch engine {
	case "sqlite", "sqlite3":

		pragma := database.DefaultSQLitePragma()
		err := database.ConfigureSQLitePragma(ctx, conn, pragma)

		if err != nil {
			return nil, fmt.Errorf("Failed to assign pragma, %w", err)
		}
	}

	m := &SQLCacheManager{
		feature_collection: conn,
		is_tmp:             is_tmp,
		tmp_path:           tmp_path,
	}

	return m, nil
}

func (m *SQLCacheManager) CacheFeature(ctx context.Context, body []byte) (*FeatureCache, error) {

	if m.feature_collection == nil {
		return nil, fmt.Errorf("No feature collection defined")
	}

	fc, err := NewFeatureCache(body)

	if err != nil {
		return nil, fmt.Errorf("Failed to create feature cache, %w", err)
	}

	q := "INSERT OR REPLACE INTO features (id, body) VALUES (?,?)"

	_, err = m.feature_collection.ExecContext(ctx, q, fc.Id, fc.Body)

	if err != nil {
		return nil, fmt.Errorf("Failed to store feature cache for %s, %w", fc.Id, err)
	}

	return fc, nil
}

func (m *SQLCacheManager) GetFeatureCache(ctx context.Context, id string) (*FeatureCache, error) {

	if m.feature_collection == nil {
		return nil, fmt.Errorf("No feature collection defined")
	}

	var body string

	q := "SELECT body FROM features WHERE id=?"

	row := m.feature_collection.QueryRowContext(ctx, q, id)
	err := row.Scan(&body)

	switch {
	case err == sql.ErrNoRows:
		return nil, fmt.Errorf("Failed to retrieve feature, %w", err)
	case err != nil:
		return nil, fmt.Errorf("Failed to query ID, %w", err)
	default:
		//
	}

	fc := FeatureCache{
		Id:   id,
		Body: body,
	}

	return &fc, nil
}

/*
func (m *SQLCacheManager) pruneCaches(ctx context.Context, t time.Time) {
	go m.pruneFeatureCache(ctx, t)
}

func (m *SQLCacheManager) pruneFeatureCache(ctx context.Context, t time.Time) error {

	if m.feature_collection == nil {
		return nil
	}

	slog.Debug("Prune tile cache", "older than", t)

	ts := t.Unix()

	q := m.feature_collection.Query()
	q = q.Where("Created", "<=", ts)

	iter := q.Get(ctx)

	defer iter.Stop()

	for {

		var fc FeatureCache

		err := iter.Next(ctx, &fc)

		if err == io.EOF {
			break
		} else if err != nil {
			slog.Error("Failed to get next iterator", "error", err)
		} else {

			slog.Debug("Remove from feature cache", "id", fc.Id, "created", fc.Created)

			err := m.feature_collection.Delete(ctx, &fc)

			if err != nil {
				slog.Error("Failed to delete from feature cache", "id", fc.Id, "error", err)
			}
		}
	}

	return nil
}

*/

func (m *SQLCacheManager) Close() error {

	if m.feature_collection != nil {
		m.feature_collection.Close()
	}

	if m.is_tmp {
		os.Remove(m.tmp_path)
	}

	return nil
}

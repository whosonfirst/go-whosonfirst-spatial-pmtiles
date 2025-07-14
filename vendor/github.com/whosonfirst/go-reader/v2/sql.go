package reader

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/whosonfirst/go-ioutil"
	wof_uri "github.com/whosonfirst/go-whosonfirst-uri"
)

// readFunc is a function type to convert paths passed to the `Read` or `Exists` methods
// to values stored in the underlying database's "ID" column.
type readFunc func(string) (string, error)

// queryFunc is a function type to convert paths passed to the `Read` or `Exists` methods
// to query condintions used to perform record searches.
type queryFunc func(string) (string, []interface{}, error)

// VALID_TABLE is a `regexp.Regexp` for validating table names. The default is `^[a-zA-Z0-9-_]+$`.
var VALID_TABLE *regexp.Regexp

// VALID_ID is a `regexp.Regexp` for validating "ID" column names. The default is `^[a-zA-Z0-9-_]+$`.
var VALID_ID *regexp.Regexp

// VALID_BODY is a `regexp.Regexp` for validating "body" column names. The default is `^[a-zA-Z0-9-_]+$`.
var VALID_BODY *regexp.Regexp

// URI_READFUNC is a custom function to convert paths passed to the `Read` or `Exists` methods
// to values stored in the underlying database's "ID" column. The default is nil.
var URI_READFUNC readFunc

// URI_QUERYFUNC is a custom function to convert paths passed to the `Read` or `Exists` methods
// to query condintions used to perform record searches. The default is nil.
var URI_QUERYFUNC queryFunc

func init() {

	VALID_TABLE = regexp.MustCompile(`^[a-zA-Z0-9-_]+$`)
	VALID_ID = regexp.MustCompile(`^[a-zA-Z0-9-_]+$`)
	VALID_BODY = regexp.MustCompile(`^[a-zA-Z0-9-_]+$`)

	ctx := context.Background()
	err := RegisterReader(ctx, "sql", NewSQLReader)

	if err != nil {
		panic(err)
	}
}

// SQLReader is a struct that implements the `Reader` interface for reading documents from a `database/sql` compatible database engine.
type SQLReader struct {
	Reader
	conn  *sql.DB
	table string
	key   string
	value string
}

// NewSQLReader returns a new `SQLReader` instance for reading documents from from a `database/sql` compatible database engine
// configured by 'uri' in the form of:
//
//	sql://{ENGINE}/{TABLE}/{ID_COLUMN}/{BODY_COLUMN}?dsn={DSN}
//
// For example:
//
//	sql://sqlite/geojson/id/body?dsn=test.db
//
// The expectation is that `{TABLE}` will have a `{BODY_COLUMN}` column containing a Who's On First record which can be retrieved with
// a unique identifer defined in the `{ID_COLUMN}` column.
func NewSQLReader(ctx context.Context, uri string) (Reader, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	q := u.Query()

	driver := u.Host
	path := u.Path

	path = strings.TrimLeft(path, "/")
	parts := strings.Split(path, "/")

	if len(parts) != 3 {
		return nil, fmt.Errorf("Invalid path")
	}

	table := parts[0]
	key := parts[1]
	value := parts[2]
	dsn := q.Get("dsn")

	if dsn == "" {
		return nil, fmt.Errorf("Missing dsn parameter")
	}

	conn, err := sql.Open(driver, dsn)

	if err != nil {
		return nil, err
	}

	if !VALID_TABLE.MatchString(table) {
		return nil, fmt.Errorf("Invalid table")
	}

	if !VALID_ID.MatchString(key) {
		return nil, fmt.Errorf("Invalid key")
	}

	if !VALID_BODY.MatchString(value) {
		return nil, fmt.Errorf("Invalid value")
	}

	if q.Has("parse-uri") {

		v, err := strconv.ParseBool(q.Get("parse-uri"))

		if err != nil {
			return nil, fmt.Errorf("Failed to parse ?parse-uri= parameter, %w", err)
		}

		if v {

			URI_READFUNC = func(k string) (string, error) {

				id, _, err := wof_uri.ParseURI(k)

				if err != nil {
					return "", err
				}

				return strconv.FormatInt(id, 10), nil
			}
		}
	}

	r := &SQLReader{
		conn:  conn,
		table: table,
		key:   key,
		value: value,
	}

	return r, nil
}

// Read will open a `io.ReadSeekCloser` instance for the record whose "ID" column matches 'raw_uri'.
// See notes about `URI_READFUNC` and `URI_QUERYFUNC` for modifying, or deriving query criteria from, 'raw_uri'
// before database queries are performed.
func (r *SQLReader) Read(ctx context.Context, raw_uri string) (io.ReadSeekCloser, error) {

	q, q_args, err := r.deriveQuery(ctx, raw_uri, r.key)

	if err != nil {
		return nil, err
	}

	row := r.conn.QueryRowContext(ctx, q, q_args...)

	var body string
	err = row.Scan(&body)

	if err != nil {
		return nil, err
	}

	sr := strings.NewReader(body)
	fh, err := ioutil.NewReadSeekCloser(sr)

	if err != nil {
		return nil, err
	}

	return fh, nil
}

// Exists returns a boolean value indicating whether 'path' already exists (meaning it will always return false).
// Read will open a `io.ReadSeekCloser` instance for the record whose "ID" column matches 'raw_uri'.
// See notes about `URI_READFUNC` and `URI_QUERYFUNC` for modifying, or deriving query criteria from, 'raw_uri'
// before database queries are performed.
func (r *SQLReader) Exists(ctx context.Context, raw_uri string) (bool, error) {

	q, q_args, err := r.deriveQuery(ctx, raw_uri, "1")

	if err != nil {
		return false, err
	}

	row := r.conn.QueryRowContext(ctx, q, q_args...)

	var one int
	err = row.Scan(&one)

	if err != nil {

		if err == sql.ErrNoRows {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// ReaderURI will return the value of 'raw_uri' optionally modified by `URI_READFUNC` if defined..
func (r *SQLReader) ReaderURI(ctx context.Context, raw_uri string) string {

	uri := raw_uri

	if URI_READFUNC != nil {

		new_uri, err := URI_READFUNC(raw_uri)

		if err != nil {
			return ""
		}

		uri = new_uri
	}

	return uri
}

func (r SQLReader) deriveQuery(ctx context.Context, raw_uri string, col string) (string, []any, error) {

	uri := raw_uri

	if URI_READFUNC != nil {

		new_uri, err := URI_READFUNC(raw_uri)

		if err != nil {
			return "", nil, err
		}

		uri = new_uri
	}

	q := fmt.Sprintf("SELECT %s FROM %s WHERE %s=?", r.value, r.table, col)

	q_args := []interface{}{
		uri,
	}

	if URI_QUERYFUNC != nil {

		extra_where, extra_args, err := URI_QUERYFUNC(raw_uri)

		if err != nil {
			return "", nil, err
		}

		if extra_where != "" {

			q = fmt.Sprintf("%s AND %s", q, extra_where)

			for _, a := range extra_args {
				q_args = append(q_args, a)
			}
		}
	}

	return q, q_args, nil
}

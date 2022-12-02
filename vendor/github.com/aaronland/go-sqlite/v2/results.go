package sqlite

// this is here so we can pass both sql.Row and sql.Rows to the
// ResultSetFunc below (20170824/thisisaaronland)

type ResultSet interface {
	Scan(dest ...interface{}) error
}

type ResultRow interface {
	Row() interface{}
}

type ResultSetFunc func(row ResultSet) (ResultRow, error)

# go-whosonfirst-sqlite-spr

Go package to implement the `whosonfirst/go-whosonfirst-spr` interface for "standard places result" (SPR) data stored in a SQLite database that has been indexed using the `whosonfirst/go-whosonfirst-sqlite-features` package.

## Documentation

[![Go Reference](https://pkg.go.dev/badge/github.com/whosonfirst/go-whosonfirst-sqlite-spr.svg)](https://pkg.go.dev/github.com/whosonfirst/go-whosonfirst-sqlite-spr)

## Description

`go-whosonfirst-sqlite-spr` is a Go package to implement the `whosonfirst/go-whosonfirst-spr` interface for ["standard places result"](https://github.com/whosonfirst/go-whosonfirst-spr) (SPR) data stored in a SQLite database, specifically data stored in [an `spr` table](https://github.com/whosonfirst/go-whosonfirst-sqlite-features#spr) as indexed by the `go-whosonfirst-sqlite-features` package.

This package exposes a single public method called `RetrieveSPR` that retrieves a row from a `spr` table in a SQLite database and returns it as an instance that implements the `go-whosonfirst-spr.SPR` interface. 

The method signature is:

```
func RetrieveSPR(context.Context, database.SQLiteDatabase, sqlite.Table, int64, string) (spr.StandardPlacesResult, error)
```

For example:

```
import (
        "context"
	"github.com/aaronland/go-sqlite/database"
	"github.com/whosonfirst/go-whosonfirst-sqlite-features/tables"
	"github.com/whosonfirst/go-whosonfirst-sqlite-spr"
)

ctx := context.Background()

db, _ := wof_database.NewDB("example.db")
spr_table, _ := tables.NewSPRTableWithDatabase(db)

id := 1234
alt_label := ""

spr_r, _ := spr.RetrieveSPR(ctx, db, spr_table, id, alt_label)
```

_Error handling omitted for the sake of brevity._

The `spr_r` instance returned will have the type `SQLiteStandardPlacesResult` and implements all of the `spr.StandardPlacesResult` methods. Under the hood it looks like this:

```
type SQLiteStandardPlacesResult struct {
	spr.StandardPlacesResult     `json:",omitempty"`
	WOFId                        string  `json:"wof:id"`
	WOFParentId                  string  `json:"wof:parent_id"`
	WOFName                      string  `json:"wof:name"`
	WOFCountry                   string  `json:"wof:country"`
	WOFPlacetype                 string  `json:"wof:placetype"`
	MZLatitude                   float64 `json:"mz:latitude"`
	MZLongitude                  float64 `json:"mz:longitude"`
	MZMinLatitude                float64 `json:"mz:min_latitude"`
	MZMinLongitude               float64 `json:"mz:min_longitude"`
	MZMaxLatitude                float64 `json:"mz:max_latitude"`
	MZMaxLongitude               float64 `json:"mz:max_longitude"`
	MZIsCurrent                  int64   `json:"mz:is_current"`
	MZIsDeprecated               int64   `json:"mz:is_deprecated"`
	MZIsCeased                   int64   `json:"mz:is_ceased"`
	MZIsSuperseded               int64   `json:"mz:is_superseded"`
	MZIsSuperseding              int64   `json:"mz:is_superseding"`
	WOFPath         	     string  `json:"wof:path"`
	WOFRepo         	     string  `json:"wof:repo"`
	WOFLastModified 	     int64   `json:"wof:lastmodified"`
}
```

## See also

* https://github.com/aaronland/go-sqlite
* https://github.com/whosonfirst/go-whosonfirst-spr
* https://github.com/whosonfirst/go-whosonfirst-sqlite-features
# go-whosonfirst-export

Go package for exporting Who's On First documents.

## Documentation

[![Go Reference](https://pkg.go.dev/badge/github.com/whosonfirst/go-whosonfirst-export.svg)](https://pkg.go.dev/github.com/whosonfirst/go-whosonfirst-export)

## Motivation

This package is designed to perform all the steps necessary to "export" (as in create or update) a Who's On First record taking care to ensure correct formatting, default values and validation.

## Example

Version 3.x of this package introduce major, backward-incompatible changes from earlier releases. That said, migragting from version 2.x to 3.x should be relatively straightforward as the _basic_ concepts are still the same but (hopefully) simplified. There are some important changes "under the hood" but the user-facing changes, while important, should be easy to update.

_All error handling removed for the sake of brevity._

### Simple

```
import (
	"context"
	"os

	"github.com/whosonfirst/go-whosonfirst-export/v3"
)

func main() {

	ctx := context.Background()
	body, _ := os.ReadFile(path)

	has_changed, new_body, _ := export.Export(ctx, body)

	if has_changes {
		os.Stdout.Write(new_body)
	}
}
```

This is how you would have done the same thing using the `/v2` release:

```
import (
	"context"
	"os

	"github.com/whosonfirst/go-whosonfirst-export/v2"
)

func main() {

	ctx := context.Background()

	body, _ := os.ReadFile(path)
	opts, _ := export.NewDefaultOptions(ctx)
	
	export.Export(body, opts, os.Stdout)
}
```

### Exporter

The `Exporter` interface provides a common interface to allow for customized export functionality in your code which can supplement the default export functionality with application-specific needs. The interface consists of a single method whose signature matches the standard `Export` method:

```
type Exporter interface {
	Export(context.Context, []byte) (bool, []byte, error)
}
```

Custom `Exporter` implementations are "registered" at runtime with the `export.RegisterExporter` method. For example:

```
type CustomExporter struct {
	export.Exporter
}

func init() {
	ctx := context.Background()
	export.RegisterExporter(ctx, "custom", NewCustomExporter)
}

func NewWhosOnFirstExporter(ctx context.Context, uri string) (export.Exporter, error) {
	// Your code here...
}
```

_For a complete implementation consult [exporter_whosonfirst.go](exporter_whosonfirst.go)._

#### whosonfirst://

This package provides a default `whosonfirst://` exporter implementation (which is really just a thin wrapper around the `Export` method) which can be used like this:

```
import (
	"context"
	"os

	"github.com/whosonfirst/go-whosonfirst-export/v3"
)

func main() {

	ctx := context.Background()
	ex, _ := export.NewExporter(ctx, "whosonfirst://")
	
	path := "some.geojson"     	
	body, _ := os.ReadFile(path)

	has_changes, body, _ = ex.Export(ctx, body)

	if has_changes {
		os.Stdout.Write(body)
	}
}
```

This is how you would have done the same thing using the `/v2` release:

```
import (
	"context"
	"os

	"github.com/whosonfirst/go-whosonfirst-export/v2"
)

func main() {

	ctx := context.Background()
	ex, _ := export.NewExporter(ctx, "whosonfirst://")
	
	path := "some.geojson"     	
	body, _ := os.ReadFile(path)

	body, _ = ex.Export(ctx, body)
	os.Stdout.Write(body)
}
```

## See also

* https://github.com/whosonfirst/go-whosonfirst-format
* https://github.com/whosonfirst/go-whosonfirst-validate
* https://github.com/whosonfirst/go-whosonfirst-feature
* https://github.com/whosonfirst/go-whosonfirst-id
* https://github.com/tidwall/gjson
* https://github.com/tidwall/sjson

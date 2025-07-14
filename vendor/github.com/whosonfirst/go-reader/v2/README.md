# go-reader

There are many interfaces for reading files. This one is ours. It returns `io.ReadSeekCloser` instances.

## Documentation

[![Go Reference](https://pkg.go.dev/badge/github.com/whosonfirst/go-reader.svg)](https://pkg.go.dev/github.com/whosonfirst/go-reader/v2)

### Example

Readers are instantiated with the `reader.NewReader` method which takes as its arguments a `context.Context` instance and a URI string. The URI's scheme represents the type of reader it implements and the remaining (URI) properties are used by that reader type to instantiate itself.

For example to read files from a directory on the local filesystem you would write:

```
package main

import (
	"context"
	"io"
	"os"

	"github.com/whosonfirst/go-reader/v2"
)

func main() {
	ctx := context.Background()
	r, _ := reader.NewReader(ctx, "file:///usr/local/data")
	fh, _ := r.Read(ctx, "example.txt")
	defer fh.Close()
	io.Copy(os.Stdout, fh)
}
```

There is also a handy "null" reader in case you need a "pretend" reader that doesn't actually do anything:

```
package main

import (
	"context"
	"io"
	"os"

	"github.com/whosonfirst/go-reader/v2"
)

func main() {
	ctx := context.Background()
	r, _ := reader.NewReader(ctx, "null://")
	fh, _ := r.Read(ctx, "example.txt")
	defer fh.Close()
	io.Copy(os.Stdout, fh)
}
```

## Interfaces

### reader.Reader

```
// Reader is an interface for reading data from multiple sources or targets.
type Reader interface {
	// Reader returns a `io.ReadSeekCloser` instance for a URI resolved by the instance implementing the `Reader` interface.
	Read(context.Context, string) (io.ReadSeekCloser, error)
	// Exists returns a boolean value indicating whether a URI already exists.
	Exists(context.Context, string) (bool, error)
	// The absolute path for the file is determined by the instance implementing the `Reader` interface.
	ReaderURI(context.Context, string) string
}
```

## Custom readers

Custom readers need to:

1. Implement the interface above.
2. Announce their availability using the `go-reader.RegisterReader` method on initialization, passing in an initialization function implementing the `go-reader.ReaderInitializationFunc` interface.

For example, this is how the [http://](http.go) reader is implemented:

```
package reader

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/whosonfirst/go-ioutil"
)

type HTTPReader struct {
	Reader
	url        *url.URL
	throttle   <-chan time.Time
	user_agent string
}

func init() {

	ctx := context.Background()

	schemes := []string{
		"http",
		"https",
	}

	for _, s := range schemes {

		err := RegisterReader(ctx, s, NewHTTPReader)

		if err != nil {
			panic(err)
		}
	}
}

func NewHTTPReader(ctx context.Context, uri string) (Reader, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	rate := time.Second / 3
	throttle := time.Tick(rate)

	r := HTTPReader{
		throttle: throttle,
		url:      u,
	}

	q := u.Query()
	ua := q.Get("user-agent")

	if ua != "" {
		r.user_agent = ua
	}

	return &r, nil
}

func (r *HTTPReader) Exists(ctx context.Context, uri string) (bool, error) {

	<-r.throttle

	u, _ := url.Parse(r.url.String())
	u.Path = filepath.Join(u.Path, uri)

	url := u.String()

	req, err := http.NewRequest(http.MethodHead, url, nil)

	if err != nil {
		return false, fmt.Errorf("Failed to create new request, %w", err)
	}

	if r.user_agent != "" {
		req.Header.Set("User-Agent", r.user_agent)
	}

	cl := &http.Client{}

	rsp, err := cl.Do(req)

	if err != nil {
		return false, err
	}

	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return false, nil
	}

	return true, nil
}

func (r *HTTPReader) Read(ctx context.Context, uri string) (io.ReadSeekCloser, error) {

	<-r.throttle

	u, _ := url.Parse(r.url.String())
	u.Path = filepath.Join(u.Path, uri)

	url := u.String()

	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		return nil, fmt.Errorf("Failed to create new request, %w", err)
	}

	if r.user_agent != "" {
		req.Header.Set("User-Agent", r.user_agent)
	}

	cl := &http.Client{}

	rsp, err := cl.Do(req)

	if err != nil {
		return nil, fmt.Errorf("Failed to execute request, %w", err)
	}

	if rsp.StatusCode != 200 {
		return nil, fmt.Errorf("Unexpected status code: %s", rsp.Status)
	}

	fh, err := ioutil.NewReadSeekCloser(rsp.Body)

	if err != nil {
		return nil, fmt.Errorf("Failed to create new ReadSeekCloser, %w", err)
	}

	return fh, nil
}

func (r *HTTPReader) ReaderURI(ctx context.Context, uri string) string {
	return uri
}
```

And then to use it you would do this:

```
package main

import (
	"context"
	"io"
	"os"

	"github.com/whosonfirst/go-reader/v2"
)

func main() {
	ctx := context.Background()
	r, _ := reader.NewReader(ctx, "https://data.whosonfirst.org")
	fh, _ := r.Read(ctx, "101/736/545/101736545.geojson")
	defer fh.Close()
	io.Copy(os.Stdout, fh)
}
```

## Available readers

### "blob"

Read files from any registered [Go Cloud](https://gocloud.dev/howto/blob/) `Blob` source. For example:

```
import (
	"context"

	_ "github.com/whosonfirst/go-reader-blob/v2"
	_ "gocloud.dev/blob/s3blob"	

	"github.com/whosonfirst/go-reader/v2"
)

func main() {
	ctx := context.Background()
	r, _ := reader.NewReader(ctx, "s3://whosonfirst-data?region=us-west-1")
}
```

* https://github.com/whosonfirst/go-reader-blob

### findingaid://

Read files derived from a [Who's On First style findingaid](https://github.com/whosonfirst/go-whosonfirst-findingaid) endpoint.

```
import (
       "context"
       "fmt"

	_ "github.com/whosonfirst/go-reader-findingaid/v2"

	"github.com/whosonfirst/go-reader/v2"
)

func main() {

	cwd, _ := os.Getwd()
	template := fmt.Sprintf("fs://%s/fixtures/{repo}/data", cwd)
	reader_uri := fmt.Sprintf("findingaid://sqlite?dsn=fixtures/sfomuseum-data-maps.db&template=%s", template)

	ctx := context.Background()
	r, _ := reader.NewReader(ctx, reader_uri)
}	

* https://github.com/whosonfirst/go-reader-findingaid

### github://

Read files from a GitHub repository.

```
import (
	"context"

	_ "github.com/whosonfirst/go-reader-github/v2"
	
	"github.com/whosonfirst/go-reader/v2"
)

func main() {
	ctx := context.Background()
	r, _ := reader.NewReader(ctx, "github://{GITHUB_OWNER}/{GITHUB_REPO}")

	// to specify a specific branch you would do this:
	// r, _ := reader.NewReader(ctx, "github://{GITHUB_OWNER}/{GITHUB_REPO}?branch={GITHUB_BRANCH}")
}
```

* https://github.com/whosonfirst/go-reader-github

### githubapi://

Read files from a GitHub repository using the GitHub API.

```
import (
	"context"

	_ "github.com/whosonfirst/go-reader-github/v2"
	
	"github.com/whosonfirst/go-reader/v2"
)

func main() {
	ctx := context.Background()
	r, _ := reader.NewReader(ctx, "githubapi://{GITHUB_OWNER}/{GITHUB_REPO}?access_token={GITHUBAPI_ACCESS_TOKEN}")

	// to specify a specific branch you would do this:
	// r, _ := reader.NewReader(ctx, "githubapi://{GITHUB_OWNER}/{GITHUB_REPO}/?branch={GITHUB_BRANCH}&access_token={GITHUBAPI_ACCESS_TOKEN}")
}
```

* https://github.com/whosonfirst/go-reader-github

### http:// and https://

Read files from an HTTP(S) endpoint.

```
import (
	"context"
	
	"github.com/whosonfirst/go-reader/v2"
)

func main() {
	ctx := context.Background()
	r, _ := reader.NewReader(ctx, "https://{HTTP_HOST_AND_PATH}")
}
```

* https://github.com/whosonfirst/go-reader-http

### file://

Read files from a local filesystem.

```
import (
	"context"
	
	"github.com/whosonfirst/go-reader/v2"
)

func main() {
	ctx := context.Background()
	r, _ := reader.NewReader(ctx, "file://{PATH_TO_DIRECTORY}")
}
```

If you are importing the `go-reader-blob` package and using the GoCloud's [fileblob](https://gocloud.dev/howto/blob/#local) driver then instantiating the `file://` scheme will fail since it will have already been registered. You can work around this by using the `fs://` scheme. For example:

```
r, _ := reader.NewReader(ctx, "fs://{PATH_TO_DIRECTORY}")
```

* https://github.com/whosonfirst/go-reader

### null://

Pretend to read files.

```
import (
	"context"
	
	"github.com/whosonfirst/go-reader/v2"
)

func main() {
	ctx := context.Background()
	r, _ := reader.NewReader(ctx, "null://")
}
```

### repo://

This is a convenience scheme for working with Who's On First data repositories.

It will update a URI by appending a `data` directory to its path and changing its scheme to `fs://` before invoking `reader.NewReader` with the updated URI.

```
import (
	"context"
	
	"github.com/whosonfirst/go-reader/v2"
)

func main() {
	ctx := context.Background()
	r, _ := reader.NewReader(ctx, "repo:///usr/local/data/whosonfirst-data-admin-ca")
}
```

### sql://

Read "files" from a `database/sql` database driver.

```
import (
	"context"

	_ "github.com/mattn/go-sqlite3"
	
	"github.com/whosonfirst/go-reader/v2"
)

func main() {
	ctx := context.Background()
	r, _ := reader.NewReader(ctx, "sql://sqlite3/geojson/id/body?dsn=example.db")
}
```

### stdin://

Read "files" from `STDIN`

```
import (
	"context"
	
	"github.com/whosonfirst/go-reader/v2"
)

func main() {
	ctx := context.Background()
	r, _ := reader.NewReader(ctx, "stdin://")
}
```

And then to use, something like:

```
> cat README.md | ./bin/read -reader-uri stdin:// - | wc -l
     339
```

Note the use of `-` for a URI. This is the convention (when reading from STDIN) but it can be whatever you want it to be.

## See also

* https://github.com/whosonfirst/go-writer

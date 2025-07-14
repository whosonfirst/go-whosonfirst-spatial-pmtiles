// go-whosonfirst-export is a Go package for exporting Who's On First documents in Go.
//
// Example
//
//	import (
//		"context"
//		"os
//
//		"github.com/whosonfirst/go-whosonfirst-export/v3"
//	)
//
//	func main() {
//
//		ctx := context.Background()
//		ex, _ := export.NewExporter(ctx, "whosonfirst://")
//
//		path := "some.geojson"
//		body, _ := os.ReadFile(path)
//
//		has_changed, body, _ = ex.Export(ctx, body)
//
//		if has_changed {
//			os.Stdout.Write(body)
//		}
//	}
package export

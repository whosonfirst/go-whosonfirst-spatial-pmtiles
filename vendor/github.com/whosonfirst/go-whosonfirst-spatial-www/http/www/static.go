package www

import (
	"fmt"
	"io/fs"
	_ "log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/whosonfirst/go-whosonfirst-spatial-www/static"
)

func StaticAssetsHandler() (http.Handler, error) {
	http_fs := http.FS(static.FS)
	return http.FileServer(http_fs), nil
}

func StaticAssetsHandlerWithPrefix(prefix string) (http.Handler, error) {

	fs_handler, err := StaticAssetsHandler()

	if err != nil {
		return nil, err
	}

	fs_handler = http.StripPrefix(prefix, fs_handler)
	return fs_handler, nil
}

func AppendStaticAssetHandlers(mux *http.ServeMux) error {
	return AppendStaticAssetHandlersWithPrefix(mux, "")
}

func AppendStaticAssetHandlersWithPrefix(mux *http.ServeMux, prefix string) error {

	asset_handler, err := StaticAssetsHandlerWithPrefix(prefix)

	if err != nil {
		return nil
	}

	walk_func := func(path string, info fs.DirEntry, err error) error {

		if path == "." {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		if prefix != "" {
			path = appendPrefix(prefix, path)
		}

		if !strings.HasPrefix(path, "/") {
			path = fmt.Sprintf("/%s", path)
		}

		// log.Println("APPEND", path)

		mux.Handle(path, asset_handler)
		return nil
	}

	return fs.WalkDir(static.FS, ".", walk_func)
}

func appendPrefix(prefix string, path string) string {

	prefix = strings.TrimRight(prefix, "/")

	if prefix != "" {
		path = strings.TrimLeft(path, "/")
		path = filepath.Join(prefix, path)
	}

	return path
}

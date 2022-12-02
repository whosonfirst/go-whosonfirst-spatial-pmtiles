package database

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func DSNFromURI(uri string) (string, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return "", fmt.Errorf("Failed to parse URI, %w", err)
	}

	host := u.Host
	path := u.Path
	q := u.RawQuery

	var dsn string

	switch host {
	case "mem":
		dsn = "file::memory:?mode=memory&cache=shared"
	case "vfs":
		// pass, TBD
	default:

		if host == "cwd" {

			cwd, err := os.Getwd()

			if err != nil {
				return "", fmt.Errorf("Failed to derived current working directory, %w", err)
			}

			path = filepath.Join(cwd, path)
		}

		dsn = fmt.Sprintf("file:%s?cache=shared&mode=rwc", path)
	}

	if q != "" {

		if strings.Contains(dsn, "?") {
			dsn = fmt.Sprintf("%s&%s", dsn, q)
		} else {
			dsn = fmt.Sprintf("%s?%s", dsn, q)
		}
	}

	return dsn, nil
}

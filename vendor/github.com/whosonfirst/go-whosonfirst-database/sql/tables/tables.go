package tables

import (
	"bufio"
	"bytes"
	"database/sql"
	"embed"
	"fmt"
	"text/template"

	database_sql "github.com/sfomuseum/go-database/sql"
)

//go:embed *.schema
var fs embed.FS

func LoadSchema(db *sql.DB, table_name string) (string, error) {

	driver := database_sql.Driver(db)

	fname := fmt.Sprintf("%s.%s.schema", table_name, driver)

	data, err := fs.ReadFile(fname)

	if err != nil {
		return "", fmt.Errorf("Failed to read %s, %w", fname, err)
	}

	t, err := template.New(table_name).Parse(string(data))

	if err != nil {
		return "", fmt.Errorf("Failed to parse %s template, %w", fname, err)
	}

	vars := struct {
		Name string
	}{
		Name: table_name,
	}

	var buf bytes.Buffer
	wr := bufio.NewWriter(&buf)

	err = t.Execute(wr, vars)

	if err != nil {
		return "", fmt.Errorf("Failed to process %s template, %w", fname, err)
	}

	wr.Flush()

	return buf.String(), nil
}

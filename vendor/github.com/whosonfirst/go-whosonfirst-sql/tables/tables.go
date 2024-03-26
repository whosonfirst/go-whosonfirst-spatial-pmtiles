package tables

import (
	"bufio"
	"bytes"
	"embed"
	"fmt"
	"text/template"
)

//go:embed *.schema
var fs embed.FS

func LoadSchema(engine string, table_name string) (string, error) {

	fname := fmt.Sprintf("%s.%s.schema", table_name, engine)

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

// Package text provides methods for loading text (.txt) templates with default functions
package text

import (
	"context"
	"fmt"
	"io/fs"
	"text/template"

	"github.com/sfomuseum/go-template/funcs"
)

// LoadTemplates loads text templates matching ".txt" from 't_fs' with default functions assigned.
func LoadTemplates(ctx context.Context, t_fs ...fs.FS) (*template.Template, error) {
	return LoadTemplatesMatching(ctx, "*.txt", t_fs...)
}

// LoadTemplatesMatching loads text templates matching 'pattern' from 't_fs' with default functions assigned.
func LoadTemplatesMatching(ctx context.Context, pattern string, t_fs ...fs.FS) (*template.Template, error) {

	funcs := TemplatesFuncMap()
	t := template.New("text").Funcs(funcs)

	var err error

	for idx, f := range t_fs {

		t, err = t.ParseFS(f, pattern)

		if err != nil {
			return nil, fmt.Errorf("Failed to load templates from FS at offset %d, %w", idx, err)
		}
	}

	return t, nil
}

// TemplatesFuncMap() returns a `template.FuncMap` instance with default functions assigned.
func TemplatesFuncMap() template.FuncMap {

	return template.FuncMap{
		// For example: {{ if (IsAvailable "Account" .) }}
		"IsAvailable": funcs.IsAvailable,
		"Add":         funcs.Add,
	}
}

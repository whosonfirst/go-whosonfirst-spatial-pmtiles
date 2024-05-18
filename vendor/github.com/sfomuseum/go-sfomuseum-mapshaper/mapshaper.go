package mapshaper

import (
	"context"
	"errors"
	"fmt"
	_ "log"
	"os"
	"os/exec"
)

type Mapshaper struct {
	path string
}

func NewMapshaper(ctx context.Context, path string) (*Mapshaper, error) {

	info, err := os.Stat(path)

	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		return nil, errors.New("Invalid path")
	}

	ms := &Mapshaper{
		path: path,
	}

	return ms, nil
}

func (ms *Mapshaper) Call(ctx context.Context, args ...string) ([]byte, error) {

	cmd := exec.CommandContext(ctx, ms.path, args...)
	out, err := cmd.CombinedOutput()

	if err != nil {
		return nil, fmt.Errorf("Failed to call Mapshaper: %w\n%s", err, string(out))
	}

	return out, nil
}

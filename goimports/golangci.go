package goimports

import (
	"bytes"
	"os"
	"path/filepath"

	"github.com/rogpeppe/go-internal/diff"
	"golang.org/x/tools/imports"
)

// Run runs goimports.
// The local prefixes (comma separated) must be defined through the global variable imports.LocalPrefix.
func Run(filename string) ([]byte, error) {
	src, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	res, err := imports.Process(filename, src, nil)
	if err != nil {
		return nil, err
	}

	if bytes.Equal(src, res) {
		return nil, nil
	}

	// formatting has changed
	newName := filepath.ToSlash(filename)
	oldName := newName + ".orig"

	return diff.Diff(oldName, src, newName, res), nil
}

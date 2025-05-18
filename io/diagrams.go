package io

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// ListDiagrams list the saved diagrams.
func ListDiagrams(dir string) ([]string, error) {
	var diagrams []string

	cwd, err := filepath.Abs(filepath.Dir(""))
	if err != nil {
		return nil, fmt.Errorf("could not return the current working directory: %w", err)
	}

	path := cwd + dir
	_, err = os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		if err = os.Mkdir(path, os.ModePerm); err != nil {
			return nil, fmt.Errorf("cannot create directory: %w", err)
		}
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read directory: %w", err)
	}

	for _, file := range files {
		diagrams = append(diagrams, file.Name())
	}

	return diagrams, nil
}

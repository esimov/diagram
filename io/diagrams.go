package io

import (
	"fmt"
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
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		diagrams = append(diagrams, file.Name())
	}

	return diagrams, nil
}

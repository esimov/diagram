package io

import (
	"io/ioutil"
	"log"
	"path/filepath"
)

// List the saved diagrams
func ListDiagrams(dir string) ([]string, error) {
	var diagrams []string

	cwd, err := filepath.Abs(filepath.Dir(""))
	if err != nil {
		log.Fatal(err)
	}
	path := cwd + dir
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		diagrams = append(diagrams, file.Name())
	}

	return diagrams, nil
}

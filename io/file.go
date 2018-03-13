package io

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

// ReadFile read the input file.
func ReadFile(input string) []byte {
	b, err := ioutil.ReadFile(input)
	if err != nil {
		log.Fatalf("Cannot read input file: %v", err.Error())
		return nil
	}
	return b
}

// SaveFile saves the diagram into the output directory.
func SaveFile(fileName string, output string, data string) (*os.File, error) {
	var file *os.File

	cwd, err := filepath.Abs(filepath.Dir(""))
	if err != nil {
		log.Fatal(err)
	}
	filePath := cwd + output
	// Check if output directory does not exits. In this case create it.
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		os.Mkdir(filePath, os.ModePerm)
	}
	// Create the output file.
	file, err = os.Create(filepath.Join(filePath, fileName))
	if isError(err) {
		return nil, err
	}

	file, err = os.OpenFile(filepath.Join(filePath, fileName), os.O_RDWR, 0644)
	if isError(err) {
		return nil, err
	}

	defer file.Close()

	_, err = file.WriteString(data)
	if isError(err) {
		return nil, err
	}
	return file, nil
}

// DeleteDiagram deletes a diagram.
func DeleteDiagram(fileName string) error {
	err := os.Remove(fileName)
	if isError(err) {
		return err
	}
	return nil
}

// isError ia a generic function to check for errors.
func isError(err error) bool {
	if err != nil {
		fmt.Errorf("Could not save file: %v", err.Error())
	}
	return (err != nil)
}

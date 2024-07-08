package io

import (
	"fmt"
	"os"
	"path/filepath"
)

// ReadFile read the input file.
func ReadFile(input string) ([]byte, error) {
	b, err := os.ReadFile(input)
	if err != nil {
		return nil, fmt.Errorf("error reading the file: %w", err)
	}
	return b, nil
}

// SaveFile saves the diagram into the output directory.
func SaveFile(fileName string, output string, data string) (*os.File, error) {
	var file *os.File

	cwd, _ := filepath.Abs(filepath.Dir(""))

	filePath := cwd + output
	// Check if the output directory does not exits. In this case create it.
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		os.Mkdir(filePath, os.ModePerm)
	}

	// Create the output file.
	file, err := os.Create(filepath.Join(filePath, fileName))
	if err != nil {
		return nil, fmt.Errorf("error creating the file: %w", err)
	}
	defer file.Close()

	file, err = os.OpenFile(filepath.Join(filePath, fileName), os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("error opening the file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(data)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// DeleteDiagram deletes a diagram.
func DeleteDiagram(fileName string) error {
	err := os.Remove(fileName)
	if err != nil {
		return fmt.Errorf("failed deleting the diagram: %w", err)
	}

	return nil
}

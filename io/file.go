package io

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ReadFile read the input file, and if it's the case
// replaces the CRLF line endings with LF.
func ReadFile(input string) (string, error) {
	f, err := os.Open(input)
	if err != nil {
		return "", fmt.Errorf("unable to open file: %w", err)
	}
	defer f.Close()

	fileInfo, err := f.Stat()
	if err != nil {
		return "", err
	}

	if isExecutable(fileInfo.Mode().Perm()) {
		return "", nil
	}

	buff, err := os.ReadFile(input)
	if err != nil {
		return "", fmt.Errorf("error reading the file: %w", err)
	}

	sb := new(strings.Builder)

	scanner := bufio.NewScanner(bytes.NewBuffer(buff))
	for scanner.Scan() {
		// Scanner already handles different line endings, just write with LF.
		sb.WriteString(scanner.Text() + "\n")
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}

	return sb.String(), nil
}

// SaveFile saves the diagram into the output directory.
func SaveFile(fileName string, output string, data string) (*os.File, error) {
	var f *os.File

	cwd, _ := filepath.Abs(filepath.Dir(""))

	filePath := cwd + output
	// Check if the output directory does not exits. In this case create it.
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if err := os.Mkdir(filePath, os.ModePerm); err != nil {
			return nil, fmt.Errorf("cannot create directory: %w", err)
		}
	}

	// Create the output file.
	f, err := os.Create(filepath.Join(filePath, fileName))
	if err != nil {
		return nil, fmt.Errorf("error creating the file: %w", err)
	}
	defer f.Close()

	f, err = os.OpenFile(filepath.Join(filePath, fileName), os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("error opening the file: %w", err)
	}
	defer f.Close()

	_, err = f.WriteString(data)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// DeleteFile deletes a diagram.
func DeleteFile(fileName string) error {
	err := os.Remove(fileName)
	if err != nil {
		return fmt.Errorf("failed deleting the diagram: %w", err)
	}

	return nil
}

func isExecutable(mode os.FileMode) bool {
	return mode&0111 > 0
}

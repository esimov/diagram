package io

import (
	"io/ioutil"
	"fmt"
	"os"
	"path/filepath"
	"log"
)

// Read file
func ReadFile(input string) []byte {
	b, err := ioutil.ReadFile(input)
	if isError(err) {
		fmt.Errorf("Cannot read input file: ", err.Error())
		return nil
	}
	return b
}

// Create saved diagrams output directory in case it not exists, and save the diagrams into this directory
func SaveFile(fileName string, dir string, data string) (*os.File, error) {
	var file *os.File

	cwd, err := filepath.Abs(filepath.Dir(""))
	if err != nil {
		log.Fatal(err)
	}
	filePath := cwd + dir
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

// Check for errors
func isError(err error) bool {
	if err != nil {
		fmt.Errorf("Could not save file: ", err.Error())
	}
	return (err != nil)
}
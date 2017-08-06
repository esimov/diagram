package io

import (
	"io/ioutil"
	"fmt"
	"os"
	"path/filepath"
	"log"
)

func ReadFile(input string) []byte {
	b, err := ioutil.ReadFile(input)
	if isError(err) {
		fmt.Errorf("Cannot read input file: ", err.Error())
		return nil
	}
	return b
}

func SaveFile(fileName string, dir string, data string) (*os.File, error) {
	var file *os.File
	cwd, err := filepath.Abs(filepath.Dir(""))
	if err != nil {
		log.Fatal(err)
	}
	file, err = os.Create(filepath.Join(cwd, fileName))
	if isError(err) {
		return nil, err
	}

	file, err = os.OpenFile(filepath.Join(cwd, fileName), os.O_RDWR, 0644)
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

func isError(err error) bool {
	if err != nil {
		fmt.Errorf("Could not save file: ", err.Error())
	}
	return (err != nil)
}
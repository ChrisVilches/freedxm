package fileutil

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

func readFileEntireContent(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func ReadTomlFile[T any](filename string) (*T, error) {
	tomlData, err := readFileEntireContent(filename)
	if err != nil {
		return nil, err
	}

	var result T
	_, err = toml.Decode(tomlData, &result)

	if err != nil {
		return nil, fmt.Errorf("could not read config file: %v", err)
	}

	return &result, nil
}

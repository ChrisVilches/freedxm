package fileutil

import (
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
		return nil, err
	}

	return &result, nil
}

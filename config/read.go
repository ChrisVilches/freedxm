package config

import (
	"os"
	"path"

	"github.com/ChrisVilches/freedxm/fileutil"
	"github.com/ChrisVilches/freedxm/model"
)

func getDefaultConfigFilePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return path.Join(home, ".config", "freedxm.toml"), nil
}

func getConfigFilePath() (string, error) {
	str, present := os.LookupEnv("CONFIG_FILEPATH")

	if !present {
		return getDefaultConfigFilePath()
	}

	return str, nil
}

func GetBlockListByName(name string) (*model.BlockList, error) {
	filepath, err := getConfigFilePath()
	if err != nil {
		return nil, err
	}
	config, err := fileutil.ReadTomlFile[Config](filepath)

	if err != nil {
		return nil, err
	}

	for _, blockList := range config.Blocklists {
		if blockList.Name == name {
			return &blockList, nil
		}
	}

	return nil, nil
}

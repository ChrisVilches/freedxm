package config

import (
	"fmt"

	"github.com/ChrisVilches/freedxm/fileutil"
	"github.com/ChrisVilches/freedxm/model"
)

// TODO: Should be configurable.
var configFilePath = "./conf/block-lists.toml"

type BlockListNotFoundError struct {
	wrongName      string
	AvailableNames []string
}

func (e *BlockListNotFoundError) Error() string {
	return fmt.Sprintf("no blocklist found with name '%s'", e.wrongName)
}

func GetBlockListByName(name string) (*model.BlockList, error) {
	config, err := fileutil.ReadTomlFile[Config](configFilePath)

	if err != nil {
		return nil, err
	}

	for _, blockList := range config.Blocklists {
		if blockList.Name == name {
			return &blockList, nil
		}
	}

	return nil, &BlockListNotFoundError{
		wrongName:      name,
		AvailableNames: config.GetAllNames(),
	}
}

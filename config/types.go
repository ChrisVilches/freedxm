package config

import (
	"github.com/ChrisVilches/freedxm/model"
)

type Config struct {
	Blocklists   []model.BlockList  `toml:"blocklist"`
	Options      model.Options      `toml:"options"`
	Notification model.Notification `toml:"notification"`
}

func (c *Config) GetAllNames() []string {
	res := []string{}
	for _, blockList := range c.Blocklists {
		res = append(res, blockList.Name)
	}
	return res
}

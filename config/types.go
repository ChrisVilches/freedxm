package config

type blockList struct {
	Name     string   `toml:"name"`
	Domains  []string `toml:"domains"`
	Programs []string `toml:"programs"`
}

type Config struct {
	Blocklists []blockList `toml:"blocklist"`
}

func (c *Config) GetAllNames() []string {
	res := []string{}
	for _, blockList := range c.Blocklists {
		res = append(res, blockList.Name)
	}
	return res
}

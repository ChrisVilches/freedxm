package config

type blockList struct {
	Name     string   `toml:"name"`
	Domains  []string `toml:"domains"`
	Programs []string `toml:"programs"`
}

type Config struct {
	Blocklists []blockList `toml:"blocklist"`
}

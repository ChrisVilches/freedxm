package model

type BlockList struct {
	Name      string   `json:"name" toml:"name"`
	Domains   []string `json:"domains" toml:"domains"`
	Processes []string `json:"processes" toml:"processes"`
}

type Session struct {
	TimeSeconds int         `json:"timeSeconds"`
	BlockLists  []BlockList `json:"blockLists"`
}

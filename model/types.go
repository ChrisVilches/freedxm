package model

type BlockList struct {
	Name      string   `json:"name" toml:"name"`
	Domains   []string `json:"domains" toml:"domains"`
	Processes []string `json:"processes" toml:"processes"`
}

type Options struct {
	LogDateTime bool   `toml:"log-date-time"`
	Notifier    string `toml:"notifier"`
}

type Notification struct {
	Normal  []string `toml:"normal"`
	Warning []string `toml:"warning"`
}

type Session struct {
	TimeSeconds int         `json:"timeSeconds"`
	BlockLists  []BlockList `json:"blockLists"`
}

package server

import "code.google.com/p/gcfg"

type Config struct {
	General struct {
		DataPath string `gcfg:"data-path"`
	}
	API struct {
		HttpPort              int    `gcfg:"http-port"`
		DevelopmentModeAssets string `gcfg:"development-mode-assets"`
	}
	Network struct {
		Port     int
		Hostname string
		Join     []string
	}
}

func DefaultConfig() (cfg *Config) {
	cfg = &Config{}
	cfg.General.DataPath = "gitchain.db"
	cfg.API.HttpPort = 3000
	cfg.Network.Port = 31000
	return
}

func ReadConfig(filename string, cfg *Config) (err error) {
	err = gcfg.ReadFileInto(cfg, filename)
	return
}

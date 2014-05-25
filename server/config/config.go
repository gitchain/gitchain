package config

import "code.google.com/p/gcfg"

type T struct {
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

func Default() (cfg *T) {
	cfg = &T{}
	cfg.General.DataPath = "gitchain.db"
	cfg.API.HttpPort = 3000
	cfg.Network.Port = 31000
	return
}

func ReadFile(filename string, cfg *T) (err error) {
	err = gcfg.ReadFileInto(cfg, filename)
	return
}

package server

import (
	"os"
	"path/filepath"

	"github.com/gitchain/gitchain/db"
	"github.com/inconshreveable/log15"
)

type T struct {
	Config *Config
	LiveUI string
	DB     *db.T
	Log    log15.Logger
}

func (srv *T) Init() error {
	err := os.MkdirAll(srv.Config.General.DataPath, os.ModeDir|0700)
	if err != nil {
		return err
	}
	database, err := db.NewDB(filepath.Join(srv.Config.General.DataPath, "db"))
	if err != nil {
		return err
	}
	srv.DB = database
	srv.Log = log15.New()
	return nil
}

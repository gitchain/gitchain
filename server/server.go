package server

import (
	"os"
	"path/filepath"

	"github.com/gitchain/gitchain/db"
)

type T struct {
	HttpPort int
	Path     string
	LiveUI   string
	DB       *db.T
}

func New(httpPort int, dbPath string, liveUI string) (*T, error) {
	srv := &T{HttpPort: httpPort, Path: dbPath, LiveUI: liveUI}
	err := os.MkdirAll(dbPath, os.ModeDir|0700)
	if err != nil {
		return nil, err
	}
	database, err := db.NewDB(filepath.Join(dbPath, "db"))
	if err != nil {
		return nil, err
	}
	srv.DB = database
	return srv, nil
}

package server

import (
	"os"
	"path/filepath"

	"github.com/gitchain/gitchain/db"
)

type T struct {
	HttpPort int
	Path     string
	DB       *db.T
}

func New(httpPort int, dbPath string) (*T, error) {
	srv := &T{HttpPort: httpPort, Path: dbPath}
	var err error
	os.MkdirAll(dbPath, os.ModeDir|0600)
	database, err := db.NewDB(filepath.Join(dbPath, "db"))
	if err != nil {
		return nil, err
	}
	srv.DB = database
	return srv, nil
}

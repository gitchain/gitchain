package server

import "github.com/gitchain/gitchain/db"

type T struct {
	HttpPort int
	DB       *db.T
}

func New(httpPort int, dbPath string) (*T, error) {
	srv := &T{HttpPort: httpPort}
	var err error
	database, err := db.NewDB(dbPath)
	if err != nil {
		return nil, err
	}
	srv.DB = database
	return srv, nil
}

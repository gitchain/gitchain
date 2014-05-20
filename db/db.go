package db

import (
	"github.com/boltdb/bolt"
)

type T struct {
	Path string
	DB   *bolt.DB
}

func NewDB(path string) (*T, error) {
	db, err := bolt.Open(path, 0666)
	return &T{Path: path, DB: db}, err
}

func writable(e *error, db *T, f func(*bolt.Tx) bool) error {
	dbtx, err := db.DB.Begin(true)
	success := false
	if err != nil {
		return err
	}
	defer func() {
		if success {
			dbtx.Commit()
		} else {
			dbtx.Rollback()
		}
	}()
	success = f(dbtx)
	if e != nil && *e == nil {
		*e = err
	}
	return err
}

func readable(e *error, db *T, f func(*bolt.Tx)) error {
	dbtx, err := db.DB.Begin(false)
	if err != nil {
		return err
	}
	defer dbtx.Rollback()
	f(dbtx)
	if e != nil && *e == nil {
		*e = err
	}
	return err
}

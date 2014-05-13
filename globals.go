package main

import (
	"flag"
	"log"

	"github.com/gitchain/gitchain/db"
)

var DB *db.T

func init() {
	var dbPath string
	flag.StringVar(&dbPath, "db", "gitchain.db", "path to database, defaults to gitchain.db")
	flag.Parse()
	var err error
	DB, err = db.NewDB(dbPath)
	if err != nil {
		log.Panicf("Can't open database %v because of %v", dbPath, err)
	}
}

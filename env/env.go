package env

import (
	"flag"
	"log"

	"../db"
)

var DB *db.T
var Port int

func init() {
	var dbPath string

	flag.StringVar(&dbPath, "db", "gitchain.db", "path to database, defaults to gitchain.db")
	flag.IntVar(&Port, "port", 3000, "port to connect to or serve on")
	flag.Parse()

	// Initialize database
	if len(flag.Args()) == 0 || flag.Arg(0) == "Serve" {
		var err error
		DB, err = db.NewDB(dbPath)
		if err != nil {
			log.Panicf("Can't open database %v because of %v", dbPath, err)
		}
	}
}

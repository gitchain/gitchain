package api

import (
	"encoding/json"
	"log"
	"net/http"

	"../block"
	"../env"
	"../server"
)

type Info struct {
	Mining    server.MiningStatus
	LastBlock *block.Block
}

func info(resp http.ResponseWriter) string {
	lastBlock, err := env.DB.GetLastBlock()
	if err != nil {
		log.Printf("Error while serving /info: %v", err)
	}
	json, err := json.Marshal(Info{Mining: server.GetMiningStatus(), LastBlock: lastBlock})
	if err != nil {
		log.Printf("Error while serving /info: %v", err)
	}
	resp.Header().Add("Content-Type", "application/json")
	return string(json)
}

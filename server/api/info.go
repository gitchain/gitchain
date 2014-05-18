package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gitchain/gitchain/block"
	"github.com/gitchain/gitchain/server"
)

type Info struct {
	Mining    server.MiningStatus
	LastBlock *block.Block
}

func info(resp http.ResponseWriter, req *http.Request) {
	lastBlock, err := srv.DB.GetLastBlock()
	if err != nil {
		log.Printf("Error while serving /info: %v", err)
	}
	json, err := json.Marshal(Info{Mining: server.GetMiningStatus(), LastBlock: lastBlock})
	if err != nil {
		log.Printf("Error while serving /info: %v", err)
	}
	resp.Header().Add("Content-Type", "application/json")
	resp.Write(json)
}

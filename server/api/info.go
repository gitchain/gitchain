package api

import (
	"encoding/json"
	"net/http"

	"github.com/gitchain/gitchain/block"
	"github.com/gitchain/gitchain/server"
	"github.com/inconshreveable/log15"
)

type Info struct {
	Mining    server.MiningStatus
	LastBlock *block.Block
}

func infoHandler(srv *server.T, log log15.Logger) func(http.ResponseWriter, *http.Request) {
	return func(resp http.ResponseWriter, req *http.Request) {
		log := log.New("http")
		lastBlock, err := srv.DB.GetLastBlock()
		if err != nil {
			log.Error("error serving /info", "err", err)
		}
		json, err := json.Marshal(Info{Mining: server.GetMiningStatus(), LastBlock: lastBlock})
		if err != nil {
			log.Error("error serving /info", "err", err)
		}
		resp.Header().Add("Content-Type", "application/json")
		resp.Write(json)
	}

}

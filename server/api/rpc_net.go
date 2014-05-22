package api

import (
	"net/http"

	"github.com/gitchain/gitchain/router"
	"github.com/gitchain/gitchain/server"
	"github.com/inconshreveable/log15"
)

type NetService struct {
	srv *server.T
	log log15.Logger
}

type JoinArgs struct {
	Host string
}

type JoinReply struct {
}

func (*NetService) Join(r *http.Request, args *JoinArgs, reply *JoinReply) error {
	router.Send("/dht/join", make(chan string), args.Host)
	return nil
}

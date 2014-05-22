package api

import (
	"net/http"

	"github.com/gitchain/gitchain/router"
)

type NetService struct{}

type JoinArgs struct {
	Host string
}

type JoinReply struct {
}

func (*NetService) Join(r *http.Request, args *JoinArgs, reply *JoinReply) error {
	router.Send("/dht/join", make(chan string), args.Host)
	return nil
}
